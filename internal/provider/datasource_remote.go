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
			"name":               schema.StringAttribute{Required: true, Description: "Remote name."},
			"package_type":       schema.StringAttribute{Computed: true},
			"base_url":           schema.StringAttribute{Computed: true},
			"description":        schema.StringAttribute{Computed: true},
			"immutable_ttl":      schema.Int64Attribute{Computed: true},
			"mutable_ttl":        schema.Int64Attribute{Computed: true},
			"check_mutable":      schema.BoolAttribute{Computed: true},
			"immutable_patterns": schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"mutable_patterns":   schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"allowlist":          schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"blocklist":          schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"ban_tags_enabled":   schema.BoolAttribute{Computed: true},
			"ban_tags":           schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"quarantine_enabled": schema.BoolAttribute{Computed: true},
			"quarantine_days":    schema.Int64Attribute{Computed: true},
			"stale_on_error":     schema.BoolAttribute{Computed: true},
			"releases_remote":    schema.StringAttribute{Computed: true},
			"managed_by":         schema.StringAttribute{Computed: true},
		},
	}
}

type remoteDataSourceModel struct {
	Name              types.String `tfsdk:"name"`
	PackageType       types.String `tfsdk:"package_type"`
	BaseURL           types.String `tfsdk:"base_url"`
	Description       types.String `tfsdk:"description"`
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
	ManagedBy         types.String `tfsdk:"managed_by"`
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
		Description:       types.StringValue(remote.Description),
		ImmutableTTL:      types.Int64Value(remote.ImmutableTTL),
		MutableTTL:        types.Int64Value(remote.MutableTTL),
		CheckMutable:      types.BoolValue(remote.CheckMutable),
		ImmutablePatterns: stringsToList(ctx, remote.ImmutablePatterns),
		MutablePatterns:   stringsToList(ctx, remote.MutablePatterns),
		Allowlist:         stringsToList(ctx, remote.Allowlist),
		Blocklist:         stringsToList(ctx, remote.Blocklist),
		BanTagsEnabled:    types.BoolValue(remote.BanTagsEnabled),
		BanTags:           stringsToList(ctx, remote.BanTags),
		QuarantineEnabled: types.BoolValue(remote.QuarantineEnabled),
		QuarantineDays:    types.Int64Value(remote.QuarantineDays),
		StaleOnError:      types.BoolValue(remote.StaleOnError),
		ReleasesRemote:    types.StringValue(remote.ReleasesRemote),
		ManagedBy:         types.StringValue(remote.ManagedBy),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
