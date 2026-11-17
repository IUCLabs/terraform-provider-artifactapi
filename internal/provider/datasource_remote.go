package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &remoteDataSource{}

type remoteDataSource struct {
	client *apiClient
}

func init() { registerDataSource(NewRemoteDataSource) }

func NewRemoteDataSource() datasource.DataSource {
	return &remoteDataSource{}
}

func (d *remoteDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_remote"
}

func (d *remoteDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Read an existing ArtifactAPI remote.",
		Attributes: map[string]schema.Attribute{
			"name":               schema.StringAttribute{Required: true, Description: "Name of the remote to read."},
			"package_type":       schema.StringAttribute{Computed: true, Description: "Package type of the repository (e.g. docker, helm, rpm)."},
			"base_url":           schema.StringAttribute{Computed: true, Description: "Upstream repository base URL."},
			"mirrorlist":         schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Extra upstream mirror base URLs (rpm/deb/alpine only)."},
			"mirror_strategy":    schema.StringAttribute{Computed: true, Description: "Mirror load-balancing strategy: round_robin or least_conn."},
			"description":        schema.StringAttribute{Computed: true, Description: "Human-readable description."},
			"immutable_ttl":      schema.Int64Attribute{Computed: true, Description: "TTL in seconds for immutable artifacts (0 = forever)."},
			"mutable_ttl":        schema.Int64Attribute{Computed: true, Description: "TTL in seconds for mutable artifacts."},
			"check_mutable":      schema.BoolAttribute{Computed: true, Description: "Whether mutable artifacts are conditionally revalidated."},
			"patterns":           schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Allowlist of paths to proxy (empty = all)."},
			"blocklist":          schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Paths that are always denied."},
			"mutable_patterns":   schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Paths forced to be treated as mutable."},
			"immutable_patterns": schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Paths forced to be treated as immutable."},
			"ban_tags_enabled":   schema.BoolAttribute{Computed: true, Description: "Whether tag banning is enabled (docker only)."},
			"ban_tags":           schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Banned tags (docker only)."},
			"quarantine_enabled": schema.BoolAttribute{Computed: true, Description: "Whether newly published upstream artifacts are quarantined."},
			"quarantine_days":    schema.Int64Attribute{Computed: true, Description: "Days new artifacts are quarantined."},
			"stale_on_error":     schema.BoolAttribute{Computed: true, Description: "Whether stale cached content is served when upstream is unreachable."},
			"serve_stale":        schema.BoolAttribute{Computed: true, Description: "Whether rpm/deb/alpine virtuals serve this member's last good index while it refreshes."},
			"cache_repodata":     schema.BoolAttribute{Computed: true, Description: "Whether generated repodata is cached (github_rpm only)."},
			"sync_interval":      schema.Int64Attribute{Computed: true, Description: "Release re-scan interval in seconds (github_rpm/github_deb/github_alpine only)."},
			"releases_remote":    schema.StringAttribute{Computed: true, Description: "Name of the backend remote that serves download bytes."},
			"managed_by":         schema.StringAttribute{Computed: true, Description: "What manages the repository (e.g. terraform)."},

			"upstream_dial_timeout":            schema.Int64Attribute{Computed: true, Description: "Upstream TCP connect timeout in seconds (0 = server default)."},
			"upstream_tls_timeout":             schema.Int64Attribute{Computed: true, Description: "Upstream TLS handshake timeout in seconds (0 = server default)."},
			"upstream_response_header_timeout": schema.Int64Attribute{Computed: true, Description: "Upstream response-header timeout in seconds (0 = server default)."},
		},
	}
}

type remoteDataSourceModel struct {
	Name              types.String `tfsdk:"name"`
	PackageType       types.String `tfsdk:"package_type"`
	BaseURL           types.String `tfsdk:"base_url"`
	Mirrorlist        types.List   `tfsdk:"mirrorlist"`
	MirrorStrategy    types.String `tfsdk:"mirror_strategy"`
	Description       types.String `tfsdk:"description"`
	ImmutableTTL      types.Int64  `tfsdk:"immutable_ttl"`
	MutableTTL        types.Int64  `tfsdk:"mutable_ttl"`
	CheckMutable      types.Bool   `tfsdk:"check_mutable"`
	Patterns          types.List   `tfsdk:"patterns"`
	Blocklist         types.List   `tfsdk:"blocklist"`
	MutablePatterns   types.List   `tfsdk:"mutable_patterns"`
	ImmutablePatterns types.List   `tfsdk:"immutable_patterns"`
	BanTagsEnabled    types.Bool   `tfsdk:"ban_tags_enabled"`
	BanTags           types.List   `tfsdk:"ban_tags"`
	QuarantineEnabled types.Bool   `tfsdk:"quarantine_enabled"`
	QuarantineDays    types.Int64  `tfsdk:"quarantine_days"`
	StaleOnError      types.Bool   `tfsdk:"stale_on_error"`
	ServeStale        types.Bool   `tfsdk:"serve_stale"`
	CacheRepodata     types.Bool   `tfsdk:"cache_repodata"`
	SyncInterval      types.Int64  `tfsdk:"sync_interval"`
	ReleasesRemote    types.String `tfsdk:"releases_remote"`
	ManagedBy         types.String `tfsdk:"managed_by"`

	UpstreamDialTimeout           types.Int64 `tfsdk:"upstream_dial_timeout"`
	UpstreamTLSTimeout            types.Int64 `tfsdk:"upstream_tls_timeout"`
	UpstreamResponseHeaderTimeout types.Int64 `tfsdk:"upstream_response_header_timeout"`
}

func (d *remoteDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiClient)
	if !ok {
		resp.Diagnostics.AddError("unexpected provider data type", fmt.Sprintf("got %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *remoteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config remoteDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var remote remoteAPI
	if err := d.client.get(ctx, "/api/v2/remotes/"+config.Name.ValueString(), &remote); err != nil {
		resp.Diagnostics.AddError("read remote failed", err.Error())
		return
	}

	state := remoteDataSourceModel{
		Name:              types.StringValue(remote.Name),
		PackageType:       types.StringValue(remote.PackageType),
		BaseURL:           types.StringValue(remote.BaseURL),
		Mirrorlist:        stringsToList(ctx, remote.Mirrorlist),
		MirrorStrategy:    types.StringValue(remote.MirrorStrategy),
		Description:       types.StringValue(remote.Description),
		ImmutableTTL:      types.Int64Value(remote.ImmutableTTL),
		MutableTTL:        types.Int64Value(remote.MutableTTL),
		CheckMutable:      types.BoolValue(remote.CheckMutable),
		Patterns:          stringsToList(ctx, remote.Patterns),
		Blocklist:         stringsToList(ctx, remote.Blocklist),
		MutablePatterns:   stringsToList(ctx, remote.MutablePatterns),
		ImmutablePatterns: stringsToList(ctx, remote.ImmutablePatterns),
		BanTagsEnabled:    types.BoolValue(remote.BanTagsEnabled),
		BanTags:           stringsToList(ctx, remote.BanTags),
		QuarantineEnabled: types.BoolValue(remote.QuarantineEnabled),
		QuarantineDays:    types.Int64Value(remote.QuarantineDays),
		StaleOnError:      types.BoolValue(remote.StaleOnError),
		ServeStale:        types.BoolValue(remote.ServeStale),
		CacheRepodata:     types.BoolValue(remote.CacheRepodata),
		SyncInterval:      types.Int64Value(remote.SyncInterval),
		ReleasesRemote:    types.StringValue(remote.ReleasesRemote),
		ManagedBy:         types.StringValue(remote.ManagedBy),

		UpstreamDialTimeout:           types.Int64Value(remote.UpstreamDialTimeout),
		UpstreamTLSTimeout:            types.Int64Value(remote.UpstreamTLSTimeout),
		UpstreamResponseHeaderTimeout: types.Int64Value(remote.UpstreamResponseHeaderTimeout),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
