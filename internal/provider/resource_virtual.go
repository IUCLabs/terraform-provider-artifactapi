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
	_ resource.Resource                = &virtualResource{}
	_ resource.ResourceWithImportState = &virtualResource{}
)

type virtualResource struct {
	client *apiClient
}

type virtualResourceModel struct {
	Name        types.String `tfsdk:"name"`
	PackageType types.String `tfsdk:"package_type"`
	Description types.String `tfsdk:"description"`
	Members     types.List   `tfsdk:"members"`
}

func init() { registerResource(NewVirtualResource) }

func NewVirtualResource() resource.Resource {
	return &virtualResource{}
}

func (r *virtualResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual"
}

func (r *virtualResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an ArtifactAPI virtual repository that merges multiple remotes.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Unique name of the virtual repository.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"package_type": schema.StringAttribute{
				Description: "Package type (must match member remotes): helm, pypi.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"members": schema.ListAttribute{
				Description: "Ordered list of member remote names. Earlier members have higher priority for duplicate entries.",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *virtualResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *virtualResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan virtualResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	api := virtualModelToAPI(ctx, plan)
	api.ManagedBy = "terraform"

	var created virtualAPI
	if err := r.client.post(ctx, "/api/v2/virtuals", api, &created); err != nil {
		resp.Diagnostics.AddError("create virtual failed", err.Error())
		return
	}

	state := virtualAPIToModel(ctx, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *virtualResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state virtualResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var virt virtualAPI
	err := r.client.get(ctx, "/api/v2/virtuals/"+state.Name.ValueString(), &virt)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read virtual failed", err.Error())
		return
	}

	newState := virtualAPIToModel(ctx, virt)
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *virtualResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan virtualResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	api := virtualModelToAPI(ctx, plan)
	api.ManagedBy = "terraform"

	var updated virtualAPI
	if err := r.client.put(ctx, "/api/v2/virtuals/"+plan.Name.ValueString(), api, &updated); err != nil {
		resp.Diagnostics.AddError("update virtual failed", err.Error())
		return
	}

	state := virtualAPIToModel(ctx, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *virtualResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state virtualResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.del(ctx, "/api/v2/virtuals/"+state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("delete virtual failed", err.Error())
		return
	}
}

func (r *virtualResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func virtualModelToAPI(ctx context.Context, m virtualResourceModel) virtualAPI {
	return virtualAPI{
		Name:        m.Name.ValueString(),
		PackageType: m.PackageType.ValueString(),
		Description: m.Description.ValueString(),
		Members:     listToStrings(ctx, m.Members),
	}
}

func virtualAPIToModel(ctx context.Context, api virtualAPI) virtualResourceModel {
	return virtualResourceModel{
		Name:        types.StringValue(api.Name),
		PackageType: types.StringValue(api.PackageType),
		Description: types.StringValue(api.Description),
		Members:     stringsToList(ctx, api.Members),
	}
}
