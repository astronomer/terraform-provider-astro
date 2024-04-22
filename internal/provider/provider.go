package provider

import (
	"context"
	"os"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients/iam"
	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/datasources"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/models"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/resources"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure AstronomerProvider satisfies various provider interfaces.
var _ provider.Provider = &AstronomerProvider{}
var _ provider.ProviderWithFunctions = &AstronomerProvider{}

// AstronomerProvider defines the provider implementation.
type AstronomerProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *AstronomerProvider) Metadata(
	ctx context.Context,
	req provider.MetadataRequest,
	resp *provider.MetadataResponse,
) {
	resp.TypeName = "astronomer"
	resp.Version = p.version
}

func (p *AstronomerProvider) Schema(
	ctx context.Context,
	req provider.SchemaRequest,
	resp *provider.SchemaResponse,
) {
	resp.Schema = providerSchema()
}

func providerSchema() schema.Schema {
	return schema.Schema{
		Attributes: schemas.ProviderSchemaAttributes(),
	}
}

func (p *AstronomerProvider) Configure(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	tflog.Info(ctx, "Configuring Astronomer Terraform Provider client")

	var data models.AstronomerProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Will use Token provided in the configuration, or fallback to the ASTRO_API_TOKEN env var
	if data.Token.IsNull() {
		data.Token = types.StringValue(os.Getenv("ASTRO_API_TOKEN"))
	}

	if len(data.Token.ValueString()) == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Astro API Token",
			"Astro API Token must be set in the configuration or the 'ASTRO_API_TOKEN' environment variable",
		)
		return
	}

	if data.Host.IsNull() {
		data.Host = types.StringValue("https://api.astronomer.io")
	}

	platformClient, err := platform.NewPlatformClient(
		data.Host.ValueString(),
		data.Token.ValueString(),
		p.version,
	)
	if err != nil {
		tflog.Error(ctx, "failed to create platform client", map[string]any{"error": err})
		resp.Diagnostics.AddError(
			"Failed to create platform client",
			"failed to create platform API client",
		)
		return
	}
	iamClient, err := iam.NewIamClient(data.Host.ValueString(), data.Token.ValueString(), p.version)
	if err != nil {
		tflog.Error(ctx, "failed to create iam client", map[string]any{"error": err})
		resp.Diagnostics.AddError("Failed to create iam client", "failed to create IAM API client")
		return
	}

	apiClientsModel := models.ApiClientsModel{
		OrganizationId: data.OrganizationId.ValueString(),
		PlatformClient: platformClient,
		IamClient:      iamClient,
	}

	// Example client configuration for data sources and resources
	resp.DataSourceData = apiClientsModel
	resp.ResourceData = apiClientsModel
}

func (p *AstronomerProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewWorkspaceResource,
		resources.NewDeploymentResource,
		resources.NewClusterResource,
	}
}

func (p *AstronomerProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewWorkspaceDataSource,
		datasources.NewWorkspacesDataSource,
		datasources.NewDeploymentDataSource,
		datasources.NewDeploymentsDataSource,
		datasources.NewOrganizationDataSource,
		datasources.NewClusterDataSource,
		datasources.NewClustersDataSource,
		datasources.NewClusterOptionsDataSource,
		datasources.NewDeploymentOptionsDataSource,
	}
}

func (p *AstronomerProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AstronomerProvider{
			version: version,
		}
	}
}
