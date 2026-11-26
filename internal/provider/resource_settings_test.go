package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var settingsIntAttrs = []string{"s3_health_up_threshold", "s3_health_down_threshold", "s3_health_interval_ms", "s3_health_timeout_ms"}

func settingsSchema(ctx context.Context) (schema.Schema, tftypes.Object) {
	var resp resource.SchemaResponse
	NewSettingsResource().Schema(ctx, resource.SchemaRequest{}, &resp)
	return resp.Schema, resp.Schema.Type().TerraformType(ctx).(tftypes.Object)
}

func settingsConfig(ctx context.Context, set map[string]any) tfsdk.Config {
	s, objType := settingsSchema(ctx)
	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, at := range objType.AttributeTypes {
		attrs[name] = tftypes.NewValue(at, nil)
	}
	for name, v := range set {
		attrs[name] = tftypes.NewValue(tftypes.Number, v)
	}
	return tfsdk.Config{Schema: s, Raw: tftypes.NewValue(objType, attrs)}
}

func emptySettingsState(ctx context.Context) tfsdk.State {
	s, objType := settingsSchema(ctx)
	return tfsdk.State{Schema: s, Raw: tftypes.NewValue(objType, nil)}
}

type recordedRequest struct {
	method string
	path   string
	auth   string
	body   map[string]any
}

func settingsServer(t *testing.T, status int, respBody string) (*apiClient, *[]recordedRequest) {
	t.Helper()
	var reqs []recordedRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := recordedRequest{method: r.Method, path: r.URL.Path, auth: r.Header.Get("Authorization")}
		if b, _ := io.ReadAll(r.Body); len(b) > 0 {
			if err := json.Unmarshal(b, &rec.body); err != nil {
				t.Errorf("request body is not JSON: %s", b)
			}
		}
		reqs = append(reqs, rec)
		w.WriteHeader(status)
		fmt.Fprint(w, respBody)
	}))
	t.Cleanup(srv.Close)
	c := newAPIClient(srv.URL)
	c.token = "tok"
	return c, &reqs
}

const settingsResponseJSON = `{"s3_health_up_threshold":5,"s3_health_down_threshold":6,"s3_health_interval_ms":3000,"s3_health_timeout_ms":1500,` +
	`"overrides":{"s3_health_up_threshold":5,"s3_health_down_threshold":null,"s3_health_interval_ms":null,"s3_health_timeout_ms":null},"valid":true}`

func assertSettingsState(t *testing.T, ctx context.Context, state tfsdk.State) {
	t.Helper()
	var m settingsResourceModel
	if diags := state.Get(ctx, &m); diags.HasError() {
		t.Fatalf("state get: %v", diags)
	}
	if m.ID.ValueString() != "settings" {
		t.Errorf("id: expected settings, got %q", m.ID.ValueString())
	}
	if m.S3HealthUpThreshold.ValueInt64() != 5 {
		t.Errorf("s3_health_up_threshold: expected override 5, got %v", m.S3HealthUpThreshold)
	}
	for name, v := range map[string]types.Int64{
		"s3_health_down_threshold": m.S3HealthDownThreshold,
		"s3_health_interval_ms":    m.S3HealthIntervalMs,
		"s3_health_timeout_ms":     m.S3HealthTimeoutMs,
	} {
		if !v.IsNull() {
			t.Errorf("%s: no override, so state must be null, got %v", name, v)
		}
	}
	want := map[string]int64{"s3_health_up_threshold": 5, "s3_health_down_threshold": 6, "s3_health_interval_ms": 3000, "s3_health_timeout_ms": 1500}
	eff := m.Effective.Attributes()
	for k, v := range want {
		if got, ok := eff[k].(types.Int64); !ok || got.ValueInt64() != v {
			t.Errorf("effective.%s: expected %d, got %v", k, v, eff[k])
		}
	}
}

func TestSettingsResource_Metadata(t *testing.T) {
	var resp resource.MetadataResponse
	NewSettingsResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "artifactapi"}, &resp)
	if resp.TypeName != "artifactapi_settings" {
		t.Errorf("expected artifactapi_settings, got %s", resp.TypeName)
	}
}

func TestSettingsResource_Schema(t *testing.T) {
	s, _ := settingsSchema(context.Background())
	if len(s.Attributes) != len(settingsIntAttrs)+2 {
		t.Errorf("expected %d attributes, got %d", len(settingsIntAttrs)+2, len(s.Attributes))
	}
	eff, ok := s.Attributes["effective"].(schema.SingleNestedAttribute)
	if !ok || !eff.IsComputed() || eff.IsOptional() || len(eff.Attributes) != len(settingsIntAttrs) {
		t.Error("effective must be a Computed-only nested attribute mirroring every setting")
	}
	id, ok := s.Attributes["id"].(schema.StringAttribute)
	if !ok || !id.IsComputed() || id.IsOptional() || id.IsRequired() {
		t.Error("id must be a Computed-only string")
	}
	for _, name := range settingsIntAttrs {
		a, ok := s.Attributes[name].(schema.Int64Attribute)
		if !ok {
			t.Errorf("%s: missing or not an int64 attribute", name)
			continue
		}
		if !a.IsOptional() || a.IsComputed() {
			t.Errorf("%s must be Optional and not Computed, so removing it from config is a diff", name)
		}
		if a.Default != nil {
			t.Errorf("%s must not set a default; the server owns it", name)
		}
	}
}

func TestSettingsResource_ValidateConfig(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name    string
		set     map[string]any
		wantErr bool
	}{
		{"all unset", nil, false},
		{"all valid", map[string]any{"s3_health_up_threshold": 1, "s3_health_down_threshold": 2, "s3_health_interval_ms": 3000, "s3_health_timeout_ms": 2999}, false},
		{"zero threshold", map[string]any{"s3_health_up_threshold": 0}, true},
		{"negative down threshold", map[string]any{"s3_health_down_threshold": -1}, true},
		{"zero interval", map[string]any{"s3_health_interval_ms": 0}, true},
		{"zero timeout", map[string]any{"s3_health_timeout_ms": 0}, true},
		{"thresholds at bounds", map[string]any{"s3_health_up_threshold": 100, "s3_health_down_threshold": 1}, false},
		{"up threshold above 100", map[string]any{"s3_health_up_threshold": 101}, true},
		{"down threshold above 100", map[string]any{"s3_health_down_threshold": 101}, true},
		{"interval at bounds", map[string]any{"s3_health_interval_ms": 500, "s3_health_timeout_ms": 100}, false},
		{"interval at max", map[string]any{"s3_health_interval_ms": 3600000}, false},
		{"interval below 500", map[string]any{"s3_health_interval_ms": 499}, true},
		{"interval above max", map[string]any{"s3_health_interval_ms": 3600001}, true},
		{"timeout below 100", map[string]any{"s3_health_timeout_ms": 99}, true},
		{"timeout equals interval", map[string]any{"s3_health_interval_ms": 1000, "s3_health_timeout_ms": 1000}, true},
		{"timeout above interval", map[string]any{"s3_health_interval_ms": 1000, "s3_health_timeout_ms": 2000}, true},
		{"timeout alone is left to the server", map[string]any{"s3_health_timeout_ms": 5000}, false},
		{"unknown interval skips comparison", map[string]any{"s3_health_interval_ms": tftypes.UnknownValue, "s3_health_timeout_ms": 5000}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var resp resource.ValidateConfigResponse
			NewSettingsResource().(*settingsResource).ValidateConfig(ctx, resource.ValidateConfigRequest{Config: settingsConfig(ctx, tc.set)}, &resp)
			if resp.Diagnostics.HasError() != tc.wantErr {
				t.Errorf("wantErr=%v, got %v", tc.wantErr, resp.Diagnostics.Errors())
			}
		})
	}
}

func TestSettingsResource_ValidateConfig_ReportsEachInvalidAttribute(t *testing.T) {
	ctx := context.Background()
	var resp resource.ValidateConfigResponse
	set := map[string]any{"s3_health_up_threshold": 0, "s3_health_down_threshold": 0}
	(&settingsResource{}).ValidateConfig(ctx, resource.ValidateConfigRequest{Config: settingsConfig(ctx, set)}, &resp)
	if n := resp.Diagnostics.ErrorsCount(); n != 2 {
		t.Errorf("expected 2 errors, got %d: %v", n, resp.Diagnostics.Errors())
	}
}

func TestSettingsResource_Create_PutsOnlyConfiguredFields(t *testing.T) {
	ctx := context.Background()
	client, reqs := settingsServer(t, http.StatusOK, settingsResponseJSON)
	r := &settingsResource{client: client}

	resp := resource.CreateResponse{State: emptySettingsState(ctx)}
	r.Create(ctx, resource.CreateRequest{Config: settingsConfig(ctx, map[string]any{"s3_health_up_threshold": 5})}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("create: %v", resp.Diagnostics)
	}

	if len(*reqs) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*reqs))
	}
	got := (*reqs)[0]
	if got.method != http.MethodPut || got.path != "/api/v2/settings" {
		t.Errorf("expected PUT /api/v2/settings, got %s %s", got.method, got.path)
	}
	if got.auth != "Bearer tok" {
		t.Errorf("expected bearer auth, got %q", got.auth)
	}
	if got.body["s3_health_up_threshold"] != float64(5) {
		t.Errorf("s3_health_up_threshold: expected 5, got %v", got.body["s3_health_up_threshold"])
	}
	for _, name := range settingsIntAttrs[1:] {
		if v, present := got.body[name]; present && v != nil {
			t.Errorf("%s: unset field must be sent null so it reverts to the env default, got %v", name, v)
		}
	}
	assertSettingsState(t, ctx, resp.State)
}

func TestSettingsResource_Update_Puts(t *testing.T) {
	ctx := context.Background()
	client, reqs := settingsServer(t, http.StatusOK, settingsResponseJSON)
	r := &settingsResource{client: client}

	resp := resource.UpdateResponse{State: emptySettingsState(ctx)}
	r.Update(ctx, resource.UpdateRequest{Config: settingsConfig(ctx, map[string]any{"s3_health_interval_ms": 3000, "s3_health_timeout_ms": 1500})}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("update: %v", resp.Diagnostics)
	}
	got := (*reqs)[0]
	if got.method != http.MethodPut || got.body["s3_health_interval_ms"] != float64(3000) || got.body["s3_health_timeout_ms"] != float64(1500) {
		t.Errorf("unexpected request: %+v", got)
	}
	assertSettingsState(t, ctx, resp.State)
}

func TestSettingsResource_Create_SurfacesServerRejection(t *testing.T) {
	ctx := context.Background()
	client, _ := settingsServer(t, http.StatusBadRequest, "s3_health_timeout_ms (5000) must be shorter than s3_health_interval_ms (3000)")
	r := &settingsResource{client: client}

	resp := resource.CreateResponse{State: emptySettingsState(ctx)}
	r.Create(ctx, resource.CreateRequest{Config: settingsConfig(ctx, map[string]any{"s3_health_timeout_ms": 5000})}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error for a 400 response")
	}
	if !resp.State.Raw.IsNull() {
		t.Error("state must stay empty after a rejected create")
	}
}

func TestSettingsResource_Read(t *testing.T) {
	ctx := context.Background()
	client, reqs := settingsServer(t, http.StatusOK, settingsResponseJSON)
	r := &settingsResource{client: client}

	resp := resource.ReadResponse{State: emptySettingsState(ctx)}
	r.Read(ctx, resource.ReadRequest{State: emptySettingsState(ctx)}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read: %v", resp.Diagnostics)
	}
	if got := (*reqs)[0]; got.method != http.MethodGet || got.path != "/api/v2/settings" {
		t.Errorf("expected GET /api/v2/settings, got %s %s", got.method, got.path)
	}
	assertSettingsState(t, ctx, resp.State)
}

func TestSettingsResource_Read_ErrorKeepsResource(t *testing.T) {
	ctx := context.Background()
	client, _ := settingsServer(t, http.StatusInternalServerError, "boom")
	r := &settingsResource{client: client}

	resp := resource.ReadResponse{State: emptySettingsState(ctx)}
	r.Read(ctx, resource.ReadRequest{State: emptySettingsState(ctx)}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error for a 500 response")
	}
}

func TestSettingsResource_Delete(t *testing.T) {
	ctx := context.Background()
	client, reqs := settingsServer(t, http.StatusNoContent, "")
	r := &settingsResource{client: client}

	var resp resource.DeleteResponse
	r.Delete(ctx, resource.DeleteRequest{State: emptySettingsState(ctx)}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("delete: %v", resp.Diagnostics)
	}
	if got := (*reqs)[0]; got.method != http.MethodDelete || got.path != "/api/v2/settings" {
		t.Errorf("expected DELETE /api/v2/settings, got %s %s", got.method, got.path)
	}
}

func TestSettingsResource_ImportAnyID(t *testing.T) {
	ctx := context.Background()
	for _, id := range []string{"settings", "anything"} {
		resp := resource.ImportStateResponse{State: emptySettingsState(ctx)}
		(&settingsResource{}).ImportState(ctx, resource.ImportStateRequest{ID: id}, &resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("import %q: %v", id, resp.Diagnostics)
		}
		var m settingsResourceModel
		resp.State.Get(ctx, &m)
		if m.ID.ValueString() != "settings" {
			t.Errorf("import %q: expected id settings, got %q", id, m.ID.ValueString())
		}
	}
}

func TestSettingsModelToAPI_UnknownIsNull(t *testing.T) {
	ctx := context.Background()
	var m settingsResourceModel
	settingsConfig(ctx, map[string]any{"s3_health_interval_ms": tftypes.UnknownValue, "s3_health_up_threshold": 2}).Get(ctx, &m)
	api := settingsModelToAPI(m)
	if api.S3HealthIntervalMs != nil {
		t.Errorf("unknown interval must be sent null, got %d", *api.S3HealthIntervalMs)
	}
	if api.S3HealthUpThreshold == nil || *api.S3HealthUpThreshold != 2 {
		t.Errorf("up threshold: expected 2, got %v", api.S3HealthUpThreshold)
	}
}

func TestSettingsResource_PutBodyHasOnlyKnownKeys(t *testing.T) {
	ctx := context.Background()
	client, reqs := settingsServer(t, http.StatusOK, settingsResponseJSON)
	resp := resource.CreateResponse{State: emptySettingsState(ctx)}
	(&settingsResource{client: client}).Create(ctx, resource.CreateRequest{Config: settingsConfig(ctx, nil)}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("create: %v", resp.Diagnostics)
	}
	known := map[string]bool{}
	for _, n := range settingsIntAttrs {
		known[n] = true
	}
	for k := range (*reqs)[0].body {
		if !known[k] {
			t.Errorf("PUT body carries unknown key %q, which the server rejects", k)
		}
	}
}

func TestSettingsResource_Read_ToleratesExtraFields(t *testing.T) {
	ctx := context.Background()
	body := strings.Replace(settingsResponseJSON, `"valid":true`, `"valid":false,"error":"stale","extra":1`, 1)
	client, _ := settingsServer(t, http.StatusOK, body)
	resp := resource.ReadResponse{State: emptySettingsState(ctx)}
	(&settingsResource{client: client}).Read(ctx, resource.ReadRequest{State: emptySettingsState(ctx)}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read must ignore unknown response fields: %v", resp.Diagnostics)
	}
	assertSettingsState(t, ctx, resp.State)
}
