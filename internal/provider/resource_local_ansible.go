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
	_ resource.Resource                = &localAnsibleResource{}
	_ resource.ResourceWithImportState = &localAnsibleResource{}
)

type localAnsibleResource struct {
	client *apiClient
}

type localAnsibleResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func init() { registerResource(NewLocalAnsibleResource) }

func NewLocalAnsibleResource() resource.Resource {
	return &localAnsibleResource{}
}

func (r *localAnsibleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_local_ansible"
}

func (r *localAnsibleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a local ArtifactAPI Ansible Galaxy repository for hosting collections directly.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Unique name of the local ansible repository.",
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

func (r *localAnsibleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *localAnsibleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan localAnsibleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	api := localAnsibleModelToAPI(plan)
	api.ManagedBy = "terraform"

	var created remoteAPI
	if err := r.client.post(ctx, "/api/v2/remotes", api, &created); err != nil {
		resp.Diagnostics.AddError("create local ansible failed", err.Error())
		return
	}

	state := localAnsibleAPIToModel(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *localAnsibleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state localAnsibleResourceModel
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
		resp.Diagnostics.AddError("read local ansible failed", err.Error())
		return
	}

	newState := localAnsibleAPIToModel(remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *localAnsibleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan localAnsibleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	api := localAnsibleModelToAPI(plan)
	api.ManagedBy = "terraform"

	var updated remoteAPI
	if err := r.client.put(ctx, "/api/v2/remotes/"+plan.Name.ValueString(), api, &updated); err != nil {
		resp.Diagnostics.AddError("update local ansible failed", err.Error())
		return
	}

	state := localAnsibleAPIToModel(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *localAnsibleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state localAnsibleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.del(ctx, "/api/v2/remotes/"+state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("delete local ansible failed", err.Error())
		return
	}
}

func (r *localAnsibleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func localAnsibleModelToAPI(m localAnsibleResourceModel) remoteAPI {
	return remoteAPI{
		Name:        m.Name.ValueString(),
		PackageType: "ansible",
		RepoType:    "local",
		Description: m.Description.ValueString(),
	}
}

func localAnsibleAPIToModel(api remoteAPI) localAnsibleResourceModel {
	return localAnsibleResourceModel{
		Name:        types.StringValue(api.Name),
		Description: types.StringValue(api.Description),
	}
}
