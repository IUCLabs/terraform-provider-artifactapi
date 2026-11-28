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
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const githubAppsPath = "/api/v2/github-apps/"

var githubAppNameRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

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
	PrivateKeyWO          types.String `tfsdk:"private_key_wo"`
	PrivateKeyWOVersion   types.Int64  `tfsdk:"private_key_wo_version"`
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
		Description: "Manages a GitHub App credential that GitHub-facing remotes reference via `github_app` to authenticate with installation tokens. Set the key with `private_key_wo` (Terraform 1.11+, never stored in plan or state) or `private_key` (stored in state). The server never returns the key; its fingerprint is compared with the configured key on every plan, so a key changed outside Terraform is rewritten on the next apply.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Unique name of the GitHub App credential: 1-128 characters of letters, digits, `.`, `_` or `-`, starting with a letter or digit.",
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
				Description: "RSA private key of the app in PEM form (PKCS#1 or PKCS#8), stored in Terraform state. Prefer `private_key_wo` on Terraform 1.11+. Exactly one of `private_key` or `private_key_wo` is required.",
				Optional:    true,
				Sensitive:   true,
			},
			"private_key_wo": schema.StringAttribute{
				Description: "Write-only RSA private key of the app in PEM form (PKCS#1 or PKCS#8); requires Terraform 1.11+ and is never stored in plan or state. Recommended over `private_key`. The key is resent whenever its fingerprint differs from the server's or `private_key_wo_version` changes.",
				Optional:    true,
				Sensitive:   true,
				WriteOnly:   true,
			},
			"private_key_wo_version": schema.Int64Attribute{
				Description: "Change this value to force `private_key_wo` to be resent; only valid with `private_key_wo`.",
				Optional:    true,
			},
			"api_base_url": schema.StringAttribute{
				Description: "Absolute https URL of the GitHub API without a trailing slash; the server defaults it to `https://api.github.com`. Set it for GitHub Enterprise Server; removing it from configuration keeps the current value.",
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

	key := configuredPrivateKey(ctx, plan, req.Config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	state, ok := r.put(ctx, plan, key, "create github app failed", &resp.Diagnostics)
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

	newState := githubAppAPIToModel(state, got)
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
	key := configuredPrivateKey(ctx, plan, req.Config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	state, ok := r.put(ctx, plan, key, "update github app failed", &resp.Diagnostics)
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
	var config githubAppResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	key := config.PrivateKey
	if key.IsNull() {
		key = config.PrivateKeyWO
	}
	fp := types.StringUnknown()
	if !key.IsNull() && !key.IsUnknown() {
		if s, err := privateKeyFingerprint(key.ValueString()); err == nil {
			fp = types.StringValue(s)
		}
	}
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("private_key_fingerprint"), fp)...)
	if req.State.Raw.IsNull() {
		return
	}
	var stateFP types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("private_key_fingerprint"), &stateFP)...)
	// A write-only key leaves no other diff, so a fingerprint change alone must still plan an update.
	if !fp.Equal(stateFP) {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("updated_at"), types.StringUnknown())...)
	}
}

func (r *githubAppResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config githubAppResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Name.IsNull() && !config.Name.IsUnknown() && !githubAppNameRE.MatchString(config.Name.ValueString()) {
		resp.Diagnostics.AddAttributeError(path.Root("name"), "invalid name",
			fmt.Sprintf("name must be 1-128 characters of letters, digits, '.', '_' or '-' and start with a letter or digit, got %q.", config.Name.ValueString()))
	}

	for name, v := range map[string]types.Int64{"app_id": config.AppID, "installation_id": config.InstallationID} {
		if knownInt64(v) && v.ValueInt64() <= 0 {
			resp.Diagnostics.AddAttributeError(path.Root(name), "invalid "+name,
				fmt.Sprintf("%s must be positive, got %d.", name, v.ValueInt64()))
		}
	}

	for name, v := range map[string]types.String{"private_key": config.PrivateKey, "private_key_wo": config.PrivateKeyWO} {
		if !v.IsNull() && !v.IsUnknown() {
			if _, err := parseRSAPrivateKey(v.ValueString()); err != nil {
				resp.Diagnostics.AddAttributeError(path.Root(name), "invalid "+name, name+" "+err.Error())
			}
		}
	}

	if !config.PrivateKey.IsUnknown() && !config.PrivateKeyWO.IsUnknown() && config.PrivateKey.IsNull() == config.PrivateKeyWO.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root("private_key_wo"), "invalid private key configuration",
			"Exactly one of private_key_wo or private_key must be set; prefer private_key_wo on Terraform 1.11+.")
	}

	if !config.PrivateKeyWOVersion.IsNull() && config.PrivateKeyWO.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root("private_key_wo_version"), "invalid private_key_wo_version",
			"private_key_wo_version only applies to private_key_wo; remove it or set private_key_wo.")
	}

	if !config.APIBaseURL.IsNull() && !config.APIBaseURL.IsUnknown() {
		if err := validateHTTPSURL(config.APIBaseURL.ValueString()); err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("api_base_url"), "invalid api_base_url", err.Error())
		}
	}
}

// configuredPrivateKey returns private_key from the plan, or private_key_wo from config since write-only values are null in plans.
func configuredPrivateKey(ctx context.Context, plan githubAppResourceModel, config tfsdk.Config, diags *diag.Diagnostics) string {
	if !plan.PrivateKey.IsNull() {
		return plan.PrivateKey.ValueString()
	}
	var wo types.String
	diags.Append(config.GetAttribute(ctx, path.Root("private_key_wo"), &wo)...)
	if diags.HasError() {
		return ""
	}
	if wo.IsNull() || wo.IsUnknown() {
		diags.AddAttributeError(path.Root("private_key_wo"), "missing private key", "Set private_key_wo or private_key.")
		return ""
	}
	if _, err := parseRSAPrivateKey(wo.ValueString()); err != nil {
		diags.AddAttributeError(path.Root("private_key_wo"), "invalid private_key_wo", "private_key_wo "+err.Error())
		return ""
	}
	return wo.ValueString()
}

func (r *githubAppResource) put(ctx context.Context, plan githubAppResourceModel, key, summary string, diags *diag.Diagnostics) (githubAppResourceModel, bool) {
	body := githubAppAPI{
		AppID:          plan.AppID.ValueInt64(),
		InstallationID: plan.InstallationID.ValueInt64(),
		PrivateKey:     key,
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
	return githubAppAPIToModel(plan, got), true
}

// githubAppAPIToModel takes name, private_key and private_key_wo_version from base, the rest from the server.
func githubAppAPIToModel(base githubAppResourceModel, api githubAppAPI) githubAppResourceModel {
	return githubAppResourceModel{
		Name:                  base.Name,
		AppID:                 types.Int64Value(api.AppID),
		InstallationID:        types.Int64Value(api.InstallationID),
		PrivateKey:            base.PrivateKey,
		PrivateKeyWO:          types.StringNull(),
		PrivateKeyWOVersion:   base.PrivateKeyWOVersion,
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
		return nil, errors.New("must be a PEM-encoded RSA private key")
	}
	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return k, nil
	}
	k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("is not a PKCS#1 or PKCS#8 private key: %w", err)
	}
	rk, ok := k.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("must be an RSA key, got %T", k)
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
	if strings.HasSuffix(s, "/") {
		return fmt.Errorf("api_base_url must not end with a slash; remove the trailing slash from %q", s)
	}
	return nil
}
