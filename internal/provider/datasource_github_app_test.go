package provider

import (
	"context"
	"math/big"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type githubAppDataSourceHarness struct {
	*resourceHarness
	fake    *fakeGithubApps
	dsType  tftypes.Object
	dsAttrs map[string]tftypes.Type
}

func newGithubAppDataSourceHarness(t *testing.T) *githubAppDataSourceHarness {
	t.Helper()
	fake := &fakeGithubApps{apps: map[string]map[string]any{}, referenced: map[string]bool{}}
	rh := newResourceHarness(t, fake, githubAppType, NewGithubAppResource)
	var s datasource.SchemaResponse
	NewGithubAppDataSource().Schema(context.Background(), datasource.SchemaRequest{}, &s)
	objType := s.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	return &githubAppDataSourceHarness{resourceHarness: rh, fake: fake, dsType: objType, dsAttrs: objType.AttributeTypes}
}

func (h *githubAppDataSourceHarness) read(name string) (map[string]tftypes.Value, []*tfprotov6.Diagnostic) {
	h.t.Helper()
	attrs := map[string]tftypes.Value{}
	for n, at := range h.dsAttrs {
		attrs[n] = tftypes.NewValue(at, nil)
	}
	attrs["name"] = tftypes.NewValue(tftypes.String, name)
	cfg, err := tfprotov6.NewDynamicValue(h.dsType, tftypes.NewValue(h.dsType, attrs))
	if err != nil {
		h.t.Fatal(err)
	}
	resp, err := h.server.ReadDataSource(h.ctx, &tfprotov6.ReadDataSourceRequest{TypeName: githubAppType, Config: &cfg})
	if err != nil {
		h.t.Fatal(err)
	}
	if resp.State == nil {
		return nil, resp.Diagnostics
	}
	v, err := resp.State.Unmarshal(h.dsType)
	if err != nil {
		h.t.Fatal(err)
	}
	var out map[string]tftypes.Value
	if !v.IsNull() {
		if err := v.As(&out); err != nil {
			h.t.Fatal(err)
		}
	}
	return out, resp.Diagnostics
}

func (h *githubAppDataSourceHarness) seed(name string, app map[string]any) {
	h.fake.mu.Lock()
	defer h.fake.mu.Unlock()
	h.fake.apps[name] = app
}

func uiApp() map[string]any {
	return map[string]any{
		"name": "ui-app", "app_id": 98765, "installation_id": 4242,
		"api_base_url": "https://ghe.example.com/api/v3", "description": "created in UI",
		"private_key_fingerprint": "SHA256:abc", "created_at": "2026-11-28T01:00:00Z", "updated_at": "2026-11-28T02:00:00Z",
	}
}

func int64Attr(t *testing.T, v tftypes.Value) *int64 {
	t.Helper()
	var f *big.Float
	if err := v.As(&f); err != nil {
		t.Fatal(err)
	}
	if f == nil {
		return nil
	}
	n, _ := f.Int64()
	return &n
}

func TestGithubAppDataSource_Metadata(t *testing.T) {
	if got := dataSourceTypeName(NewGithubAppDataSource); got != githubAppType {
		t.Errorf("expected %s, got %s", githubAppType, got)
	}
}

func TestGithubAppDataSource_SchemaHasNoPrivateKey(t *testing.T) {
	var s datasource.SchemaResponse
	NewGithubAppDataSource().Schema(context.Background(), datasource.SchemaRequest{}, &s)
	if _, ok := s.Schema.Attributes["private_key"]; ok {
		t.Error("data source must not expose private_key")
	}
	if !s.Schema.Attributes["name"].IsRequired() {
		t.Error("name must be required")
	}
}

func TestGithubAppDataSource_Found(t *testing.T) {
	h := newGithubAppDataSourceHarness(t)
	h.seed("ui-app", uiApp())

	got, diags := h.read("ui-app")
	if hasProtoErrors(diags) {
		t.Fatalf("read: %v", diags)
	}
	for name, want := range map[string]string{
		"name": "ui-app", "api_base_url": "https://ghe.example.com/api/v3", "description": "created in UI",
		"private_key_fingerprint": "SHA256:abc", "created_at": "2026-11-28T01:00:00Z", "updated_at": "2026-11-28T02:00:00Z",
	} {
		var s string
		if err := got[name].As(&s); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if s != want {
			t.Errorf("%s: expected %q, got %q", name, want, s)
		}
	}
	if n := int64Attr(t, got["app_id"]); n == nil || *n != 98765 {
		t.Errorf("app_id: expected 98765, got %v", n)
	}
	if n := int64Attr(t, got["installation_id"]); n == nil || *n != 4242 {
		t.Errorf("installation_id: expected 4242, got %v", n)
	}
}

func TestGithubAppDataSource_NotFound(t *testing.T) {
	h := newGithubAppDataSourceHarness(t)

	_, diags := h.read("missing")
	if !hasProtoErrors(diags) {
		t.Fatal("expected an error for a missing app")
	}
	if diags[0].Summary != "github app not found" || !strings.Contains(diags[0].Detail, `"missing"`) {
		t.Errorf("unexpected diagnostic: %s: %s", diags[0].Summary, diags[0].Detail)
	}
}

func TestGithubAppDataSource_UninstalledIsNull(t *testing.T) {
	for name, mutate := range map[string]func(map[string]any){
		"null":    func(a map[string]any) { a["installation_id"] = nil },
		"omitted": func(a map[string]any) { delete(a, "installation_id") },
		"zero":    func(a map[string]any) { a["installation_id"] = 0 },
	} {
		t.Run(name, func(t *testing.T) {
			h := newGithubAppDataSourceHarness(t)
			app := uiApp()
			mutate(app)
			h.seed("ui-app", app)

			got, diags := h.read("ui-app")
			if hasProtoErrors(diags) {
				t.Fatalf("read: %v", diags)
			}
			if !got["installation_id"].IsNull() {
				t.Errorf("installation_id must be null, got %v", got["installation_id"])
			}
			if n := int64Attr(t, got["app_id"]); n == nil || *n != 98765 {
				t.Errorf("app_id: expected 98765, got %v", n)
			}
		})
	}
}
