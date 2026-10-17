package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &localGemsResource{}
	_ resource.ResourceWithImportState = &localGemsResource{}
)

type localGemsResource struct {
	client *apiClient
}

type localGemsResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func init() { registerResource(NewLocalGemsResource) }

func NewLocalGemsResource() resource.Resource {
	return &localGemsResource{}
}

func (r *localGemsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_local_gems"
}

func (r *localGemsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a local ArtifactAPI RubyGems repository for hosting gems directly.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Unique name of the local gems repository.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
		},
	}
}

func (r *localGemsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *localGemsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan localGemsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	api := localGemsModelToAPI(plan)
	api.ManagedBy = "terraform"

	var created remoteAPI
	if err := r.client.post(ctx, "/api/v2/remotes", api, &created); err != nil {
		resp.Diagnostics.AddError("create local gems failed", err.Error())
		return
	}

	state := localGemsAPIToModel(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *localGemsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state localGemsResourceModel
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
		resp.Diagnostics.AddError("read local gems failed", err.Error())
		return
	}

	newState := localGemsAPIToModel(remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *localGemsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan localGemsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	api := localGemsModelToAPI(plan)
	api.ManagedBy = "terraform"

	var updated remoteAPI
	if err := r.client.put(ctx, "/api/v2/remotes/"+plan.Name.ValueString(), api, &updated); err != nil {
		resp.Diagnostics.AddError("update local gems failed", err.Error())
		return
	}

	state := localGemsAPIToModel(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *localGemsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state localGemsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.del(ctx, "/api/v2/remotes/"+state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("delete local gems failed", err.Error())
		return
	}
}

func (r *localGemsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func localGemsModelToAPI(m localGemsResourceModel) remoteAPI {
	return remoteAPI{
		Name:        m.Name.ValueString(),
		PackageType: "gems",
		RepoType:    "local",
		Description: m.Description.ValueString(),
	}
}

func localGemsAPIToModel(api remoteAPI) localGemsResourceModel {
	return localGemsResourceModel{
		Name:        types.StringValue(api.Name),
		Description: types.StringValue(api.Description),
	}
}
