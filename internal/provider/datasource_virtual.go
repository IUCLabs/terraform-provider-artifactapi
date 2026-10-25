package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &virtualDataSource{}

type virtualDataSource struct {
	client *apiClient
}

func init() { registerDataSource(NewVirtualDataSource) }

func NewVirtualDataSource() datasource.DataSource {
	return &virtualDataSource{}
}

func (d *virtualDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual"
}

func (d *virtualDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Read an existing ArtifactAPI virtual repository.",
		Attributes: map[string]schema.Attribute{
			"name":           schema.StringAttribute{Required: true, Description: "Virtual repository name."},
			"package_type":   schema.StringAttribute{Computed: true},
			"description":    schema.StringAttribute{Computed: true},
			"members":        schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"managed_by":     schema.StringAttribute{Computed: true},
			"parallel_fetch": schema.BoolAttribute{Computed: true},
		},
	}
}

type virtualDataSourceModel struct {
	Name          types.String `tfsdk:"name"`
	PackageType   types.String `tfsdk:"package_type"`
	Description   types.String `tfsdk:"description"`
	Members       types.List   `tfsdk:"members"`
	ManagedBy     types.String `tfsdk:"managed_by"`
	ParallelFetch types.Bool   `tfsdk:"parallel_fetch"`
}

func (d *virtualDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *virtualDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config virtualDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var virt virtualAPI
	if err := d.client.get(ctx, "/api/v2/virtuals/"+config.Name.ValueString(), &virt); err != nil {
		resp.Diagnostics.AddError("read virtual failed", err.Error())
		return
	}

	state := virtualDataSourceModel{
		Name:          types.StringValue(virt.Name),
		PackageType:   types.StringValue(virt.PackageType),
		Description:   types.StringValue(virt.Description),
		Members:       stringsToList(ctx, virt.Members),
		ManagedBy:     types.StringValue(virt.ManagedBy),
		ParallelFetch: types.BoolValue(virt.ParallelFetch),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
