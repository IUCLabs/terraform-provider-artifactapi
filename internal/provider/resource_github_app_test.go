package provider

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"maps"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

const githubAppType = "artifactapi_github_app"

var (
	testKeysOnce sync.Once
	testKeys     []*rsa.PrivateKey
)

func testRSAKey(t *testing.T, i int) *rsa.PrivateKey {
	t.Helper()
	testKeysOnce.Do(func() {
		for range 2 {
			k, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				panic(err)
			}
			testKeys = append(testKeys, k)
		}
	})
	return testKeys[i]
}

func pkcs1PEM(k *rsa.PrivateKey) string {
	return string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}))
}

func pkcs8PEM(t *testing.T, k any) string {
	t.Helper()
	b, err := x509.MarshalPKCS8PrivateKey(k)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: b}))
}

func expectedFingerprint(t *testing.T, k *rsa.PrivateKey) string {
	t.Helper()
	der, err := x509.MarshalPKIXPublicKey(k.Public())
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(der)
	return "SHA256:" + base64.RawStdEncoding.EncodeToString(sum[:])
}

// fakeGithubApps mimics /api/v2/github-apps: PUT creates or replaces, private_key is never returned.
// PUT responses omit name so the provider must keep the configured one.
type fakeGithubApps struct {
	mu           sync.Mutex
	apps         map[string]map[string]any
	referenced   map[string]bool
	puts         []map[string]any
	clock        int
	createStatus int
}

func (f *fakeGithubApps) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	name, ok := strings.CutPrefix(r.URL.Path, githubAppsPath)
	if !ok || name == "" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		app, ok := f.apps[name]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(app)
	case http.MethodPut:
		b, _ := io.ReadAll(r.Body)
		var body map[string]any
		if err := json.Unmarshal(b, &body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		f.puts = append(f.puts, body)
		key, _ := body["private_key"].(string)
		fp, err := privateKeyFingerprint(key)
		if err != nil {
			http.Error(w, `{"error":"invalid private_key"}`, http.StatusBadRequest)
			return
		}
		f.clock++
		now := fmt.Sprintf("2026-11-28T00:00:%02dZ", f.clock)
		created := now
		prev, existed := f.apps[name]
		if existed {
			created = prev["created_at"].(string)
		}
		base, _ := body["api_base_url"].(string)
		if base == "" {
			base = "https://api.github.com"
		}
		app := map[string]any{
			"name": name, "app_id": body["app_id"], "installation_id": body["installation_id"],
			"api_base_url": base, "description": body["description"], "private_key_fingerprint": fp,
			"created_at": created, "updated_at": now,
		}
		f.apps[name] = app
		out := maps.Clone(app)
		delete(out, "name")
		if !existed && f.createStatus != 0 {
			w.WriteHeader(f.createStatus)
		}
		_ = json.NewEncoder(w).Encode(out)
	case http.MethodDelete:
		if _, ok := f.apps[name]; !ok {
			http.NotFound(w, r)
			return
		}
		if f.referenced[name] {
			http.Error(w, `{"error":"github app is referenced by remote acme-tools"}`, http.StatusConflict)
			return
		}
		delete(f.apps, name)
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (f *fakeGithubApps) rotate(name, fingerprint string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.apps[name]["private_key_fingerprint"] = fingerprint
}

func (f *fakeGithubApps) setReferenced(name string, v bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.referenced[name] = v
}

func (f *fakeGithubApps) exists(name string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.apps[name]
	return ok
}

func (f *fakeGithubApps) lastPut(t *testing.T) map[string]any {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.puts) == 0 {
		t.Fatal("no PUT sent")
	}
	return f.puts[len(f.puts)-1]
}

// resourceHarness drives one resource type through the protocol server against a fake API.
type resourceHarness struct {
	t         *testing.T
	ctx       context.Context
	server    tfprotov6.ProviderServer
	typeName  string
	objType   tftypes.Object
	computed  map[string]bool
	writeOnly map[string]bool
}

func newResourceHarness(t *testing.T, api http.Handler, typeName string, newResource func() resource.Resource) *resourceHarness {
	t.Helper()
	ctx := context.Background()
	srv := httptest.NewServer(api)
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

	var rs resource.SchemaResponse
	newResource().Schema(ctx, resource.SchemaRequest{}, &rs)
	computed, writeOnly := map[string]bool{}, map[string]bool{}
	for name, a := range rs.Schema.Attributes {
		computed[name] = a.IsComputed()
		writeOnly[name] = a.IsWriteOnly()
	}
	return &resourceHarness{
		t: t, ctx: ctx, server: server, typeName: typeName, computed: computed, writeOnly: writeOnly,
		objType: rs.Schema.Type().TerraformType(ctx).(tftypes.Object),
	}
}

type githubAppHarness struct {
	*resourceHarness
	fake *fakeGithubApps
}

func newGithubAppHarness(t *testing.T) *githubAppHarness {
	t.Helper()
	fake := &fakeGithubApps{apps: map[string]map[string]any{}, referenced: map[string]bool{}}
	return &githubAppHarness{resourceHarness: newResourceHarness(t, fake, githubAppType, NewGithubAppResource), fake: fake}
}

func (h *resourceHarness) config(set map[string]any) tftypes.Value {
	attrs := map[string]tftypes.Value{}
	for name, at := range h.objType.AttributeTypes {
		attrs[name] = tftypes.NewValue(at, nil)
	}
	for name, v := range set {
		attrs[name] = tftypes.NewValue(h.objType.AttributeTypes[name], v)
	}
	return tftypes.NewValue(h.objType, attrs)
}

func (h *githubAppHarness) baseConfig(key string, extra map[string]any) tftypes.Value {
	set := map[string]any{"name": "acme", "app_id": 123, "installation_id": 456, "private_key": key}
	for k, v := range extra {
		set[k] = v
	}
	return h.config(set)
}

func (h *resourceHarness) attrs(v tftypes.Value) map[string]tftypes.Value {
	h.t.Helper()
	var m map[string]tftypes.Value
	if err := v.As(&m); err != nil {
		h.t.Fatal(err)
	}
	return m
}

// proposed mirrors Terraform core: config wins, null computed attributes carry the prior state, write-only attributes are null.
func (h *resourceHarness) proposed(prior, config tftypes.Value) tftypes.Value {
	var p map[string]tftypes.Value
	if !prior.IsNull() {
		p = h.attrs(prior)
	}
	out := map[string]tftypes.Value{}
	for name, v := range h.attrs(config) {
		out[name] = v
		switch {
		case h.writeOnly[name]:
			out[name] = tftypes.NewValue(v.Type(), nil)
		case v.IsNull() && h.computed[name] && p != nil:
			out[name] = p[name]
		}
	}
	return tftypes.NewValue(h.objType, out)
}

func (h *resourceHarness) dv(v tftypes.Value) *tfprotov6.DynamicValue {
	h.t.Helper()
	d, err := tfprotov6.NewDynamicValue(h.objType, v)
	if err != nil {
		h.t.Fatal(err)
	}
	return &d
}

func (h *resourceHarness) value(d *tfprotov6.DynamicValue) tftypes.Value {
	h.t.Helper()
	v, err := d.Unmarshal(h.objType)
	if err != nil {
		h.t.Fatal(err)
	}
	return v
}

func (h *resourceHarness) plan(prior, config tftypes.Value) tftypes.Value {
	h.t.Helper()
	resp, err := h.server.PlanResourceChange(h.ctx, &tfprotov6.PlanResourceChangeRequest{
		TypeName: h.typeName, PriorState: h.dv(prior), ProposedNewState: h.dv(h.proposed(prior, config)), Config: h.dv(config),
	})
	if err != nil || hasProtoErrors(resp.Diagnostics) {
		h.t.Fatalf("plan: %v %v", err, resp.Diagnostics)
	}
	if len(resp.RequiresReplace) != 0 {
		h.t.Errorf("plan must update in place, got replace on %v", resp.RequiresReplace)
	}
	return h.value(resp.PlannedState)
}

func (h *resourceHarness) applyDiags(prior, planned, config tftypes.Value) (tftypes.Value, []*tfprotov6.Diagnostic) {
	h.t.Helper()
	resp, err := h.server.ApplyResourceChange(h.ctx, &tfprotov6.ApplyResourceChangeRequest{
		TypeName: h.typeName, PriorState: h.dv(prior), PlannedState: h.dv(planned), Config: h.dv(config),
	})
	if err != nil {
		h.t.Fatal(err)
	}
	return h.value(resp.NewState), resp.Diagnostics
}

func (h *resourceHarness) apply(prior, planned, config tftypes.Value) tftypes.Value {
	h.t.Helper()
	v, diags := h.applyDiags(prior, planned, config)
	if hasProtoErrors(diags) {
		h.t.Fatalf("apply: %v", diags)
	}
	return v
}

func (h *resourceHarness) refresh(state tftypes.Value) tftypes.Value {
	h.t.Helper()
	resp, err := h.server.ReadResource(h.ctx, &tfprotov6.ReadResourceRequest{TypeName: h.typeName, CurrentState: h.dv(state)})
	if err != nil || hasProtoErrors(resp.Diagnostics) {
		h.t.Fatalf("read: %v %v", err, resp.Diagnostics)
	}
	return h.value(resp.NewState)
}

func (h *resourceHarness) converge(prior, config tftypes.Value) tftypes.Value {
	return h.apply(prior, h.plan(prior, config), config)
}

func (h *resourceHarness) str(v tftypes.Value, name string) string {
	h.t.Helper()
	var s *string
	if err := h.attrs(v)[name].As(&s); err != nil {
		h.t.Fatalf("%s: %v", name, err)
	}
	if s == nil {
		return "<null>"
	}
	return *s
}

func (h *resourceHarness) null() tftypes.Value { return tftypes.NewValue(h.objType, nil) }

func TestGithubAppResource_Metadata(t *testing.T) {
	if got := resourceTypeName(NewGithubAppResource); got != githubAppType {
		t.Errorf("expected %s, got %s", githubAppType, got)
	}
}

func TestGithubAppResource_Schema(t *testing.T) {
	var resp resource.SchemaResponse
	NewGithubAppResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	a := resp.Schema.Attributes
	for _, name := range []string{"name", "app_id", "installation_id"} {
		if !a[name].IsRequired() {
			t.Errorf("%s must be required", name)
		}
	}
	for _, name := range []string{"private_key", "private_key_wo"} {
		if !a[name].IsOptional() || !a[name].IsSensitive() || a[name].IsComputed() {
			t.Errorf("%s must be optional, sensitive and not computed", name)
		}
	}
	if !a["private_key_wo"].IsWriteOnly() || a["private_key"].IsWriteOnly() {
		t.Error("only private_key_wo must be write-only")
	}
	if !a["private_key_wo_version"].IsOptional() || a["private_key_wo_version"].IsWriteOnly() {
		t.Error("private_key_wo_version must be optional and stored in state")
	}
	if !a["api_base_url"].IsOptional() || !a["api_base_url"].IsComputed() {
		t.Error("api_base_url must be optional and computed")
	}
	for _, name := range []string{"private_key_fingerprint", "created_at", "updated_at"} {
		if !a[name].IsComputed() || a[name].IsOptional() {
			t.Errorf("%s must be computed only", name)
		}
	}
}

func TestPrivateKeyFingerprint(t *testing.T) {
	k := testRSAKey(t, 0)
	want := expectedFingerprint(t, k)
	for name, s := range map[string]string{"pkcs1": pkcs1PEM(k), "pkcs8": pkcs8PEM(t, k)} {
		got, err := privateKeyFingerprint(s)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got != want {
			t.Errorf("%s: expected %s, got %s", name, want, got)
		}
		if strings.HasSuffix(got, "=") {
			t.Errorf("%s: fingerprint must be unpadded: %s", name, got)
		}
	}
	if other, _ := privateKeyFingerprint(pkcs1PEM(testRSAKey(t, 1))); other == want {
		t.Error("different keys must have different fingerprints")
	}
}

func TestGithubAppResource_ValidateConfig(t *testing.T) {
	ctx := context.Background()
	k := testRSAKey(t, 0)
	ec, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	unknown := tftypes.UnknownValue
	cases := []struct {
		name    string
		set     map[string]any
		wantErr string
	}{
		{"pkcs1", nil, ""},
		{"pkcs8", map[string]any{"private_key": pkcs8PEM(t, k)}, ""},
		{"ghes", map[string]any{"api_base_url": "https://ghe.example.com/api/v3"}, ""},
		{"unknown-values", map[string]any{"app_id": unknown, "private_key": unknown, "api_base_url": unknown}, ""},
		{"app-id-zero", map[string]any{"app_id": 0}, "app_id"},
		{"app-id-negative", map[string]any{"app_id": -5}, "app_id"},
		{"installation-id-zero", map[string]any{"installation_id": 0}, "installation_id"},
		{"key-not-pem", map[string]any{"private_key": "not a key"}, "private_key"},
		{"key-not-rsa", map[string]any{"private_key": pkcs8PEM(t, ec)}, "private_key"},
		{"key-garbage-der", map[string]any{"private_key": "-----BEGIN RSA PRIVATE KEY-----\nAAAA\n-----END RSA PRIVATE KEY-----\n"}, "private_key"},
		{"url-http", map[string]any{"api_base_url": "http://api.github.com"}, "api_base_url"},
		{"url-relative", map[string]any{"api_base_url": "/api/v3"}, "api_base_url"},
		{"url-empty", map[string]any{"api_base_url": ""}, "api_base_url"},
		{"url-trailing-slash", map[string]any{"api_base_url": "https://ghe.example.com/api/v3/"}, "trailing slash"},
		{"url-root-slash", map[string]any{"api_base_url": "https://api.github.com/"}, "trailing slash"},
		{"name-dots-dashes", map[string]any{"name": "Acme_bot-1.prod"}, ""},
		{"name-max-length", map[string]any{"name": "a" + strings.Repeat("b", 127)}, ""},
		{"name-unknown", map[string]any{"name": unknown}, ""},
		{"name-empty", map[string]any{"name": ""}, "name"},
		{"name-leading-dot", map[string]any{"name": ".acme"}, "name"},
		{"name-slash", map[string]any{"name": "acme/bot"}, "name"},
		{"name-space", map[string]any{"name": "acme bot"}, "name"},
		{"name-too-long", map[string]any{"name": "a" + strings.Repeat("b", 128)}, "name"},
		{"wo-only", map[string]any{"private_key": nil, "private_key_wo": pkcs8PEM(t, k)}, ""},
		{"wo-with-version", map[string]any{"private_key": nil, "private_key_wo": pkcs1PEM(k), "private_key_wo_version": 3}, ""},
		{"wo-unknown", map[string]any{"private_key": nil, "private_key_wo": unknown}, ""},
		{"key-unknown-and-wo", map[string]any{"private_key": unknown, "private_key_wo": pkcs1PEM(k)}, ""},
		{"wo-not-pem", map[string]any{"private_key": nil, "private_key_wo": "not a key"}, "private_key_wo"},
		{"wo-not-rsa", map[string]any{"private_key": nil, "private_key_wo": pkcs8PEM(t, ec)}, "private_key_wo"},
		{"both-keys", map[string]any{"private_key_wo": pkcs1PEM(k)}, "Exactly one of private_key_wo or private_key"},
		{"no-key", map[string]any{"private_key": nil}, "Exactly one of private_key_wo or private_key"},
		{"version-without-wo", map[string]any{"private_key_wo_version": 1}, "private_key_wo_version only applies"},
	}
	r := NewGithubAppResource().(*githubAppResource)
	var sr resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &sr)
	objType := sr.Schema.Type().TerraformType(ctx).(tftypes.Object)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			set := map[string]any{"name": "acme", "app_id": 1, "installation_id": 2, "private_key": pkcs1PEM(k)}
			for k, v := range tc.set {
				set[k] = v
			}
			attrs := map[string]tftypes.Value{}
			for name, at := range objType.AttributeTypes {
				attrs[name] = tftypes.NewValue(at, nil)
			}
			for name, v := range set {
				attrs[name] = tftypes.NewValue(objType.AttributeTypes[name], v)
			}
			var resp resource.ValidateConfigResponse
			r.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: sr.Schema, Raw: tftypes.NewValue(objType, attrs)}}, &resp)
			errs := resp.Diagnostics.Errors()
			if tc.wantErr == "" {
				if len(errs) != 0 {
					t.Fatalf("unexpected errors: %v", errs)
				}
				return
			}
			if len(errs) != 1 {
				t.Fatalf("expected one error on %s, got %v", tc.wantErr, errs)
			}
			if !strings.Contains(errs[0].Detail(), tc.wantErr) {
				t.Errorf("error must name %s: %s", tc.wantErr, errs[0].Detail())
			}
			if strings.Contains(errs[0].Detail(), "PRIVATE KEY") {
				t.Errorf("error must not echo the key: %s", errs[0].Detail())
			}
		})
	}
}

func TestGithubAppPlan_CreateReadUpdate(t *testing.T) {
	h := newGithubAppHarness(t)
	k := testRSAKey(t, 0)
	key := pkcs1PEM(k)
	cfg := h.baseConfig(key, map[string]any{"description": "acme bot"})

	planned := h.plan(h.null(), cfg)
	if got := h.str(planned, "private_key_fingerprint"); got != expectedFingerprint(t, k) {
		t.Errorf("planned fingerprint must be known from config: %s", got)
	}
	if h.attrs(planned)["api_base_url"].IsKnown() {
		t.Error("unset api_base_url must be unknown until the server defaults it")
	}
	state := h.apply(h.null(), planned, cfg)

	put := h.fake.lastPut(t)
	if put["app_id"] != float64(123) || put["installation_id"] != float64(456) || put["private_key"] != key || put["description"] != "acme bot" {
		t.Errorf("unexpected PUT body: %v", put)
	}
	if _, ok := put["api_base_url"]; ok {
		t.Errorf("unset api_base_url must be omitted so the server defaults it: %v", put)
	}
	if h.str(state, "private_key") != key {
		t.Error("state must keep the configured private_key")
	}
	if h.str(state, "api_base_url") != "https://api.github.com" {
		t.Errorf("api_base_url: expected server default, got %s", h.str(state, "api_base_url"))
	}
	created := h.str(state, "created_at")

	state = h.refresh(state)
	if h.str(state, "private_key") != key {
		t.Error("refresh must keep a private_key matching the server fingerprint")
	}
	if p := h.plan(state, cfg); !p.Equal(state) {
		t.Errorf("unchanged config must plan no changes, got %v", p)
	}

	cfg = h.baseConfig(key, map[string]any{"description": "renamed"})
	planned = h.plan(state, cfg)
	if h.attrs(planned)["updated_at"].IsKnown() || h.str(planned, "created_at") != created || h.str(planned, "api_base_url") != "https://api.github.com" {
		t.Errorf("update must keep created_at/api_base_url and recompute updated_at: %v", planned)
	}
	state = h.apply(state, planned, cfg)
	if h.str(state, "description") != "renamed" || h.str(state, "created_at") != created || h.str(state, "updated_at") == created {
		t.Errorf("unexpected state after update: %v", state)
	}
	if h.fake.lastPut(t)["api_base_url"] != "https://api.github.com" {
		t.Error("update must resend the known api_base_url")
	}
}

func TestGithubAppPlan_FingerprintDriftPlansKeyRewrite(t *testing.T) {
	h := newGithubAppHarness(t)
	k := testRSAKey(t, 0)
	key := pkcs1PEM(k)
	cfg := h.baseConfig(key, nil)
	state := h.converge(h.null(), cfg)

	rotated := expectedFingerprint(t, testRSAKey(t, 1))
	h.fake.rotate("acme", rotated)
	state = h.refresh(state)
	if !h.attrs(state)["private_key"].IsNull() {
		t.Fatal("a fingerprint mismatch must clear private_key from state")
	}
	if h.str(state, "private_key_fingerprint") != rotated {
		t.Errorf("state must report the server fingerprint, got %s", h.str(state, "private_key_fingerprint"))
	}

	planned := h.plan(state, cfg)
	if planned.Equal(state) {
		t.Fatal("drift must plan an update")
	}
	if h.str(planned, "private_key") != key || h.str(planned, "private_key_fingerprint") != expectedFingerprint(t, k) {
		t.Errorf("plan must rewrite the configured key: %v", planned)
	}
	state = h.apply(state, planned, cfg)
	if h.fake.lastPut(t)["private_key"] != key {
		t.Error("apply must PUT the configured key")
	}
	if p := h.plan(h.refresh(state), cfg); !p.Equal(h.refresh(state)) {
		t.Error("after rewriting the key the plan must be clean")
	}
}

func TestGithubAppPlan_KeyChangeIsInPlaceUpdate(t *testing.T) {
	h := newGithubAppHarness(t)
	state := h.converge(h.null(), h.baseConfig(pkcs1PEM(testRSAKey(t, 0)), nil))
	k2 := testRSAKey(t, 1)
	cfg := h.baseConfig(pkcs8PEM(t, k2), nil)
	planned := h.plan(state, cfg)
	if h.str(planned, "private_key_fingerprint") != expectedFingerprint(t, k2) {
		t.Errorf("planned fingerprint must follow the new key: %s", h.str(planned, "private_key_fingerprint"))
	}
	h.apply(state, planned, cfg)
}

func TestGithubAppPlan_ImportRewritesKey(t *testing.T) {
	h := newGithubAppHarness(t)
	key := pkcs1PEM(testRSAKey(t, 0))
	cfg := h.baseConfig(key, map[string]any{"api_base_url": "https://ghe.example.com/api/v3"})
	h.converge(h.null(), cfg)

	resp, err := h.server.ImportResourceState(h.ctx, &tfprotov6.ImportResourceStateRequest{TypeName: githubAppType, ID: "acme"})
	if err != nil || hasProtoErrors(resp.Diagnostics) || len(resp.ImportedResources) != 1 {
		t.Fatalf("import: %v %v", err, resp)
	}
	state := h.refresh(h.value(resp.ImportedResources[0].State))
	if h.str(state, "api_base_url") != "https://ghe.example.com/api/v3" || h.str(state, "name") != "acme" {
		t.Errorf("import must read server attributes: %v", state)
	}
	if !h.attrs(state)["private_key"].IsNull() {
		t.Fatal("imported private_key must be null")
	}
	if planned := h.plan(state, cfg); h.str(planned, "private_key") != key {
		t.Error("the apply after import must write the configured key")
	}
}

func TestGithubAppResource_CreateRefusesExisting(t *testing.T) {
	h := newGithubAppHarness(t)
	cfg := h.baseConfig(pkcs1PEM(testRSAKey(t, 0)), nil)
	h.converge(h.null(), cfg)
	_, diags := h.applyDiags(h.null(), h.plan(h.null(), cfg), cfg)
	if !hasProtoErrors(diags) || !strings.Contains(diags[0].Detail, "import") {
		t.Errorf("create over an existing app must fail and point at import, got %v", diags)
	}
}

func TestGithubAppResource_ReadMissingRemovesState(t *testing.T) {
	h := newGithubAppHarness(t)
	state := h.converge(h.null(), h.baseConfig(pkcs1PEM(testRSAKey(t, 0)), nil))
	h.fake.mu.Lock()
	delete(h.fake.apps, "acme")
	h.fake.mu.Unlock()
	if !h.refresh(state).IsNull() {
		t.Error("a missing app must be removed from state")
	}
}

func TestGithubAppResource_Delete(t *testing.T) {
	h := newGithubAppHarness(t)
	state := h.converge(h.null(), h.baseConfig(pkcs1PEM(testRSAKey(t, 0)), nil))

	h.fake.setReferenced("acme", true)
	_, diags := h.applyDiags(state, h.null(), h.null())
	if !hasProtoErrors(diags) {
		t.Fatal("deleting a referenced app must fail")
	}
	if d := diags[0]; !strings.Contains(d.Summary, "referenced") || !strings.Contains(d.Detail, "referenced by remote acme-tools") {
		t.Errorf("409 must surface the server error: %s: %s", d.Summary, d.Detail)
	}
	if !h.fake.exists("acme") {
		t.Fatal("app must survive a rejected delete")
	}

	h.fake.setReferenced("acme", false)
	if _, diags := h.applyDiags(state, h.null(), h.null()); hasProtoErrors(diags) {
		t.Fatalf("delete: %v", diags)
	}
	if h.fake.exists("acme") {
		t.Error("app must be deleted")
	}
	if _, diags := h.applyDiags(state, h.null(), h.null()); hasProtoErrors(diags) {
		t.Errorf("deleting an already-missing app must succeed: %v", diags)
	}
}

func TestGithubAppResource_ServerErrorSurfaces(t *testing.T) {
	h := newGithubAppHarness(t)
	cfg := h.baseConfig(pkcs1PEM(testRSAKey(t, 0)), nil)
	planned := h.plan(h.null(), cfg)
	attrs := maps.Clone(h.attrs(planned))
	attrs["private_key"] = tftypes.NewValue(tftypes.String, "bogus")
	bad := tftypes.NewValue(h.objType, attrs)
	_, diags := h.applyDiags(h.null(), bad, bad)
	if !hasProtoErrors(diags) || !strings.Contains(diags[0].Detail, "api error 400") {
		t.Errorf("server rejection must surface: %v", diags)
	}
}

func TestGithubAppResource_CreateAccepts201(t *testing.T) {
	h := newGithubAppHarness(t)
	h.fake.createStatus = http.StatusCreated
	state := h.converge(h.null(), h.baseConfig(pkcs1PEM(testRSAKey(t, 0)), nil))
	if h.str(state, "name") != "acme" || h.str(state, "api_base_url") != "https://api.github.com" {
		t.Errorf("201 create must be read like 200: %v", state)
	}
}

func (h *githubAppHarness) woConfig(key string, extra map[string]any) tftypes.Value {
	set := map[string]any{"name": "acme", "app_id": 123, "installation_id": 456, "private_key_wo": key}
	for k, v := range extra {
		set[k] = v
	}
	return h.config(set)
}

func (h *githubAppHarness) putCount() int {
	h.fake.mu.Lock()
	defer h.fake.mu.Unlock()
	return len(h.fake.puts)
}

func (h *githubAppHarness) assertNoKeyIn(what string, v tftypes.Value) {
	h.t.Helper()
	a := h.attrs(v)
	if !a["private_key_wo"].IsNull() || !a["private_key"].IsNull() {
		h.t.Errorf("%s must not hold the private key: private_key=%v private_key_wo=%v", what, a["private_key"], a["private_key_wo"])
	}
}

func TestGithubAppResource_WriteOnlyNeedsTerraform111(t *testing.T) {
	h := newGithubAppHarness(t)
	cfg := h.woConfig(pkcs1PEM(testRSAKey(t, 0)), nil)
	for _, allowed := range []bool{false, true} {
		resp, err := h.server.ValidateResourceConfig(h.ctx, &tfprotov6.ValidateResourceConfigRequest{
			TypeName: githubAppType, Config: h.dv(cfg),
			ClientCapabilities: &tfprotov6.ValidateResourceConfigClientCapabilities{WriteOnlyAttributesAllowed: allowed},
		})
		if err != nil {
			t.Fatal(err)
		}
		if got := hasProtoErrors(resp.Diagnostics); got == allowed {
			t.Errorf("write-only allowed=%v: unexpected diagnostics %v", allowed, resp.Diagnostics)
		}
		if !allowed && !strings.Contains(resp.Diagnostics[0].Detail, "Terraform 1.11") {
			t.Errorf("rejection must name the Terraform version: %s", resp.Diagnostics[0].Detail)
		}
	}
}

func TestGithubAppWriteOnly_CreateReadUpdate(t *testing.T) {
	h := newGithubAppHarness(t)
	k := testRSAKey(t, 0)
	key := pkcs1PEM(k)
	cfg := h.woConfig(key, map[string]any{"description": "acme bot"})

	planned := h.plan(h.null(), cfg)
	h.assertNoKeyIn("plan", planned)
	if got := h.str(planned, "private_key_fingerprint"); got != expectedFingerprint(t, k) {
		t.Errorf("planned fingerprint must be computed from the write-only key: %s", got)
	}
	state := h.apply(h.null(), planned, cfg)
	h.assertNoKeyIn("state", state)
	if put := h.fake.lastPut(t); put["private_key"] != key {
		t.Errorf("create must PUT the write-only key, got %v", put["private_key"])
	}
	if h.str(state, "private_key_fingerprint") != expectedFingerprint(t, k) {
		t.Errorf("state fingerprint: %s", h.str(state, "private_key_fingerprint"))
	}

	state = h.refresh(state)
	h.assertNoKeyIn("refreshed state", state)
	if p := h.plan(state, cfg); !p.Equal(state) {
		t.Errorf("unchanged config must plan no changes, got %v", p)
	}

	cfg = h.woConfig(key, map[string]any{"description": "renamed"})
	puts := h.putCount()
	state = h.converge(state, cfg)
	h.assertNoKeyIn("updated state", state)
	if h.putCount() != puts+1 || h.fake.lastPut(t)["private_key"] != key {
		t.Error("update must resend the write-only key in the PUT body")
	}
}

func TestGithubAppWriteOnly_KeyRotationPlansUpdate(t *testing.T) {
	h := newGithubAppHarness(t)
	state := h.converge(h.null(), h.woConfig(pkcs1PEM(testRSAKey(t, 0)), nil))

	k2 := testRSAKey(t, 1)
	key2 := pkcs8PEM(t, k2)
	cfg := h.woConfig(key2, nil)
	planned := h.plan(state, cfg)
	if planned.Equal(state) {
		t.Fatal("a changed write-only key must plan an update")
	}
	if h.str(planned, "private_key_fingerprint") != expectedFingerprint(t, k2) || h.attrs(planned)["updated_at"].IsKnown() {
		t.Errorf("plan must show the new fingerprint and recompute updated_at: %v", planned)
	}
	state = h.apply(state, planned, cfg)
	h.assertNoKeyIn("state", state)
	if h.fake.lastPut(t)["private_key"] != key2 {
		t.Error("apply must PUT the rotated key")
	}
	if p := h.plan(h.refresh(state), cfg); !p.Equal(h.refresh(state)) {
		t.Error("after rotation the plan must be clean")
	}
}

func TestGithubAppWriteOnly_ServerDriftPlansRewrite(t *testing.T) {
	h := newGithubAppHarness(t)
	k := testRSAKey(t, 0)
	key := pkcs1PEM(k)
	cfg := h.woConfig(key, nil)
	state := h.converge(h.null(), cfg)

	rotated := expectedFingerprint(t, testRSAKey(t, 1))
	h.fake.rotate("acme", rotated)
	state = h.refresh(state)
	if h.str(state, "private_key_fingerprint") != rotated {
		t.Fatalf("refresh must record the server fingerprint, got %s", h.str(state, "private_key_fingerprint"))
	}
	planned := h.plan(state, cfg)
	if h.str(planned, "private_key_fingerprint") != expectedFingerprint(t, k) {
		t.Fatalf("drift must plan the configured key's fingerprint: %v", planned)
	}
	h.apply(state, planned, cfg)
	if h.fake.lastPut(t)["private_key"] != key {
		t.Error("apply must PUT the configured key")
	}
}

func TestGithubAppWriteOnly_VersionBumpResends(t *testing.T) {
	h := newGithubAppHarness(t)
	key := pkcs1PEM(testRSAKey(t, 0))
	state := h.converge(h.null(), h.woConfig(key, map[string]any{"private_key_wo_version": 1}))
	if p := h.plan(state, h.woConfig(key, map[string]any{"private_key_wo_version": 1})); !p.Equal(state) {
		t.Fatal("same version and key must plan no changes")
	}

	cfg := h.woConfig(key, map[string]any{"private_key_wo_version": 2})
	planned := h.plan(state, cfg)
	if planned.Equal(state) {
		t.Fatal("a version bump must plan an update")
	}
	puts := h.putCount()
	state = h.apply(state, planned, cfg)
	h.assertNoKeyIn("state", state)
	if h.putCount() != puts+1 || h.fake.lastPut(t)["private_key"] != key {
		t.Error("a version bump must resend the key")
	}
	var v *big.Float
	if err := h.attrs(state)["private_key_wo_version"].As(&v); err != nil || v == nil || v.String() != "2" {
		t.Errorf("state must record private_key_wo_version 2, got %v", v)
	}
}

func TestGithubAppWriteOnly_Import(t *testing.T) {
	h := newGithubAppHarness(t)
	key := pkcs1PEM(testRSAKey(t, 0))
	h.converge(h.null(), h.woConfig(key, nil))

	resp, err := h.server.ImportResourceState(h.ctx, &tfprotov6.ImportResourceStateRequest{TypeName: githubAppType, ID: "acme"})
	if err != nil || hasProtoErrors(resp.Diagnostics) || len(resp.ImportedResources) != 1 {
		t.Fatalf("import: %v %v", err, resp)
	}
	state := h.refresh(h.value(resp.ImportedResources[0].State))
	h.assertNoKeyIn("imported state", state)
	if p := h.plan(state, h.woConfig(key, nil)); !p.Equal(state) {
		t.Errorf("a matching key after import must plan no changes, got %v", p)
	}
	if p := h.plan(state, h.woConfig(key, map[string]any{"private_key_wo_version": 1})); p.Equal(state) {
		t.Error("setting private_key_wo_version after import must plan a resend")
	}
}

func TestGithubAppWriteOnly_MigrationDropsStoredKey(t *testing.T) {
	h := newGithubAppHarness(t)
	key := pkcs1PEM(testRSAKey(t, 0))
	state := h.converge(h.null(), h.baseConfig(key, nil))
	if h.str(state, "private_key") != key {
		t.Fatal("private_key must be stored in state")
	}
	state = h.converge(state, h.woConfig(key, nil))
	h.assertNoKeyIn("state after switching to private_key_wo", state)
	if h.fake.lastPut(t)["private_key"] != key {
		t.Error("switching to private_key_wo must PUT the key")
	}
}
