package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &ArtifactAPIProvider{}

type ArtifactAPIProvider struct {
	version string
}

type artifactAPIProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ArtifactAPIProvider{version: version}
	}
}

func (p *ArtifactAPIProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "artifactapi"
	resp.Version = p.version
}

func (p *ArtifactAPIProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage ArtifactAPI remotes and virtual repositories.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "The ArtifactAPI server endpoint URL (e.g. https://artifactapi.example.com).",
				Required:    true,
			},
			"token": schema.StringAttribute{
				Description: "Bearer token sent as the Authorization header. Falls back to the ARTIFACTAPI_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *ArtifactAPIProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config artifactAPIProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := newAPIClient(config.Endpoint.ValueString())
	client.token = config.Token.ValueString()
	if client.token == "" {
		client.token = os.Getenv("ARTIFACTAPI_TOKEN")
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ArtifactAPIProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newRemoteResource("generic"),
		newRemoteResource("docker"),
		newRemoteResource("helm"),
		newRemoteResource("pypi"),
		newRemoteResource("npm"),
		newRemoteResource("rpm"),
		newRemoteResource("deb"),
		newRemoteResource("alpine"),
		newRemoteResource("puppet"),
		newRemoteResource("terraform"),
		newRemoteResource("goproxy"),
		newRemoteResource("cargo"),
		newRemoteResource("ansible"),
		newRemoteResource("gems"),
		newRemoteResource("github_rpm"),
		newRemoteResource("github_deb"),
		newRemoteResource("github_alpine"),
		NewVirtualResource,
		NewLocalTerraformResource,
		NewLocalPyPIResource,
		NewLocalNPMResource,
		NewLocalRPMResource,
		NewLocalDebResource,
		NewLocalAlpineResource,
		NewLocalDockerResource,
		NewLocalGenericResource,
	}
}

func (p *ArtifactAPIProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewRemoteDataSource,
		NewVirtualDataSource,
	}
}
