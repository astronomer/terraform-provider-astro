package datasources

import (
	"context"
	"fmt"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &deploymentOptionsDataSource{}
var _ datasource.DataSourceWithConfigure = &deploymentOptionsDataSource{}

func NewDeploymentOptionsDataSource() datasource.DataSource {
	return &deploymentOptionsDataSource{}
}

// deploymentOptionsDataSource defines the data source implementation.
type deploymentOptionsDataSource struct {
	PlatformClient platform.ClientWithResponsesInterface
	OrganizationId string
}

func (d *deploymentOptionsDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_deployment_options"
}

func (d *deploymentOptionsDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Deployment options data source",
		Attributes:          schemas.DeploymentOptionsDataSourceSchemaAttributes(),
	}
}

func (d *deploymentOptionsDataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	apiClients, ok := req.ProviderData.(models.ApiClientsModel)
	if !ok {
		utils.DataSourceApiClientConfigureError(ctx, req, resp)
		return
	}

	d.PlatformClient = apiClients.PlatformClient
	d.OrganizationId = apiClients.OrganizationId
}

func (d *deploymentOptionsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data models.DeploymentOptions

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := platform.GetDeploymentOptionsParams{}

	deploymentIdParam := data.DeploymentId.ValueString()
	if len(deploymentIdParam) > 0 {
		params.DeploymentId = &deploymentIdParam
	}
	deploymentTypeParam := data.DeploymentType.ValueString()
	if len(deploymentTypeParam) > 0 {
		params.DeploymentType = (*platform.GetDeploymentOptionsParamsDeploymentType)(&deploymentTypeParam)
	}
	executorParam := data.Executor.ValueString()
	if len(executorParam) > 0 {
		params.Executor = (*platform.GetDeploymentOptionsParamsExecutor)(&executorParam)
	}
	cloudProviderParam := data.CloudProvider.ValueString()
	if len(cloudProviderParam) > 0 {
		params.CloudProvider = (*platform.GetDeploymentOptionsParamsCloudProvider)(&cloudProviderParam)
	}

	options, err := d.PlatformClient.GetDeploymentOptionsWithResponse(
		ctx,
		d.OrganizationId,
		&params,
	)
	if err != nil {
		tflog.Error(ctx, "failed to get deployment options", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read deployment options, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, options.HTTPResponse, options.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}
	if options.JSON200 == nil {
		tflog.Error(ctx, "failed to get deployment options", map[string]interface{}{"error": "nil response"})
		resp.Diagnostics.AddError("Client Error", "Unable to read deployment options, got nil response")
		return
	}

	// Populate the model with the response data
	diags := data.ReadFromResponse(ctx, options.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
