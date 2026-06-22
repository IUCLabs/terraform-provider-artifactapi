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
	_ resource.Resource                = &localTerraformResource{}
	_ resource.ResourceWithImportState = &localTerraformResource{}
)

type localTerraformResource struct {
	client *apiClient
}

type localTerraformResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func NewLocalTerraformResource() resource.Resource {
	return &localTerraformResource{}
}

func (r *localTerraformResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_local_terraform"
}

func (r *localTerraformResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a local ArtifactAPI terraform registry for hosting providers directly.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Unique name of the local terraform repository.",
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

func (r *localTerraformResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *localTerraformResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan localTerraformResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	api := localTerraformModelToAPI(plan)
	api.ManagedBy = "terraform"

	var created remoteAPI
	if err := r.client.post(ctx, "/api/v2/remotes", api, &created); err != nil {
		resp.Diagnostics.AddError("create local terraform failed", err.Error())
		return
	}

	state := localTerraformAPIToModel(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *localTerraformResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state localTerraformResourceModel
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
		resp.Diagnostics.AddError("read local terraform failed", err.Error())
		return
	}

	newState := localTerraformAPIToModel(remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *localTerraformResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan localTerraformResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	api := localTerraformModelToAPI(plan)
	api.ManagedBy = "terraform"

	var updated remoteAPI
	if err := r.client.put(ctx, "/api/v2/remotes/"+plan.Name.ValueString(), api, &updated); err != nil {
		resp.Diagnostics.AddError("update local terraform failed", err.Error())
		return
	}

	state := localTerraformAPIToModel(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *localTerraformResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state localTerraformResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.del(ctx, "/api/v2/remotes/"+state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("delete local terraform failed", err.Error())
		return
	}
}

func (r *localTerraformResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func localTerraformModelToAPI(m localTerraformResourceModel) remoteAPI {
	return remoteAPI{
		Name:        m.Name.ValueString(),
		PackageType: "terraform",
		RepoType:    "local",
		Description: m.Description.ValueString(),
	}
}

func localTerraformAPIToModel(api remoteAPI) localTerraformResourceModel {
	return localTerraformResourceModel{
		Name:        types.StringValue(api.Name),
		Description: types.StringValue(api.Description),
	}
}
