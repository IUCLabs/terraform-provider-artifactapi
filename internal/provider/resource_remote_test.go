package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestModelToAPI_FullFields(t *testing.T) {
	ctx := context.Background()
	r := &remoteResource{packageType: "docker"}

	model := remoteResourceModel{
		Name:              types.StringValue("my-remote"),
		BaseURL:           types.StringValue("https://registry.example.com"),
		Description:       types.StringValue("A test remote"),
		Username:          types.StringValue("user"),
		Password:          types.StringValue("pass"),
		ImmutableTTL:      types.Int64Value(86400),
		MutableTTL:        types.Int64Value(3600),
		CheckMutable:      types.BoolValue(true),
		Patterns:          stringsToList(ctx, []string{"*.tar.gz", "*.whl"}),
		Blocklist:         stringsToList(ctx, []string{"blocked/*"}),
		MutablePatterns:   stringsToList(ctx, []string{"latest"}),
		ImmutablePatterns: stringsToList(ctx, []string{"v*"}),
		BanTagsEnabled:    types.BoolValue(true),
		BanTags:           stringsToList(ctx, []string{"latest", "dev"}),
		QuarantineEnabled: types.BoolValue(true),
		QuarantineDays:    types.Int64Value(7),
		StaleOnError:      types.BoolValue(false),
		ReleasesRemote:    types.StringValue("cdn-remote"),

		UpstreamDialTimeout:           types.Int64Value(3),
		UpstreamTLSTimeout:            types.Int64Value(4),
		UpstreamResponseHeaderTimeout: types.Int64Value(5),
	}

	api := r.modelToAPI(ctx, model)

	if api.UpstreamDialTimeout != 3 || api.UpstreamTLSTimeout != 4 || api.UpstreamResponseHeaderTimeout != 5 {
		t.Errorf("upstream timeouts: got %d/%d/%d, want 3/4/5",
			api.UpstreamDialTimeout, api.UpstreamTLSTimeout, api.UpstreamResponseHeaderTimeout)
	}
	if api.Name != "my-remote" {
		t.Errorf("Name: expected my-remote, got %s", api.Name)
	}
	if api.PackageType != "docker" {
		t.Errorf("PackageType: expected docker, got %s", api.PackageType)
	}
	if api.BaseURL != "https://registry.example.com" {
		t.Errorf("BaseURL: expected https://registry.example.com, got %s", api.BaseURL)
	}
	if api.Description != "A test remote" {
		t.Errorf("Description: expected 'A test remote', got %s", api.Description)
	}
	if api.Username != "user" {
		t.Errorf("Username: expected user, got %s", api.Username)
	}
	if api.Password != "pass" {
		t.Errorf("Password: expected pass, got %s", api.Password)
	}
	if api.ImmutableTTL != 86400 {
		t.Errorf("ImmutableTTL: expected 86400, got %d", api.ImmutableTTL)
	}
	if api.MutableTTL != 3600 {
		t.Errorf("MutableTTL: expected 3600, got %d", api.MutableTTL)
	}
	if !api.CheckMutable {
		t.Error("CheckMutable: expected true")
	}
	if len(api.Patterns) != 2 || api.Patterns[0] != "*.tar.gz" {
		t.Errorf("Patterns: expected [*.tar.gz *.whl], got %v", api.Patterns)
	}
	if len(api.Blocklist) != 1 || api.Blocklist[0] != "blocked/*" {
		t.Errorf("Blocklist: expected [blocked/*], got %v", api.Blocklist)
	}
	if len(api.MutablePatterns) != 1 || api.MutablePatterns[0] != "latest" {
		t.Errorf("MutablePatterns: expected [latest], got %v", api.MutablePatterns)
	}
	if len(api.ImmutablePatterns) != 1 || api.ImmutablePatterns[0] != "v*" {
		t.Errorf("ImmutablePatterns: expected [v*], got %v", api.ImmutablePatterns)
	}
	if !api.BanTagsEnabled {
		t.Error("BanTagsEnabled: expected true")
	}
	if len(api.BanTags) != 2 || api.BanTags[0] != "latest" {
		t.Errorf("BanTags: expected [latest dev], got %v", api.BanTags)
	}
	if !api.QuarantineEnabled {
		t.Error("QuarantineEnabled: expected true")
	}
	if api.QuarantineDays != 7 {
		t.Errorf("QuarantineDays: expected 7, got %d", api.QuarantineDays)
	}
	if api.StaleOnError {
		t.Error("StaleOnError: expected false")
	}
	if api.ReleasesRemote != "cdn-remote" {
		t.Errorf("ReleasesRemote: expected cdn-remote, got %s", api.ReleasesRemote)
	}
}

func TestModelToAPI_NullLists(t *testing.T) {
	ctx := context.Background()
	r := &remoteResource{packageType: "generic"}

	model := remoteResourceModel{
		Name:              types.StringValue("minimal"),
		BaseURL:           types.StringValue("https://example.com"),
		Description:       types.StringValue(""),
		Username:          types.StringValue(""),
		Password:          types.StringValue(""),
		ImmutableTTL:      types.Int64Value(0),
		MutableTTL:        types.Int64Value(3600),
		CheckMutable:      types.BoolValue(true),
		Patterns:          types.ListNull(types.StringType),
		Blocklist:         types.ListNull(types.StringType),
		MutablePatterns:   types.ListNull(types.StringType),
		ImmutablePatterns: types.ListNull(types.StringType),
		BanTagsEnabled:    types.BoolValue(false),
		BanTags:           types.ListNull(types.StringType),
		QuarantineEnabled: types.BoolValue(false),
		QuarantineDays:    types.Int64Value(3),
		StaleOnError:      types.BoolValue(true),
		ReleasesRemote:    types.StringValue(""),
	}

	api := r.modelToAPI(ctx, model)

	if api.Patterns != nil {
		t.Errorf("Patterns: expected nil, got %v", api.Patterns)
	}
	if api.Blocklist != nil {
		t.Errorf("Blocklist: expected nil, got %v", api.Blocklist)
	}
	if api.MutablePatterns != nil {
		t.Errorf("MutablePatterns: expected nil, got %v", api.MutablePatterns)
	}
	if api.ImmutablePatterns != nil {
		t.Errorf("ImmutablePatterns: expected nil, got %v", api.ImmutablePatterns)
	}
	if api.BanTags != nil {
		t.Errorf("BanTags: expected nil, got %v", api.BanTags)
	}
}

func TestModelToAPI_PackageTypeFromResource(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		pkgType string
	}{
		{"generic"},
		{"docker"},
		{"helm"},
		{"pypi"},
		{"npm"},
		{"rpm"},
		{"alpine"},
		{"puppet"},
		{"terraform"},
		{"goproxy"},
		{"cargo"},
		{"ansible"},
	}

	for _, tt := range tests {
		t.Run(tt.pkgType, func(t *testing.T) {
			r := &remoteResource{packageType: tt.pkgType}
			model := remoteResourceModel{
				Name:              types.StringValue("test"),
				BaseURL:           types.StringValue("https://example.com"),
				Description:       types.StringValue(""),
				Username:          types.StringValue(""),
				Password:          types.StringValue(""),
				ImmutableTTL:      types.Int64Value(0),
				MutableTTL:        types.Int64Value(0),
				CheckMutable:      types.BoolValue(false),
				Patterns:          types.ListNull(types.StringType),
				Blocklist:         types.ListNull(types.StringType),
				MutablePatterns:   types.ListNull(types.StringType),
				ImmutablePatterns: types.ListNull(types.StringType),
				BanTagsEnabled:    types.BoolValue(false),
				BanTags:           types.ListNull(types.StringType),
				QuarantineEnabled: types.BoolValue(false),
				QuarantineDays:    types.Int64Value(0),
				StaleOnError:      types.BoolValue(false),
				ReleasesRemote:    types.StringValue(""),
			}
			api := r.modelToAPI(ctx, model)
			if api.PackageType != tt.pkgType {
				t.Errorf("expected package_type %s, got %s", tt.pkgType, api.PackageType)
			}
		})
	}
}

func TestAPIToModel_FullFields(t *testing.T) {
	ctx := context.Background()
	r := &remoteResource{packageType: "docker"}

	api := remoteAPI{
		Name:              "my-remote",
		PackageType:       "docker",
		BaseURL:           "https://registry.example.com",
		Description:       "A test remote",
		Username:          "user",
		Password:          "pass",
		ImmutableTTL:      86400,
		MutableTTL:        3600,
		CheckMutable:      true,
		Patterns:          []string{"*.tar.gz"},
		Blocklist:         []string{"blocked/*"},
		MutablePatterns:   []string{"latest"},
		ImmutablePatterns: []string{"v*"},
		BanTagsEnabled:    true,
		BanTags:           []string{"latest"},
		QuarantineEnabled: true,
		QuarantineDays:    7,
		StaleOnError:      false,
		ReleasesRemote:    "cdn-remote",
		ManagedBy:         "terraform",

		UpstreamDialTimeout:           3,
		UpstreamTLSTimeout:            4,
		UpstreamResponseHeaderTimeout: 5,
	}

	model := r.apiToModel(ctx, api)

	if model.UpstreamDialTimeout.ValueInt64() != 3 ||
		model.UpstreamTLSTimeout.ValueInt64() != 4 ||
		model.UpstreamResponseHeaderTimeout.ValueInt64() != 5 {
		t.Errorf("upstream timeouts: got %d/%d/%d, want 3/4/5",
			model.UpstreamDialTimeout.ValueInt64(),
			model.UpstreamTLSTimeout.ValueInt64(),
			model.UpstreamResponseHeaderTimeout.ValueInt64())
	}
	if model.Name.ValueString() != "my-remote" {
		t.Errorf("Name: expected my-remote, got %s", model.Name.ValueString())
	}
	if model.BaseURL.ValueString() != "https://registry.example.com" {
		t.Errorf("BaseURL: expected https://registry.example.com, got %s", model.BaseURL.ValueString())
	}
	if model.Description.ValueString() != "A test remote" {
		t.Errorf("Description: expected 'A test remote', got %s", model.Description.ValueString())
	}
	if model.Username.ValueString() != "user" {
		t.Errorf("Username: expected user, got %s", model.Username.ValueString())
	}
	if model.Password.ValueString() != "pass" {
		t.Errorf("Password: expected pass, got %s", model.Password.ValueString())
	}
	if model.ImmutableTTL.ValueInt64() != 86400 {
		t.Errorf("ImmutableTTL: expected 86400, got %d", model.ImmutableTTL.ValueInt64())
	}
	if model.MutableTTL.ValueInt64() != 3600 {
		t.Errorf("MutableTTL: expected 3600, got %d", model.MutableTTL.ValueInt64())
	}
	if !model.CheckMutable.ValueBool() {
		t.Error("CheckMutable: expected true")
	}
	if !model.QuarantineEnabled.ValueBool() {
		t.Error("QuarantineEnabled: expected true")
	}
	if model.QuarantineDays.ValueInt64() != 7 {
		t.Errorf("QuarantineDays: expected 7, got %d", model.QuarantineDays.ValueInt64())
	}
	if model.StaleOnError.ValueBool() {
		t.Error("StaleOnError: expected false")
	}
	if !model.BanTagsEnabled.ValueBool() {
		t.Error("BanTagsEnabled: expected true")
	}
	if model.ReleasesRemote.ValueString() != "cdn-remote" {
		t.Errorf("ReleasesRemote: expected cdn-remote, got %s", model.ReleasesRemote.ValueString())
	}

	// Verify list fields via round-trip
	patterns := listToStrings(ctx, model.Patterns)
	if len(patterns) != 1 || patterns[0] != "*.tar.gz" {
		t.Errorf("Patterns: expected [*.tar.gz], got %v", patterns)
	}
	blocklist := listToStrings(ctx, model.Blocklist)
	if len(blocklist) != 1 || blocklist[0] != "blocked/*" {
		t.Errorf("Blocklist: expected [blocked/*], got %v", blocklist)
	}
	banTags := listToStrings(ctx, model.BanTags)
	if len(banTags) != 1 || banTags[0] != "latest" {
		t.Errorf("BanTags: expected [latest], got %v", banTags)
	}
}

func TestAPIToModel_EmptyLists(t *testing.T) {
	ctx := context.Background()
	r := &remoteResource{packageType: "generic"}

	api := remoteAPI{
		Name:        "minimal",
		PackageType: "generic",
		BaseURL:     "https://example.com",
		// All lists are nil/empty
	}

	model := r.apiToModel(ctx, api)

	if !model.Patterns.IsNull() {
		t.Errorf("Patterns: expected null for nil input, got %v", model.Patterns)
	}
	if !model.Blocklist.IsNull() {
		t.Errorf("Blocklist: expected null for nil input, got %v", model.Blocklist)
	}
	if !model.MutablePatterns.IsNull() {
		t.Errorf("MutablePatterns: expected null for nil input, got %v", model.MutablePatterns)
	}
	if !model.ImmutablePatterns.IsNull() {
		t.Errorf("ImmutablePatterns: expected null for nil input, got %v", model.ImmutablePatterns)
	}
	if !model.BanTags.IsNull() {
		t.Errorf("BanTags: expected null for nil input, got %v", model.BanTags)
	}
}

func TestModelToAPI_RoundTrip(t *testing.T) {
	ctx := context.Background()
	r := &remoteResource{packageType: "helm"}

	original := remoteAPI{
		Name:              "helm-remote",
		PackageType:       "helm",
		BaseURL:           "https://charts.example.com",
		Description:       "Helm chart mirror",
		Username:          "helmuser",
		Password:          "helmpass",
		ImmutableTTL:      172800,
		MutableTTL:        7200,
		CheckMutable:      true,
		Patterns:          []string{"stable/*", "incubator/*"},
		Blocklist:         []string{"deprecated/*"},
		MutablePatterns:   []string{"latest"},
		ImmutablePatterns: []string{"v1.*"},
		BanTagsEnabled:    false,
		BanTags:           nil,
		QuarantineEnabled: false,
		QuarantineDays:    3,
		StaleOnError:      true,
		ReleasesRemote:    "",

		UpstreamDialTimeout:           5,
		UpstreamTLSTimeout:            0,
		UpstreamResponseHeaderTimeout: 45,
	}

	// API -> Model -> API round-trip
	model := r.apiToModel(ctx, original)
	result := r.modelToAPI(ctx, model)

	if result.UpstreamDialTimeout != original.UpstreamDialTimeout ||
		result.UpstreamTLSTimeout != original.UpstreamTLSTimeout ||
		result.UpstreamResponseHeaderTimeout != original.UpstreamResponseHeaderTimeout {
		t.Errorf("upstream timeouts round-trip: got %d/%d/%d, want %d/%d/%d",
			result.UpstreamDialTimeout, result.UpstreamTLSTimeout, result.UpstreamResponseHeaderTimeout,
			original.UpstreamDialTimeout, original.UpstreamTLSTimeout, original.UpstreamResponseHeaderTimeout)
	}

	if result.Name != original.Name {
		t.Errorf("Name: expected %s, got %s", original.Name, result.Name)
	}
	if result.PackageType != original.PackageType {
		t.Errorf("PackageType: expected %s, got %s", original.PackageType, result.PackageType)
	}
	if result.BaseURL != original.BaseURL {
		t.Errorf("BaseURL: expected %s, got %s", original.BaseURL, result.BaseURL)
	}
	if result.Description != original.Description {
		t.Errorf("Description: expected %s, got %s", original.Description, result.Description)
	}
	if result.ImmutableTTL != original.ImmutableTTL {
		t.Errorf("ImmutableTTL: expected %d, got %d", original.ImmutableTTL, result.ImmutableTTL)
	}
	if result.MutableTTL != original.MutableTTL {
		t.Errorf("MutableTTL: expected %d, got %d", original.MutableTTL, result.MutableTTL)
	}
	if result.CheckMutable != original.CheckMutable {
		t.Errorf("CheckMutable: expected %v, got %v", original.CheckMutable, result.CheckMutable)
	}
	if len(result.Patterns) != len(original.Patterns) {
		t.Errorf("Patterns length: expected %d, got %d", len(original.Patterns), len(result.Patterns))
	}
	for i := range original.Patterns {
		if result.Patterns[i] != original.Patterns[i] {
			t.Errorf("Patterns[%d]: expected %s, got %s", i, original.Patterns[i], result.Patterns[i])
		}
	}
	if result.QuarantineDays != original.QuarantineDays {
		t.Errorf("QuarantineDays: expected %d, got %d", original.QuarantineDays, result.QuarantineDays)
	}
	if result.StaleOnError != original.StaleOnError {
		t.Errorf("StaleOnError: expected %v, got %v", original.StaleOnError, result.StaleOnError)
	}
}

func TestRemoteResource_Metadata(t *testing.T) {
	tests := []struct {
		pkgType  string
		expected string
	}{
		{"generic", "artifactapi_remote_generic"},
		{"docker", "artifactapi_remote_docker"},
		{"helm", "artifactapi_remote_helm"},
		{"pypi", "artifactapi_remote_pypi"},
		{"npm", "artifactapi_remote_npm"},
		{"github_rpm", "artifactapi_remote_github_rpm"},
		{"github_deb", "artifactapi_remote_github_deb"},
		{"github_alpine", "artifactapi_remote_github_alpine"},
	}

	for _, tt := range tests {
		t.Run(tt.pkgType, func(t *testing.T) {
			r := &remoteResource{packageType: tt.pkgType}
			req := resource.MetadataRequest{ProviderTypeName: "artifactapi"}
			var resp resource.MetadataResponse
			r.Metadata(context.Background(), req, &resp)
			if resp.TypeName != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, resp.TypeName)
			}
		})
	}
}

func TestModelToAPI_GitHubRPM(t *testing.T) {
	ctx := context.Background()
	r := &remoteResource{packageType: "github_rpm"}

	model := remoteResourceModel{
		Name:           types.StringValue("acme-tools"),
		BaseURL:        types.StringValue("https://api.github.com/repos/acme/tools"),
		ReleasesRemote: types.StringValue("github"),
		MutableTTL:     types.Int64Value(3600),
		Patterns:       stringsToList(ctx, []string{`.*\.x86_64\.rpm$`}),
	}

	api := r.modelToAPI(ctx, model)

	if api.PackageType != "github_rpm" {
		t.Errorf("PackageType: expected github_rpm, got %s", api.PackageType)
	}
	if api.BaseURL != "https://api.github.com/repos/acme/tools" {
		t.Errorf("BaseURL: got %s", api.BaseURL)
	}
	if api.ReleasesRemote != "github" {
		t.Errorf("ReleasesRemote: expected github, got %s", api.ReleasesRemote)
	}
	if len(api.Patterns) != 1 || api.Patterns[0] != `.*\.x86_64\.rpm$` {
		t.Errorf("Patterns: got %v", api.Patterns)
	}
}

func TestModelToAPI_GitHubDeb(t *testing.T) {
	ctx := context.Background()
	r := &remoteResource{packageType: "github_deb"}

	model := remoteResourceModel{
		Name:           types.StringValue("acme-tools"),
		BaseURL:        types.StringValue("https://api.github.com/repos/acme/tools"),
		ReleasesRemote: types.StringValue("github"),
		MutableTTL:     types.Int64Value(3600),
		Patterns:       stringsToList(ctx, []string{`.*_amd64\.deb$`}),
	}

	api := r.modelToAPI(ctx, model)

	if api.PackageType != "github_deb" {
		t.Errorf("PackageType: expected github_deb, got %s", api.PackageType)
	}
	if api.BaseURL != "https://api.github.com/repos/acme/tools" {
		t.Errorf("BaseURL: got %s", api.BaseURL)
	}
	if api.ReleasesRemote != "github" {
		t.Errorf("ReleasesRemote: expected github, got %s", api.ReleasesRemote)
	}
	if len(api.Patterns) != 1 || api.Patterns[0] != `.*_amd64\.deb$` {
		t.Errorf("Patterns: got %v", api.Patterns)
	}
}

func TestRemoteResource_Schema(t *testing.T) {
	r := &remoteResource{packageType: "docker"}
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	expectedAttrs := []string{
		"name", "base_url", "description", "username", "password",
		"immutable_ttl", "mutable_ttl", "check_mutable",
		"patterns", "blocklist", "mutable_patterns", "immutable_patterns",
		"ban_tags_enabled", "ban_tags",
		"quarantine_enabled", "quarantine_days",
		"stale_on_error", "releases_remote",
		"upstream_dial_timeout", "upstream_tls_timeout", "upstream_response_header_timeout",
	}

	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing expected attribute: %s", attr)
		}
	}
}

func TestNewRemoteResource_Constructors(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() resource.Resource
		expected string
	}{
		{"generic", func() resource.Resource { return NewRemoteGeneric() }, "generic"},
		{"docker", func() resource.Resource { return NewRemoteDocker() }, "docker"},
		{"helm", func() resource.Resource { return NewRemoteHelm() }, "helm"},
		{"pypi", func() resource.Resource { return NewRemotePyPI() }, "pypi"},
		{"npm", func() resource.Resource { return NewRemoteNPM() }, "npm"},
		{"rpm", func() resource.Resource { return NewRemoteRPM() }, "rpm"},
		{"alpine", func() resource.Resource { return NewRemoteAlpine() }, "alpine"},
		{"puppet", func() resource.Resource { return NewRemotePuppet() }, "puppet"},
		{"terraform", func() resource.Resource { return NewRemoteTerraform() }, "terraform"},
		{"goproxy", func() resource.Resource { return NewRemoteGoProxy() }, "goproxy"},
		{"cargo", func() resource.Resource { return NewRemoteCargo() }, "cargo"},
		{"ansible", func() resource.Resource { return NewRemoteAnsible() }, "ansible"},
		{"github_rpm", func() resource.Resource { return NewRemoteGitHubRPM() }, "github_rpm"},
		{"github_deb", func() resource.Resource { return NewRemoteGitHubDeb() }, "github_deb"},
		{"github_alpine", func() resource.Resource { return NewRemoteGitHubAlpine() }, "github_alpine"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.fn()
			rr, ok := r.(*remoteResource)
			if !ok {
				t.Fatal("expected *remoteResource")
			}
			if rr.packageType != tt.expected {
				t.Errorf("expected packageType %s, got %s", tt.expected, rr.packageType)
			}
		})
	}
}

func TestModelToAPI_Mirrorlist(t *testing.T) {
	ctx := context.Background()
	mirrors := []string{"https://mirror-a.example.com/rpm", "https://mirror-b.example.com/rpm"}

	for _, pkgType := range []string{"rpm", "deb", "alpine"} {
		t.Run(pkgType, func(t *testing.T) {
			r := &remoteResource{packageType: pkgType}
			model := remoteResourceModel{
				Name:       types.StringValue("mirror-remote"),
				BaseURL:    types.StringValue("https://primary.example.com"),
				Mirrorlist: stringsToList(ctx, mirrors),
			}
			api := r.modelToAPI(ctx, model)
			if len(api.Mirrorlist) != 2 || api.Mirrorlist[0] != mirrors[0] || api.Mirrorlist[1] != mirrors[1] {
				t.Errorf("Mirrorlist: expected %v, got %v", mirrors, api.Mirrorlist)
			}
		})
	}
}

func TestModelToAPI_MirrorlistNull(t *testing.T) {
	ctx := context.Background()
	r := &remoteResource{packageType: "rpm"}
	model := remoteResourceModel{
		Name:       types.StringValue("no-mirror"),
		BaseURL:    types.StringValue("https://primary.example.com"),
		Mirrorlist: types.ListNull(types.StringType),
	}
	api := r.modelToAPI(ctx, model)
	if api.Mirrorlist != nil {
		t.Errorf("Mirrorlist: expected nil (omitted) for null list, got %v", api.Mirrorlist)
	}
}

func TestAPIToModel_Mirrorlist(t *testing.T) {
	ctx := context.Background()
	r := &remoteResource{packageType: "rpm"}
	mirrors := []string{"https://mirror-a.example.com/rpm", "https://mirror-b.example.com/rpm"}

	// Round-trip: mirrorlist survives API -> Model -> API without churn.
	api := remoteAPI{Name: "rt", PackageType: "rpm", BaseURL: "https://primary.example.com", Mirrorlist: mirrors}
	model := r.apiToModel(ctx, api)
	got := listToStrings(ctx, model.Mirrorlist)
	if len(got) != 2 || got[0] != mirrors[0] || got[1] != mirrors[1] {
		t.Errorf("Mirrorlist round-trip: expected %v, got %v", mirrors, got)
	}

	// nil mirrorlist maps to a null list (not empty) to avoid perpetual diff.
	empty := r.apiToModel(ctx, remoteAPI{Name: "e", PackageType: "rpm", BaseURL: "https://x"})
	if !empty.Mirrorlist.IsNull() {
		t.Errorf("Mirrorlist: expected null for nil input, got %v", empty.Mirrorlist)
	}
}

func TestRemoteResource_SchemaHasMirrorlist(t *testing.T) {
	r := &remoteResource{packageType: "rpm"}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if _, ok := resp.Schema.Attributes["mirrorlist"]; !ok {
		t.Fatal("missing mirrorlist attribute in schema")
	}
}

// mirrorlistConfig builds a tfsdk.Config for the remote schema with every
// attribute null except mirrorlist, which is set to the given values (or null
// when values is nil).
func mirrorlistConfig(ctx context.Context, t *testing.T, r *remoteResource, values []string) tfsdk.Config {
	t.Helper()
	var resp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &resp)
	objType := resp.Schema.Type().TerraformType(ctx).(tftypes.Object)

	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, at := range objType.AttributeTypes {
		attrs[name] = tftypes.NewValue(at, nil)
	}
	if values != nil {
		elems := make([]tftypes.Value, len(values))
		for i, v := range values {
			elems[i] = tftypes.NewValue(tftypes.String, v)
		}
		attrs["mirrorlist"] = tftypes.NewValue(objType.AttributeTypes["mirrorlist"], elems)
	}
	return tfsdk.Config{Schema: resp.Schema, Raw: tftypes.NewValue(objType, attrs)}
}

func TestValidateConfig_Mirrorlist(t *testing.T) {
	ctx := context.Background()
	mirrors := []string{"https://mirror.example.com"}

	// Supported types accept mirrorlist.
	for _, pkgType := range []string{"rpm", "deb", "alpine"} {
		t.Run("allowed/"+pkgType, func(t *testing.T) {
			r := &remoteResource{packageType: pkgType}
			var resp resource.ValidateConfigResponse
			r.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: mirrorlistConfig(ctx, t, r, mirrors)}, &resp)
			if resp.Diagnostics.HasError() {
				t.Errorf("unexpected error for %s: %v", pkgType, resp.Diagnostics.Errors())
			}
		})
	}

	// Unsupported types reject a set mirrorlist.
	for _, pkgType := range []string{"docker", "pypi", "helm", "generic"} {
		t.Run("rejected/"+pkgType, func(t *testing.T) {
			r := &remoteResource{packageType: pkgType}
			var resp resource.ValidateConfigResponse
			r.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: mirrorlistConfig(ctx, t, r, mirrors)}, &resp)
			if !resp.Diagnostics.HasError() {
				t.Errorf("expected error setting mirrorlist on %s remote", pkgType)
			}
		})
	}

	// A null mirrorlist is fine on any type.
	t.Run("null-on-docker", func(t *testing.T) {
		r := &remoteResource{packageType: "docker"}
		var resp resource.ValidateConfigResponse
		r.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: mirrorlistConfig(ctx, t, r, nil)}, &resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("unexpected error for null mirrorlist on docker: %v", resp.Diagnostics.Errors())
		}
	})
}

func TestModelToAPI_MirrorStrategy(t *testing.T) {
	ctx := context.Background()

	for _, pkgType := range []string{"rpm", "deb", "alpine"} {
		t.Run(pkgType, func(t *testing.T) {
			r := &remoteResource{packageType: pkgType}
			model := remoteResourceModel{
				Name:           types.StringValue("strategy-remote"),
				BaseURL:        types.StringValue("https://primary.example.com"),
				MirrorStrategy: types.StringValue("least_conn"),
			}
			api := r.modelToAPI(ctx, model)
			if api.MirrorStrategy != "least_conn" {
				t.Errorf("MirrorStrategy: expected least_conn, got %q", api.MirrorStrategy)
			}
		})
	}
}

func TestModelToAPI_MirrorStrategyNull(t *testing.T) {
	ctx := context.Background()
	r := &remoteResource{packageType: "rpm"}
	model := remoteResourceModel{
		Name:           types.StringValue("no-strategy"),
		BaseURL:        types.StringValue("https://primary.example.com"),
		MirrorStrategy: types.StringNull(),
	}
	api := r.modelToAPI(ctx, model)
	if api.MirrorStrategy != "" {
		t.Errorf("MirrorStrategy: expected \"\" (omitted) for null, got %q", api.MirrorStrategy)
	}
}

func TestAPIToModel_MirrorStrategy(t *testing.T) {
	ctx := context.Background()
	r := &remoteResource{packageType: "rpm"}

	// Round-trip: mirror_strategy survives API -> Model -> API without churn.
	api := remoteAPI{Name: "rt", PackageType: "rpm", BaseURL: "https://primary.example.com", MirrorStrategy: "least_conn"}
	model := r.apiToModel(ctx, api)
	if model.MirrorStrategy.ValueString() != "least_conn" {
		t.Errorf("MirrorStrategy round-trip: expected least_conn, got %q", model.MirrorStrategy.ValueString())
	}
	if got := r.modelToAPI(ctx, model); got.MirrorStrategy != "least_conn" {
		t.Errorf("MirrorStrategy round-trip back to API: expected least_conn, got %q", got.MirrorStrategy)
	}

	// The refresh/GET path: the API stores mirror_strategy NOT NULL DEFAULT
	// 'round_robin' and always returns it, even when the config omitted it.
	// Read must keep round_robin in state (not collapse to null) so that a null
	// config — which the Computed+Default schema resolves to round_robin —
	// produces no post-apply diff. This is the perpetual-churn regression guard.
	getResp := r.apiToModel(ctx, remoteAPI{Name: "g", PackageType: "rpm", BaseURL: "https://x", MirrorStrategy: "round_robin"})
	if getResp.MirrorStrategy.IsNull() {
		t.Fatal("MirrorStrategy: GET returning round_robin must not map to null (would churn every plan)")
	}
	if getResp.MirrorStrategy.ValueString() != "round_robin" {
		t.Errorf("MirrorStrategy: expected round_robin from GET, got %q", getResp.MirrorStrategy.ValueString())
	}

	// A create response that omits the field settles to the server default so a
	// Computed+Default (round_robin) attribute does not raise inconsistent-result.
	empty := r.apiToModel(ctx, remoteAPI{Name: "e", PackageType: "rpm", BaseURL: "https://x"})
	if empty.MirrorStrategy.ValueString() != "round_robin" {
		t.Errorf("MirrorStrategy: expected round_robin default for empty input, got %q", empty.MirrorStrategy.ValueString())
	}
}

func TestRemoteResource_SchemaHasMirrorStrategy(t *testing.T) {
	r := &remoteResource{packageType: "rpm"}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	attr, ok := resp.Schema.Attributes["mirror_strategy"]
	if !ok {
		t.Fatal("missing mirror_strategy attribute in schema")
	}
	// Must be Computed (with a server-default) so an API-defaulted round_robin
	// can live in state while the config is null, without a perpetual diff.
	if !attr.IsComputed() {
		t.Error("mirror_strategy must be Computed to hold the API's round_robin default without churn")
	}
	if !attr.IsOptional() {
		t.Error("mirror_strategy must remain Optional so users can set least_conn")
	}
}

// mirrorStrategyConfig builds a tfsdk.Config for the remote schema with every
// attribute null except mirror_strategy, which is set to the given value (or
// null when value is "").
func mirrorStrategyConfig(ctx context.Context, t *testing.T, r *remoteResource, value string) tfsdk.Config {
	t.Helper()
	var resp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &resp)
	objType := resp.Schema.Type().TerraformType(ctx).(tftypes.Object)

	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, at := range objType.AttributeTypes {
		attrs[name] = tftypes.NewValue(at, nil)
	}
	if value != "" {
		attrs["mirror_strategy"] = tftypes.NewValue(tftypes.String, value)
	}
	return tfsdk.Config{Schema: resp.Schema, Raw: tftypes.NewValue(objType, attrs)}
}

func TestValidateConfig_MirrorStrategy(t *testing.T) {
	ctx := context.Background()

	// Supported types accept a valid mirror_strategy.
	for _, pkgType := range []string{"rpm", "deb", "alpine"} {
		t.Run("allowed/"+pkgType, func(t *testing.T) {
			r := &remoteResource{packageType: pkgType}
			var resp resource.ValidateConfigResponse
			r.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: mirrorStrategyConfig(ctx, t, r, "least_conn")}, &resp)
			if resp.Diagnostics.HasError() {
				t.Errorf("unexpected error for %s: %v", pkgType, resp.Diagnostics.Errors())
			}
		})
	}

	// Unsupported types reject a set mirror_strategy.
	for _, pkgType := range []string{"docker", "pypi", "helm", "generic"} {
		t.Run("rejected/"+pkgType, func(t *testing.T) {
			r := &remoteResource{packageType: pkgType}
			var resp resource.ValidateConfigResponse
			r.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: mirrorStrategyConfig(ctx, t, r, "round_robin")}, &resp)
			if !resp.Diagnostics.HasError() {
				t.Errorf("expected error setting mirror_strategy on %s remote", pkgType)
			}
		})
	}

	// An invalid enum value is rejected even on a supported type.
	t.Run("invalid-value-on-rpm", func(t *testing.T) {
		r := &remoteResource{packageType: "rpm"}
		var resp resource.ValidateConfigResponse
		r.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: mirrorStrategyConfig(ctx, t, r, "bogus")}, &resp)
		if !resp.Diagnostics.HasError() {
			t.Error("expected error for invalid mirror_strategy value on rpm")
		}
	})

	// A null mirror_strategy is fine on any type.
	t.Run("null-on-docker", func(t *testing.T) {
		r := &remoteResource{packageType: "docker"}
		var resp resource.ValidateConfigResponse
		r.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: mirrorStrategyConfig(ctx, t, r, "")}, &resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("unexpected error for null mirror_strategy on docker: %v", resp.Diagnostics.Errors())
		}
	})
}
