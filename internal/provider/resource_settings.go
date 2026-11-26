package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	settingsPath = "/api/v2/settings"
	settingsID   = "settings"
)

var (
	_ resource.Resource                   = &settingsResource{}
	_ resource.ResourceWithImportState    = &settingsResource{}
	_ resource.ResourceWithValidateConfig = &settingsResource{}
)

type settingsResource struct {
	client *apiClient
}

type settingsResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	S3HealthUpThreshold   types.Int64  `tfsdk:"s3_health_up_threshold"`
	S3HealthDownThreshold types.Int64  `tfsdk:"s3_health_down_threshold"`
	S3HealthIntervalMs    types.Int64  `tfsdk:"s3_health_interval_ms"`
	S3HealthTimeoutMs     types.Int64  `tfsdk:"s3_health_timeout_ms"`
}

func init() { registerResource(NewSettingsResource) }

func NewSettingsResource() resource.Resource {
	return &settingsResource{}
}

func (r *settingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_settings"
}

func settingsIntAttribute(description string) schema.Int64Attribute {
	return schema.Int64Attribute{
		Description: description + " Must be at least 1. Unset reverts to the server's environment default.",
		Optional:    true,
		Computed:    true,
	}
}

func (r *settingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the singleton ArtifactAPI global settings. Only one instance should exist per server; destroying it reverts every setting to the server's environment defaults.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Fixed identifier of the singleton, always \"settings\".",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"s3_health_up_threshold":   settingsIntAttribute("Consecutive successful S3 health probes before the backend is marked up."),
			"s3_health_down_threshold": settingsIntAttribute("Consecutive failed S3 health probes before the backend is marked down."),
			"s3_health_interval_ms":    settingsIntAttribute("Interval between S3 health probes, in milliseconds."),
			"s3_health_timeout_ms":     settingsIntAttribute("Timeout of each S3 health probe, in milliseconds; must be shorter than s3_health_interval_ms."),
		},
	}
}

func (r *settingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *settingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.write(ctx, req.Config, &resp.State, &resp.Diagnostics, "create settings failed")
}

func (r *settingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var got settingsAPI
	if err := r.client.get(ctx, settingsPath, &got); err != nil {
		resp.Diagnostics.AddError("read settings failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, settingsAPIToModel(got))...)
}

func (r *settingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.write(ctx, req.Config, &resp.State, &resp.Diagnostics, "update settings failed")
}

func (r *settingsResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	if err := r.client.del(ctx, settingsPath); err != nil {
		resp.Diagnostics.AddError("delete settings failed", err.Error())
	}
}

func (r *settingsResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), settingsID)...)
}

func (r *settingsResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config settingsResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for name, v := range map[string]types.Int64{
		"s3_health_up_threshold":   config.S3HealthUpThreshold,
		"s3_health_down_threshold": config.S3HealthDownThreshold,
		"s3_health_interval_ms":    config.S3HealthIntervalMs,
		"s3_health_timeout_ms":     config.S3HealthTimeoutMs,
	} {
		if knownInt64(v) && v.ValueInt64() < 1 {
			resp.Diagnostics.AddAttributeError(
				path.Root(name),
				"invalid "+name,
				fmt.Sprintf("%s must be at least 1, got %d.", name, v.ValueInt64()),
			)
		}
	}

	if knownInt64(config.S3HealthIntervalMs) && knownInt64(config.S3HealthTimeoutMs) &&
		config.S3HealthTimeoutMs.ValueInt64() >= config.S3HealthIntervalMs.ValueInt64() {
		resp.Diagnostics.AddAttributeError(
			path.Root("s3_health_timeout_ms"),
			"invalid s3_health_timeout_ms",
			fmt.Sprintf("s3_health_timeout_ms (%d) must be shorter than s3_health_interval_ms (%d).",
				config.S3HealthTimeoutMs.ValueInt64(), config.S3HealthIntervalMs.ValueInt64()),
		)
	}
}

// write PUTs the configured values, not the plan, so unset fields are sent null and revert to the env default.
func (r *settingsResource) write(ctx context.Context, cfg tfsdk.Config, state *tfsdk.State, diags *diag.Diagnostics, summary string) {
	var config settingsResourceModel
	diags.Append(cfg.Get(ctx, &config)...)
	if diags.HasError() {
		return
	}

	var effective settingsAPI
	if err := r.client.put(ctx, settingsPath, settingsModelToAPI(config), &effective); err != nil {
		diags.AddError(summary, err.Error())
		return
	}
	diags.Append(state.Set(ctx, settingsAPIToModel(effective))...)
}

func knownInt64(v types.Int64) bool {
	return !v.IsNull() && !v.IsUnknown()
}

func int64Ptr(v types.Int64) *int64 {
	if !knownInt64(v) {
		return nil
	}
	return v.ValueInt64Pointer()
}

func settingsModelToAPI(m settingsResourceModel) settingsAPI {
	return settingsAPI{
		S3HealthUpThreshold:   int64Ptr(m.S3HealthUpThreshold),
		S3HealthDownThreshold: int64Ptr(m.S3HealthDownThreshold),
		S3HealthIntervalMs:    int64Ptr(m.S3HealthIntervalMs),
		S3HealthTimeoutMs:     int64Ptr(m.S3HealthTimeoutMs),
	}
}

func settingsAPIToModel(api settingsAPI) settingsResourceModel {
	return settingsResourceModel{
		ID:                    types.StringValue(settingsID),
		S3HealthUpThreshold:   types.Int64PointerValue(api.S3HealthUpThreshold),
		S3HealthDownThreshold: types.Int64PointerValue(api.S3HealthDownThreshold),
		S3HealthIntervalMs:    types.Int64PointerValue(api.S3HealthIntervalMs),
		S3HealthTimeoutMs:     types.Int64PointerValue(api.S3HealthTimeoutMs),
	}
}
