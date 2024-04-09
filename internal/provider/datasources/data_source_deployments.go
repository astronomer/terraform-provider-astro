package datasources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/samber/lo"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients"
	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/models"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/schemas"
	"github.com/astronomer/astronomer-terraform-provider/internal/utils"
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
	var data models.DeploymentsDataSource

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	params := &platform.ListDeploymentsParams{
		Limit: lo.ToPtr(1000),
	}
	deploymentIds := data.DeploymentIds.Elements()
	if len(deploymentIds) > 0 {
		deploymentIdsParam := lo.Map(deploymentIds, func(id attr.Value, _ int) string {
			// Terraform includes quotes around the string, so we need to remove them
			return strings.ReplaceAll(id.String(), `"`, "")
		})
		params.DeploymentIds = &deploymentIdsParam
	}
	names := data.Names.Elements()
	if len(names) > 0 {
		namesParam := lo.Map(names, func(name attr.Value, _ int) string {
			// Terraform includes quotes around the string, so we need to remove them
			return strings.ReplaceAll(name.String(), `"`, "")
		})
		params.Names = &namesParam
	}

	if resp.Diagnostics.HasError() {
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
			tflog.Error(ctx, "failed to get deployment", map[string]interface{}{"error": err})
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to read deployment, got error: %s", err),
			)
			return
		}
		_, diagnostic := clients.NormalizeAPIError(ctx, deploymentsResp.HTTPResponse, deploymentsResp.Body)
		if diagnostic != nil {
			resp.Diagnostics.Append(diagnostic)
			return
		}
		if deploymentsResp.JSON200 == nil {
			tflog.Error(ctx, "failed to get deployment", map[string]interface{}{"error": "nil response"})
			resp.Diagnostics.AddError("Client Error", "Unable to read deployment, got nil response")
			return
		}

		deployments = append(deployments, deploymentsResp.JSON200.Deployments...)

		if deploymentsResp.JSON200.TotalCount <= offset {
			break
		}

		offset += 1000
	}

	// Populate the model with the response data
	diags := data.ReadFromResponse(ctx, deployments)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
