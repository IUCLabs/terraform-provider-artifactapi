package provider

import (
	"context"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// fakeSettingsServer mimics /api/v2/settings: PUT replaces every override, GET reports overrides over env defaults.
type fakeSettingsServer struct {
	mu        sync.Mutex
	overrides map[string]*int64
	puts      []map[string]any
}

var fakeSettingsDefaults = map[string]int64{"s3_health_up_threshold": 3, "s3_health_down_threshold": 6, "s3_health_interval_ms": 3000, "s3_health_timeout_ms": 1500}

func (f *fakeSettingsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	switch r.Method {
	case http.MethodPut:
		b, _ := io.ReadAll(r.Body)
		var body map[string]any
		var parsed map[string]*int64
		if json.Unmarshal(b, &body) != nil || json.Unmarshal(b, &parsed) != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		f.puts = append(f.puts, body)
		f.overrides = parsed
	case http.MethodDelete:
		f.overrides = nil
		w.WriteHeader(http.StatusNoContent)
		return
	}
	resp := map[string]any{"valid": true}
	overrides := map[string]*int64{}
	for name, def := range fakeSettingsDefaults {
		overrides[name] = f.overrides[name]
		resp[name] = def
		if v := f.overrides[name]; v != nil {
			resp[name] = *v
		}
	}
	resp["overrides"] = overrides
	_ = json.NewEncoder(w).Encode(resp)
}

func (f *fakeSettingsServer) setOverride(name string, v int64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.overrides == nil {
		f.overrides = map[string]*int64{}
	}
	f.overrides[name] = &v
}

type settingsHarness struct {
	t       *testing.T
	ctx     context.Context
	server  tfprotov6.ProviderServer
	objType tftypes.Object
	fake    *fakeSettingsServer
}

func newSettingsHarness(t *testing.T) *settingsHarness {
	t.Helper()
	ctx := context.Background()
	fake := &fakeSettingsServer{}
	srv := httptest.NewServer(fake)
	t.Cleanup(srv.Close)

	server := providerserver.NewProtocol6(New("test")())()
	var ps fwprovider.SchemaResponse
	(&ArtifactAPIProvider{}).Schema(ctx, fwprovider.SchemaRequest{}, &ps)
	pType := ps.Schema.Type().TerraformType(ctx)
	cfg, err := tfprotov6.NewDynamicValue(pType, tftypes.NewValue(pType, map[string]tftypes.Value{
		"endpoint": tftypes.NewValue(tftypes.String, srv.URL),
		"token":    tftypes.NewValue(tftypes.String, "tok"),
	}))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := server.ConfigureProvider(ctx, &tfprotov6.ConfigureProviderRequest{TerraformVersion: "1.9.0", Config: &cfg})
	if err != nil || hasProtoErrors(resp.Diagnostics) {
		t.Fatalf("configure provider: %v %v", err, resp.Diagnostics)
	}
	_, objType := settingsSchema(ctx)
	return &settingsHarness{t: t, ctx: ctx, server: server, objType: objType, fake: fake}
}

func hasProtoErrors(diags []*tfprotov6.Diagnostic) bool {
	for _, d := range diags {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			return true
		}
	}
	return false
}

func (h *settingsHarness) config(set map[string]int64) tftypes.Value {
	attrs := map[string]tftypes.Value{}
	for name, at := range h.objType.AttributeTypes {
		attrs[name] = tftypes.NewValue(at, nil)
	}
	for name, v := range set {
		attrs[name] = tftypes.NewValue(tftypes.Number, v)
	}
	return tftypes.NewValue(h.objType, attrs)
}

// proposed mirrors Terraform core: configured attributes come from config, null computed ones from prior state.
func (h *settingsHarness) proposed(prior, config tftypes.Value) tftypes.Value {
	cfg := h.attrs(config)
	if prior.IsNull() {
		return config
	}
	out := h.attrs(prior)
	for _, name := range settingsIntAttrs {
		out[name] = cfg[name]
	}
	return tftypes.NewValue(h.objType, out)
}

func (h *settingsHarness) attrs(v tftypes.Value) map[string]tftypes.Value {
	h.t.Helper()
	var m map[string]tftypes.Value
	if err := v.As(&m); err != nil {
		h.t.Fatal(err)
	}
	return m
}

func (h *settingsHarness) dv(v tftypes.Value) *tfprotov6.DynamicValue {
	h.t.Helper()
	d, err := tfprotov6.NewDynamicValue(h.objType, v)
	if err != nil {
		h.t.Fatal(err)
	}
	return &d
}

func (h *settingsHarness) value(d *tfprotov6.DynamicValue) tftypes.Value {
	h.t.Helper()
	v, err := d.Unmarshal(h.objType)
	if err != nil {
		h.t.Fatal(err)
	}
	return v
}

func (h *settingsHarness) plan(prior, config tftypes.Value) tftypes.Value {
	h.t.Helper()
	resp, err := h.server.PlanResourceChange(h.ctx, &tfprotov6.PlanResourceChangeRequest{
		TypeName: "artifactapi_settings", PriorState: h.dv(prior), ProposedNewState: h.dv(h.proposed(prior, config)), Config: h.dv(config),
	})
	if err != nil || hasProtoErrors(resp.Diagnostics) {
		h.t.Fatalf("plan: %v %v", err, resp.Diagnostics)
	}
	if len(resp.RequiresReplace) != 0 {
		h.t.Errorf("plan must update in place, got replace on %v", resp.RequiresReplace)
	}
	return h.value(resp.PlannedState)
}

func (h *settingsHarness) apply(prior, planned, config tftypes.Value) tftypes.Value {
	h.t.Helper()
	resp, err := h.server.ApplyResourceChange(h.ctx, &tfprotov6.ApplyResourceChangeRequest{
		TypeName: "artifactapi_settings", PriorState: h.dv(prior), PlannedState: h.dv(planned), Config: h.dv(config),
	})
	if err != nil || hasProtoErrors(resp.Diagnostics) {
		h.t.Fatalf("apply: %v %v", err, resp.Diagnostics)
	}
	return h.value(resp.NewState)
}

func (h *settingsHarness) refresh(state tftypes.Value) tftypes.Value {
	h.t.Helper()
	resp, err := h.server.ReadResource(h.ctx, &tfprotov6.ReadResourceRequest{TypeName: "artifactapi_settings", CurrentState: h.dv(state)})
	if err != nil || hasProtoErrors(resp.Diagnostics) {
		h.t.Fatalf("read: %v %v", err, resp.Diagnostics)
	}
	return h.value(resp.NewState)
}

// converge plans and applies config from prior, as terraform apply does.
func (h *settingsHarness) converge(prior tftypes.Value, set map[string]int64) tftypes.Value {
	cfg := h.config(set)
	return h.apply(prior, h.plan(prior, cfg), cfg)
}

func (h *settingsHarness) intAttr(v tftypes.Value, name string) *int64 {
	h.t.Helper()
	var f *big.Float
	if err := h.attrs(v)[name].As(&f); err != nil {
		h.t.Fatal(err)
	}
	if f == nil {
		return nil
	}
	i, _ := f.Int64()
	return &i
}

func (h *settingsHarness) effective(v tftypes.Value, name string) tftypes.Value {
	var m map[string]tftypes.Value
	if err := h.attrs(v)["effective"].As(&m); err != nil {
		h.t.Fatal(err)
	}
	return m[name]
}

func (h *settingsHarness) lastPut() map[string]any {
	h.t.Helper()
	h.fake.mu.Lock()
	defer h.fake.mu.Unlock()
	if len(h.fake.puts) == 0 {
		h.t.Fatal("no PUT sent")
	}
	return h.fake.puts[len(h.fake.puts)-1]
}

func (h *settingsHarness) putCount() int {
	h.fake.mu.Lock()
	defer h.fake.mu.Unlock()
	return len(h.fake.puts)
}

func assertPutNull(t *testing.T, body map[string]any, name string) {
	t.Helper()
	if v, present := body[name]; present && v != nil {
		t.Errorf("PUT %s: expected null to clear the override, got %v", name, v)
	}
}

func TestSettingsPlan_RemovingAttributeClearsOverride(t *testing.T) {
	h := newSettingsHarness(t)
	state := h.converge(tftypes.NewValue(h.objType, nil), map[string]int64{"s3_health_up_threshold": 2, "s3_health_down_threshold": 5})
	if got := h.intAttr(state, "s3_health_down_threshold"); got == nil || *got != 5 {
		t.Fatalf("down threshold after create: expected 5, got %v", got)
	}

	state = h.refresh(state)
	if planned := h.plan(state, h.config(map[string]int64{"s3_health_up_threshold": 2, "s3_health_down_threshold": 5})); !planned.Equal(state) {
		t.Errorf("unchanged config must plan no changes, got %v", planned)
	}

	cfg := h.config(map[string]int64{"s3_health_up_threshold": 2})
	planned := h.plan(state, cfg)
	if planned.Equal(state) {
		t.Fatal("removing s3_health_down_threshold must plan a change")
	}
	if got := h.intAttr(planned, "s3_health_down_threshold"); got != nil {
		t.Errorf("planned down threshold: expected null, got %d", *got)
	}
	if h.attrs(planned)["effective"].IsKnown() {
		t.Error("effective must be unknown until apply")
	}

	state = h.apply(state, planned, cfg)
	body := h.lastPut()
	assertPutNull(t, body, "s3_health_down_threshold")
	if body["s3_health_up_threshold"] != float64(2) {
		t.Errorf("PUT up threshold: expected 2, got %v", body["s3_health_up_threshold"])
	}
	if !h.effective(state, "s3_health_down_threshold").Equal(tftypes.NewValue(tftypes.Number, 6)) {
		t.Errorf("effective down threshold: expected env default 6, got %v", h.effective(state, "s3_health_down_threshold"))
	}
}

func TestSettingsPlan_RevertsOutOfBandOverride(t *testing.T) {
	h := newSettingsHarness(t)
	set := map[string]int64{"s3_health_up_threshold": 2}
	state := h.converge(tftypes.NewValue(h.objType, nil), set)
	puts := h.putCount()

	h.fake.setOverride("s3_health_interval_ms", 9000)
	state = h.refresh(state)
	if got := h.intAttr(state, "s3_health_interval_ms"); got == nil || *got != 9000 {
		t.Fatalf("refresh must surface the out-of-band override, got %v", got)
	}

	cfg := h.config(set)
	planned := h.plan(state, cfg)
	if got := h.intAttr(planned, "s3_health_interval_ms"); got != nil {
		t.Fatalf("planned interval: expected null to revert the override, got %d", *got)
	}
	state = h.apply(state, planned, cfg)
	if h.putCount() != puts+1 {
		t.Fatalf("expected one PUT to revert, got %d", h.putCount()-puts)
	}
	assertPutNull(t, h.lastPut(), "s3_health_interval_ms")
	if got := h.intAttr(h.refresh(state), "s3_health_interval_ms"); got != nil {
		t.Errorf("override must be cleared on the server, got %d", *got)
	}
}

func TestSettingsPlan_ImportSeesExistingOverrides(t *testing.T) {
	h := newSettingsHarness(t)
	h.fake.setOverride("s3_health_timeout_ms", 1000)

	resp, err := h.server.ImportResourceState(h.ctx, &tfprotov6.ImportResourceStateRequest{TypeName: "artifactapi_settings", ID: "settings"})
	if err != nil || hasProtoErrors(resp.Diagnostics) || len(resp.ImportedResources) != 1 {
		t.Fatalf("import: %v %v", err, resp)
	}
	state := h.refresh(h.value(resp.ImportedResources[0].State))
	if got := h.intAttr(state, "s3_health_timeout_ms"); got == nil || *got != 1000 {
		t.Fatalf("imported timeout: expected 1000, got %v", got)
	}
	if got := h.intAttr(h.plan(state, h.config(nil)), "s3_health_timeout_ms"); got != nil {
		t.Errorf("an empty config must plan the imported override away, got %d", *got)
	}
}
