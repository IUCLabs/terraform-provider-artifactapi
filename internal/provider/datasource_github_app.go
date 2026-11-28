package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &githubAppDataSource{}

type githubAppDataSource struct {
	client *apiClient
}

type githubAppDataSourceModel struct {
	Name                  types.String `tfsdk:"name"`
	AppID                 types.Int64  `tfsdk:"app_id"`
	InstallationID        types.Int64  `tfsdk:"installation_id"`
	APIBaseURL            types.String `tfsdk:"api_base_url"`
	Description           types.String `tfsdk:"description"`
	PrivateKeyFingerprint types.String `tfsdk:"private_key_fingerprint"`
	CreatedAt             types.String `tfsdk:"created_at"`
	UpdatedAt             types.String `tfsdk:"updated_at"`
}

func init() { registerDataSource(NewGithubAppDataSource) }

func NewGithubAppDataSource() datasource.DataSource {
	return &githubAppDataSource{}
}

func (d *githubAppDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_github_app"
}

func (d *githubAppDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Read an existing GitHub App credential without managing it. GitHub Apps created in the ArtifactAPI UI are referenced this way: pass `name` to a remote's `github_app`. The private key stays on the server and is never exposed.",
		Attributes: map[string]schema.Attribute{
			"name":                    schema.StringAttribute{Required: true, Description: "Name of the GitHub App credential to read."},
			"app_id":                  schema.Int64Attribute{Computed: true, Description: "GitHub App ID."},
			"installation_id":         schema.Int64Attribute{Computed: true, Description: "Installation ID of the app; null while the app is not yet installed."},
			"api_base_url":            schema.StringAttribute{Computed: true, Description: "GitHub API base URL the app authenticates against."},
			"description":             schema.StringAttribute{Computed: true, Description: "Human-readable description."},
			"private_key_fingerprint": schema.StringAttribute{Computed: true, Description: "`SHA256:` plus the unpadded base64 SHA-256 of the DER PKIX public key."},
			"created_at":              schema.StringAttribute{Computed: true, Description: "Creation timestamp reported by the server."},
			"updated_at":              schema.StringAttribute{Computed: true, Description: "Last update timestamp reported by the server."},
		},
	}
}

func (d *githubAppDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *githubAppDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config githubAppDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var got githubAppAPI
	if err := d.client.get(ctx, githubAppsPath+config.Name.ValueString(), &got); err != nil {
		if isNotFound(err) {
			resp.Diagnostics.AddError("github app not found", fmt.Sprintf("No GitHub App named %q exists.", config.Name.ValueString()))
			return
		}
		resp.Diagnostics.AddError("read github app failed", err.Error())
		return
	}

	installationID := types.Int64Null()
	if got.InstallationID != 0 {
		installationID = types.Int64Value(got.InstallationID)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, githubAppDataSourceModel{
		Name:                  config.Name,
		AppID:                 types.Int64Value(got.AppID),
		InstallationID:        installationID,
		APIBaseURL:            types.StringValue(got.APIBaseURL),
		Description:           types.StringValue(got.Description),
		PrivateKeyFingerprint: types.StringValue(got.PrivateKeyFingerprint),
		CreatedAt:             types.StringValue(got.CreatedAt),
		UpdatedAt:             types.StringValue(got.UpdatedAt),
	})...)
}
