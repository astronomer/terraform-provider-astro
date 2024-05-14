package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/samber/lo"

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
var _ datasource.DataSource = &deploymentsDataSource{}
var _ datasource.DataSourceWithConfigure = &deploymentsDataSource{}

func NewDeploymentsDataSource() datasource.DataSource {
	return &deploymentsDataSource{}
}

// deploymentsDataSource defines the data source implementation.
type deploymentsDataSource struct {
	PlatformClient platform.ClientWithResponsesInterface
	OrganizationId string
}

func (d *deploymentsDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_deployments"
}

func (d *deploymentsDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Deployments data source",
		Attributes:          schemas.DeploymentsDataSourceSchemaAttributes(),
	}
}

func (d *deploymentsDataSource) Configure(
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

func (d *deploymentsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data models.Deployments

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := &platform.ListDeploymentsParams{
		Limit: lo.ToPtr(1000),
	}
	var diags diag.Diagnostics
	deploymentIds, diags := utils.TypesSetToStringSlice(ctx, data.DeploymentIds)
	if len(deploymentIds) > 0 {
		params.DeploymentIds = &deploymentIds
	}
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	workspaceIds, diags := utils.TypesSetToStringSlice(ctx, data.WorkspaceIds)
	if len(workspaceIds) > 0 {
		params.WorkspaceIds = &workspaceIds
	}
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	names, diags := utils.TypesSetToStringSlice(ctx, data.Names)
	if len(names) > 0 {
		params.Names = &names
	}
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var deployments []platform.Deployment
	offset := 0
	for {
		params.Offset = &offset
		deploymentsResp, err := d.PlatformClient.ListDeploymentsWithResponse(
			ctx,
			d.OrganizationId,
			params,
		)
		if err != nil {
			tflog.Error(ctx, "failed to list deployments", map[string]interface{}{"error": err})
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to read deployments, got error: %s", err),
			)
			return
		}
		_, diagnostic := clients.NormalizeAPIError(ctx, deploymentsResp.HTTPResponse, deploymentsResp.Body)
		if diagnostic != nil {
			resp.Diagnostics.Append(diagnostic)
			return
		}
		if deploymentsResp.JSON200 == nil {
			tflog.Error(ctx, "failed to list deployments", map[string]interface{}{"error": "nil response"})
			resp.Diagnostics.AddError("Client Error", "Unable to read deployments, got nil response")
			return
		}

		deployments = append(deployments, deploymentsResp.JSON200.Deployments...)

		if deploymentsResp.JSON200.TotalCount <= offset {
			break
		}

		offset += 1000
	}

	// Populate the model with the response data
	diags = data.ReadFromResponse(ctx, deployments)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
