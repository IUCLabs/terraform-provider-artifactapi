package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &remoteResource{}
	_ resource.ResourceWithImportState = &remoteResource{}
)

type remoteResource struct {
	client      *apiClient
	packageType string
}

type remoteResourceModel struct {
	Name              types.String `tfsdk:"name"`
	BaseURL           types.String `tfsdk:"base_url"`
	Description       types.String `tfsdk:"description"`
	Username          types.String `tfsdk:"username"`
	Password          types.String `tfsdk:"password"`
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
	ReleasesRemote    types.String `tfsdk:"releases_remote"`

	UpstreamDialTimeout           types.Int64 `tfsdk:"upstream_dial_timeout"`
	UpstreamTLSTimeout            types.Int64 `tfsdk:"upstream_tls_timeout"`
	UpstreamResponseHeaderTimeout types.Int64 `tfsdk:"upstream_response_header_timeout"`
}

func newRemoteResource(packageType string) func() resource.Resource {
	return func() resource.Resource {
		return &remoteResource{packageType: packageType}
	}
}

func NewRemoteGeneric() resource.Resource   { return &remoteResource{packageType: "generic"} }
func NewRemoteDocker() resource.Resource    { return &remoteResource{packageType: "docker"} }
func NewRemoteHelm() resource.Resource      { return &remoteResource{packageType: "helm"} }
func NewRemotePyPI() resource.Resource      { return &remoteResource{packageType: "pypi"} }
func NewRemoteNPM() resource.Resource       { return &remoteResource{packageType: "npm"} }
func NewRemoteRPM() resource.Resource       { return &remoteResource{packageType: "rpm"} }
func NewRemoteDeb() resource.Resource       { return &remoteResource{packageType: "deb"} }
func NewRemoteAlpine() resource.Resource    { return &remoteResource{packageType: "alpine"} }
func NewRemotePuppet() resource.Resource    { return &remoteResource{packageType: "puppet"} }
func NewRemoteTerraform() resource.Resource { return &remoteResource{packageType: "terraform"} }
func NewRemoteGoProxy() resource.Resource   { return &remoteResource{packageType: "goproxy"} }
func NewRemoteGitHubRPM() resource.Resource { return &remoteResource{packageType: "github_rpm"} }
func NewRemoteGitHubDeb() resource.Resource { return &remoteResource{packageType: "github_deb"} }

func (r *remoteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_remote_" + r.packageType
}

func (r *remoteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Description: "Unique name of the remote.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"base_url": schema.StringAttribute{
			Description: "Upstream repository base URL.",
			Required:    true,
		},
		"description": schema.StringAttribute{
			Optional: true, Computed: true, Default: stringdefault.StaticString(""),
		},
		"username": schema.StringAttribute{
			Optional: true, Computed: true, Sensitive: true, Default: stringdefault.StaticString(""),
		},
		"password": schema.StringAttribute{
			Optional: true, Computed: true, Sensitive: true, Default: stringdefault.StaticString(""),
		},
		"immutable_ttl": schema.Int64Attribute{
			Description: "TTL in seconds for immutable artifacts (0 = forever).",
			Optional:    true, Computed: true, Default: int64default.StaticInt64(0),
		},
		"mutable_ttl": schema.Int64Attribute{
			Description: "TTL in seconds for mutable artifacts.",
			Optional:    true, Computed: true, Default: int64default.StaticInt64(3600),
		},
		"check_mutable": schema.BoolAttribute{
			Description: "Enable conditional revalidation for mutable artifacts.",
			Optional:    true, Computed: true, Default: booldefault.StaticBool(true),
		},
		"patterns": schema.ListAttribute{
			Description: "Paths to proxy (empty = all). This is the allowlist.",
			Optional:    true,
			ElementType: types.StringType,
		},
		"blocklist": schema.ListAttribute{
			Description: "Paths to always deny (checked before patterns).",
			Optional:    true,
			ElementType: types.StringType,
		},
		"mutable_patterns": schema.ListAttribute{
			Description: "Override: paths that should be mutable even if the provider would classify as immutable.",
			Optional:    true,
			ElementType: types.StringType,
		},
		"immutable_patterns": schema.ListAttribute{
			Description: "Override: paths that should be immutable even if the provider would classify as mutable.",
			Optional:    true,
			ElementType: types.StringType,
		},
		"quarantine_enabled": schema.BoolAttribute{
			Optional: true, Computed: true, Default: booldefault.StaticBool(false),
		},
		"quarantine_days": schema.Int64Attribute{
			Optional: true, Computed: true, Default: int64default.StaticInt64(3),
		},
		"stale_on_error": schema.BoolAttribute{
			Description: "Serve stale cached content when upstream is unreachable.",
			Optional:    true, Computed: true, Default: booldefault.StaticBool(true),
		},
		"ban_tags_enabled": schema.BoolAttribute{
			Description: "Enable tag banning (docker only).",
			Optional:    true, Computed: true, Default: booldefault.StaticBool(false),
		},
		"ban_tags": schema.ListAttribute{
			Description: "Tags to ban (docker only).",
			Optional:    true,
			ElementType: types.StringType,
		},
		"releases_remote": schema.StringAttribute{
			Description: "Name of a backend remote that serves the actual download bytes. Used by the terraform remote (download URL rewriting) and the github_rpm/github_deb remotes (302 redirect of .rpm/.deb downloads to a generic github.com remote).",
			Optional:    true, Computed: true, Default: stringdefault.StaticString(""),
		},
		"upstream_dial_timeout": schema.Int64Attribute{
			Description: "Upstream TCP connect timeout in seconds (0 = server default).",
			Optional:    true, Computed: true, Default: int64default.StaticInt64(0),
		},
		"upstream_tls_timeout": schema.Int64Attribute{
			Description: "Upstream TLS handshake timeout in seconds (0 = server default).",
			Optional:    true, Computed: true, Default: int64default.StaticInt64(0),
		},
		"upstream_response_header_timeout": schema.Int64Attribute{
			Description: "Upstream response-header timeout in seconds (0 = server default).",
			Optional:    true, Computed: true, Default: int64default.StaticInt64(0),
		},
	}

	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("Manages an ArtifactAPI %s remote.", r.packageType),
		Attributes:  attrs,
	}
}

func (r *remoteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiClient)
	if !ok {
		resp.Diagnostics.AddError("unexpected provider data type", fmt.Sprintf("got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *remoteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan remoteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	api := r.modelToAPI(ctx, plan)
	api.ManagedBy = "terraform"

	var created remoteAPI
	if err := r.client.post(ctx, "/api/v2/remotes", api, &created); err != nil {
		resp.Diagnostics.AddError("create remote failed", err.Error())
		return
	}

	state := r.apiToModel(ctx, created)
	reconcileOptionalLists(&plan, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *remoteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state remoteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var remote remoteAPI
	err := r.client.get(ctx, "/api/v2/remotes/"+state.Name.ValueString(), &remote)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read remote failed", err.Error())
		return
	}

	newState := r.apiToModel(ctx, remote)
	reconcileOptionalLists(&state, &newState)
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *remoteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan remoteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	api := r.modelToAPI(ctx, plan)
	api.ManagedBy = "terraform"

	var updated remoteAPI
	if err := r.client.put(ctx, "/api/v2/remotes/"+plan.Name.ValueString(), api, &updated); err != nil {
		resp.Diagnostics.AddError("update remote failed", err.Error())
		return
	}

	state := r.apiToModel(ctx, updated)
	reconcileOptionalLists(&plan, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *remoteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state remoteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.del(ctx, "/api/v2/remotes/"+state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("delete remote failed", err.Error())
		return
	}
}

func (r *remoteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func reconcileOptionalLists(prior, current *remoteResourceModel) {
	current.Patterns = preserveListNullEmptySemantics(prior.Patterns, current.Patterns)
	current.Blocklist = preserveListNullEmptySemantics(prior.Blocklist, current.Blocklist)
	current.MutablePatterns = preserveListNullEmptySemantics(prior.MutablePatterns, current.MutablePatterns)
	current.ImmutablePatterns = preserveListNullEmptySemantics(prior.ImmutablePatterns, current.ImmutablePatterns)
	current.BanTags = preserveListNullEmptySemantics(prior.BanTags, current.BanTags)
}

func (r *remoteResource) modelToAPI(ctx context.Context, m remoteResourceModel) remoteAPI {
	api := remoteAPI{
		Name:              m.Name.ValueString(),
		PackageType:       r.packageType,
		BaseURL:           m.BaseURL.ValueString(),
		Description:       m.Description.ValueString(),
		Username:          m.Username.ValueString(),
		Password:          m.Password.ValueString(),
		ImmutableTTL:      m.ImmutableTTL.ValueInt64(),
		MutableTTL:        m.MutableTTL.ValueInt64(),
		CheckMutable:      m.CheckMutable.ValueBool(),
		QuarantineEnabled: m.QuarantineEnabled.ValueBool(),
		QuarantineDays:    m.QuarantineDays.ValueInt64(),
		StaleOnError:      m.StaleOnError.ValueBool(),

		UpstreamDialTimeout:           m.UpstreamDialTimeout.ValueInt64(),
		UpstreamTLSTimeout:            m.UpstreamTLSTimeout.ValueInt64(),
		UpstreamResponseHeaderTimeout: m.UpstreamResponseHeaderTimeout.ValueInt64(),
	}
	api.Patterns = listToStrings(ctx, m.Patterns)
	api.Blocklist = listToStrings(ctx, m.Blocklist)
	api.MutablePatterns = listToStrings(ctx, m.MutablePatterns)
	api.ImmutablePatterns = listToStrings(ctx, m.ImmutablePatterns)
	api.BanTagsEnabled = m.BanTagsEnabled.ValueBool()
	api.BanTags = listToStrings(ctx, m.BanTags)
	api.ReleasesRemote = m.ReleasesRemote.ValueString()
	return api
}

func (r *remoteResource) apiToModel(ctx context.Context, api remoteAPI) remoteResourceModel {
	m := remoteResourceModel{
		Name:              types.StringValue(api.Name),
		BaseURL:           types.StringValue(api.BaseURL),
		Description:       types.StringValue(api.Description),
		Username:          types.StringValue(api.Username),
		Password:          types.StringValue(api.Password),
		ImmutableTTL:      types.Int64Value(api.ImmutableTTL),
		MutableTTL:        types.Int64Value(api.MutableTTL),
		CheckMutable:      types.BoolValue(api.CheckMutable),
		Patterns:          stringsToList(ctx, api.Patterns),
		Blocklist:         stringsToList(ctx, api.Blocklist),
		MutablePatterns:   stringsToList(ctx, api.MutablePatterns),
		ImmutablePatterns: stringsToList(ctx, api.ImmutablePatterns),
		QuarantineEnabled: types.BoolValue(api.QuarantineEnabled),
		QuarantineDays:    types.Int64Value(api.QuarantineDays),
		StaleOnError:      types.BoolValue(api.StaleOnError),
		BanTagsEnabled:    types.BoolValue(api.BanTagsEnabled),
		BanTags:           stringsToList(ctx, api.BanTags),
		ReleasesRemote:    types.StringValue(api.ReleasesRemote),

		UpstreamDialTimeout:           types.Int64Value(api.UpstreamDialTimeout),
		UpstreamTLSTimeout:            types.Int64Value(api.UpstreamTLSTimeout),
		UpstreamResponseHeaderTimeout: types.Int64Value(api.UpstreamResponseHeaderTimeout),
	}
	return m
}
