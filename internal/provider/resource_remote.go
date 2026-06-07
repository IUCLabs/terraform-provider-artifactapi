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
	client *apiClient
}

type remoteResourceModel struct {
	Name              types.String `tfsdk:"name"`
	PackageType       types.String `tfsdk:"package_type"`
	BaseURL           types.String `tfsdk:"base_url"`
	Description       types.String `tfsdk:"description"`
	Username          types.String `tfsdk:"username"`
	Password          types.String `tfsdk:"password"`
	ImmutableTTL      types.Int64  `tfsdk:"immutable_ttl"`
	MutableTTL        types.Int64  `tfsdk:"mutable_ttl"`
	CheckMutable      types.Bool   `tfsdk:"check_mutable"`
	ImmutablePatterns types.List   `tfsdk:"immutable_patterns"`
	MutablePatterns   types.List   `tfsdk:"mutable_patterns"`
	Allowlist         types.List   `tfsdk:"allowlist"`
	Blocklist         types.List   `tfsdk:"blocklist"`
	BanTagsEnabled    types.Bool   `tfsdk:"ban_tags_enabled"`
	BanTags           types.List   `tfsdk:"ban_tags"`
	QuarantineEnabled types.Bool   `tfsdk:"quarantine_enabled"`
	QuarantineDays    types.Int64  `tfsdk:"quarantine_days"`
	StaleOnError      types.Bool   `tfsdk:"stale_on_error"`
	ReleasesRemote    types.String `tfsdk:"releases_remote"`
}

func NewRemoteResource() resource.Resource {
	return &remoteResource{}
}

func (r *remoteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_remote"
}

func (r *remoteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an ArtifactAPI remote proxy repository.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Unique name of the remote.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"package_type": schema.StringAttribute{
				Description: "Package type: generic, docker, helm, pypi, npm, rpm, alpine, puppet, terraform, goproxy.",
				Required:    true,
			},
			"base_url": schema.StringAttribute{
				Description: "Upstream repository base URL.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"username": schema.StringAttribute{
				Description: "Username for upstream authentication.",
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Default:     stringdefault.StaticString(""),
			},
			"password": schema.StringAttribute{
				Description: "Password for upstream authentication.",
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Default:     stringdefault.StaticString(""),
			},
			"immutable_ttl": schema.Int64Attribute{
				Description: "TTL in seconds for immutable artifacts (0 = forever).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"mutable_ttl": schema.Int64Attribute{
				Description: "TTL in seconds for mutable artifacts.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3600),
			},
			"check_mutable": schema.BoolAttribute{
				Description: "Enable conditional revalidation (ETag/If-None-Match) for mutable artifacts.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"immutable_patterns": schema.ListAttribute{
				Description: "Regex patterns that identify immutable artifacts.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"mutable_patterns": schema.ListAttribute{
				Description: "Additional regex patterns for mutable artifacts (merged with provider built-ins).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"allowlist": schema.ListAttribute{
				Description: "If non-empty, only paths matching these patterns are proxied. Empty = allow all.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"blocklist": schema.ListAttribute{
				Description: "Paths matching these patterns are always denied (checked before allowlist).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"ban_tags_enabled": schema.BoolAttribute{
				Description: "Enable tag banning (Docker only).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"ban_tags": schema.ListAttribute{
				Description: "Tags to ban (Docker only).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"quarantine_enabled": schema.BoolAttribute{
				Description: "Enable quarantine for newly published artifacts.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"quarantine_days": schema.Int64Attribute{
				Description: "Number of days to quarantine new artifacts.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3),
			},
			"stale_on_error": schema.BoolAttribute{
				Description: "Serve stale cached content when upstream is unreachable.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"releases_remote": schema.StringAttribute{
				Description: "Name of the CDN remote for download URL rewriting (terraform package type).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
		},
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

	api := modelToAPI(ctx, plan)
	api.ManagedBy = "terraform"

	var created remoteAPI
	if err := r.client.post(ctx, "/api/v2/remotes", api, &created); err != nil {
		resp.Diagnostics.AddError("create remote failed", err.Error())
		return
	}

	state := apiToModel(ctx, created)
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

	newState := apiToModel(ctx, remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *remoteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan remoteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	api := modelToAPI(ctx, plan)
	api.ManagedBy = "terraform"

	var updated remoteAPI
	if err := r.client.put(ctx, "/api/v2/remotes/"+plan.Name.ValueString(), api, &updated); err != nil {
		resp.Diagnostics.AddError("update remote failed", err.Error())
		return
	}

	state := apiToModel(ctx, updated)
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

func modelToAPI(ctx context.Context, m remoteResourceModel) remoteAPI {
	api := remoteAPI{
		Name:              m.Name.ValueString(),
		PackageType:       m.PackageType.ValueString(),
		BaseURL:           m.BaseURL.ValueString(),
		Description:       m.Description.ValueString(),
		Username:          m.Username.ValueString(),
		Password:          m.Password.ValueString(),
		ImmutableTTL:      m.ImmutableTTL.ValueInt64(),
		MutableTTL:        m.MutableTTL.ValueInt64(),
		CheckMutable:      m.CheckMutable.ValueBool(),
		BanTagsEnabled:    m.BanTagsEnabled.ValueBool(),
		QuarantineEnabled: m.QuarantineEnabled.ValueBool(),
		QuarantineDays:    m.QuarantineDays.ValueInt64(),
		StaleOnError:      m.StaleOnError.ValueBool(),
		ReleasesRemote:    m.ReleasesRemote.ValueString(),
	}
	api.ImmutablePatterns = listToStrings(ctx, m.ImmutablePatterns)
	api.MutablePatterns = listToStrings(ctx, m.MutablePatterns)
	api.Allowlist = listToStrings(ctx, m.Allowlist)
	api.Blocklist = listToStrings(ctx, m.Blocklist)
	api.BanTags = listToStrings(ctx, m.BanTags)
	return api
}

func apiToModel(ctx context.Context, api remoteAPI) remoteResourceModel {
	return remoteResourceModel{
		Name:              types.StringValue(api.Name),
		PackageType:       types.StringValue(api.PackageType),
		BaseURL:           types.StringValue(api.BaseURL),
		Description:       types.StringValue(api.Description),
		Username:          types.StringValue(api.Username),
		Password:          types.StringValue(api.Password),
		ImmutableTTL:      types.Int64Value(api.ImmutableTTL),
		MutableTTL:        types.Int64Value(api.MutableTTL),
		CheckMutable:      types.BoolValue(api.CheckMutable),
		ImmutablePatterns: stringsToList(ctx, api.ImmutablePatterns),
		MutablePatterns:   stringsToList(ctx, api.MutablePatterns),
		Allowlist:         stringsToList(ctx, api.Allowlist),
		Blocklist:         stringsToList(ctx, api.Blocklist),
		BanTagsEnabled:    types.BoolValue(api.BanTagsEnabled),
		BanTags:           stringsToList(ctx, api.BanTags),
		QuarantineEnabled: types.BoolValue(api.QuarantineEnabled),
		QuarantineDays:    types.Int64Value(api.QuarantineDays),
		StaleOnError:      types.BoolValue(api.StaleOnError),
		ReleasesRemote:    types.StringValue(api.ReleasesRemote),
	}
}

func listToStrings(ctx context.Context, l types.List) []string {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	var result []string
	l.ElementsAs(ctx, &result, false)
	return result
}

func stringsToList(ctx context.Context, ss []string) types.List {
	if ss == nil {
		ss = []string{}
	}
	elems := make([]types.String, len(ss))
	for i, s := range ss {
		elems[i] = types.StringValue(s)
	}
	list, _ := types.ListValueFrom(ctx, types.StringType, elems)
	return list
}
