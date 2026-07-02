package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
