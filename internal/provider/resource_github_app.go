package provider

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const githubAppsPath = "/api/v2/github-apps/"

var (
	_ resource.Resource                   = &githubAppResource{}
	_ resource.ResourceWithImportState    = &githubAppResource{}
	_ resource.ResourceWithValidateConfig = &githubAppResource{}
	_ resource.ResourceWithModifyPlan     = &githubAppResource{}
)

type githubAppResource struct {
	client *apiClient
}

type githubAppResourceModel struct {
	Name                  types.String `tfsdk:"name"`
	AppID                 types.Int64  `tfsdk:"app_id"`
	InstallationID        types.Int64  `tfsdk:"installation_id"`
	PrivateKey            types.String `tfsdk:"private_key"`
	APIBaseURL            types.String `tfsdk:"api_base_url"`
	Description           types.String `tfsdk:"description"`
	PrivateKeyFingerprint types.String `tfsdk:"private_key_fingerprint"`
	CreatedAt             types.String `tfsdk:"created_at"`
	UpdatedAt             types.String `tfsdk:"updated_at"`
}

func init() { registerResource(NewGithubAppResource) }

func NewGithubAppResource() resource.Resource {
	return &githubAppResource{}
}

func (r *githubAppResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_github_app"
}

func (r *githubAppResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a GitHub App credential that GitHub-facing remotes reference via `github_app` to authenticate with installation tokens. The server never returns the private key; the configured key is kept in state and compared by fingerprint on refresh, so a key changed outside Terraform is rewritten on the next apply.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Unique name of the GitHub App credential.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_id": schema.Int64Attribute{
				Description: "GitHub App ID; must be positive.",
				Required:    true,
			},
			"installation_id": schema.Int64Attribute{
				Description: "Installation ID of the app on the target account or organisation; must be positive.",
				Required:    true,
			},
			"private_key": schema.StringAttribute{
				Description: "RSA private key of the app in PEM form (PKCS#1 or PKCS#8). Write-only on the server and stored in Terraform state; after import it is unknown, so the next apply rewrites it.",
				Required:    true,
				Sensitive:   true,
			},
			"api_base_url": schema.StringAttribute{
				Description: "Absolute https URL of the GitHub API; the server defaults it to `https://api.github.com`. Set it for GitHub Enterprise Server; removing it from configuration keeps the current value.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description.",
				Optional:    true, Computed: true, Default: stringdefault.StaticString(""),
			},
			"private_key_fingerprint": schema.StringAttribute{
				Description: "`SHA256:` plus the unpadded base64 SHA-256 of the DER PKIX public key.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation timestamp reported by the server.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "Last update timestamp reported by the server.",
				Computed:    true,
			},
		},
	}
}

func (r *githubAppResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *githubAppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan githubAppResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var existing githubAppAPI
	err := r.client.get(ctx, githubAppsPath+plan.Name.ValueString(), &existing)
	if err == nil {
		resp.Diagnostics.AddError("create github app failed",
			fmt.Sprintf("GitHub App %q already exists; import it instead of creating it.", plan.Name.ValueString()))
		return
	}
	if !isNotFound(err) {
		resp.Diagnostics.AddError("create github app failed", err.Error())
		return
	}

	state, ok := r.put(ctx, plan, "create github app failed", &resp.Diagnostics)
	if ok {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *githubAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state githubAppResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var got githubAppAPI
	if err := r.client.get(ctx, githubAppsPath+state.Name.ValueString(), &got); err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read github app failed", err.Error())
		return
	}

	newState := githubAppAPIToModel(got, state.PrivateKey)
	if !privateKeyMatches(newState.PrivateKey, got.PrivateKeyFingerprint) {
		newState.PrivateKey = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *githubAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan githubAppResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state, ok := r.put(ctx, plan, "update github app failed", &resp.Diagnostics)
	if ok {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	}
}

func (r *githubAppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state githubAppResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.del(ctx, githubAppsPath+state.Name.ValueString())
	switch {
	case err == nil, isNotFound(err):
	case isConflict(err):
		resp.Diagnostics.AddError("github app is still referenced by a remote",
			fmt.Sprintf("Remove github_app from every remote using %q before deleting it: %s", state.Name.ValueString(), err))
	default:
		resp.Diagnostics.AddError("delete github app failed", err.Error())
	}
}

func (r *githubAppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *githubAppResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	var key types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("private_key"), &key)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fp := types.StringUnknown()
	if !key.IsNull() && !key.IsUnknown() {
		if s, err := privateKeyFingerprint(key.ValueString()); err == nil {
			fp = types.StringValue(s)
		}
	}
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("private_key_fingerprint"), fp)...)
}

func (r *githubAppResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config githubAppResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for name, v := range map[string]types.Int64{"app_id": config.AppID, "installation_id": config.InstallationID} {
		if knownInt64(v) && v.ValueInt64() <= 0 {
			resp.Diagnostics.AddAttributeError(path.Root(name), "invalid "+name,
				fmt.Sprintf("%s must be positive, got %d.", name, v.ValueInt64()))
		}
	}

	if !config.PrivateKey.IsNull() && !config.PrivateKey.IsUnknown() {
		if _, err := parseRSAPrivateKey(config.PrivateKey.ValueString()); err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("private_key"), "invalid private_key", err.Error())
		}
	}

	if !config.APIBaseURL.IsNull() && !config.APIBaseURL.IsUnknown() {
		if err := validateHTTPSURL(config.APIBaseURL.ValueString()); err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("api_base_url"), "invalid api_base_url", err.Error())
		}
	}
}

func (r *githubAppResource) put(ctx context.Context, plan githubAppResourceModel, summary string, diags *diag.Diagnostics) (githubAppResourceModel, bool) {
	body := githubAppAPI{
		AppID:          plan.AppID.ValueInt64(),
		InstallationID: plan.InstallationID.ValueInt64(),
		PrivateKey:     plan.PrivateKey.ValueString(),
		Description:    plan.Description.ValueString(),
	}
	if !plan.APIBaseURL.IsUnknown() {
		body.APIBaseURL = plan.APIBaseURL.ValueString()
	}
	var got githubAppAPI
	if err := r.client.put(ctx, githubAppsPath+plan.Name.ValueString(), body, &got); err != nil {
		diags.AddError(summary, err.Error())
		return githubAppResourceModel{}, false
	}
	return githubAppAPIToModel(got, plan.PrivateKey), true
}

func githubAppAPIToModel(api githubAppAPI, privateKey types.String) githubAppResourceModel {
	return githubAppResourceModel{
		Name:                  types.StringValue(api.Name),
		AppID:                 types.Int64Value(api.AppID),
		InstallationID:        types.Int64Value(api.InstallationID),
		PrivateKey:            privateKey,
		APIBaseURL:            types.StringValue(api.APIBaseURL),
		Description:           types.StringValue(api.Description),
		PrivateKeyFingerprint: types.StringValue(api.PrivateKeyFingerprint),
		CreatedAt:             types.StringValue(api.CreatedAt),
		UpdatedAt:             types.StringValue(api.UpdatedAt),
	}
}

// privateKeyMatches reports whether the key in state still belongs to the server's fingerprint.
func privateKeyMatches(key types.String, serverFingerprint string) bool {
	if key.IsNull() || key.IsUnknown() {
		return true
	}
	if serverFingerprint == "" {
		return true
	}
	fp, err := privateKeyFingerprint(key.ValueString())
	return err == nil && fp == serverFingerprint
}

func parseRSAPrivateKey(s string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(s))
	if block == nil {
		return nil, errors.New("private_key must be a PEM-encoded RSA private key")
	}
	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return k, nil
	}
	k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("private_key is not a PKCS#1 or PKCS#8 private key: %w", err)
	}
	rk, ok := k.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private_key must be an RSA key, got %T", k)
	}
	return rk, nil
}

func privateKeyFingerprint(s string) (string, error) {
	k, err := parseRSAPrivateKey(s)
	if err != nil {
		return "", err
	}
	der, err := x509.MarshalPKIXPublicKey(&k.PublicKey)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(der)
	return "SHA256:" + base64.RawStdEncoding.EncodeToString(sum[:]), nil
}

func validateHTTPSURL(s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("api_base_url is not a valid URL: %w", err)
	}
	if u.Scheme != "https" || u.Host == "" {
		return fmt.Errorf("api_base_url must be an absolute https URL, got %q", s)
	}
	return nil
}
