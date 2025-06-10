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
var _ datasource.DataSource = &alertsDataSource{}
var _ datasource.DataSourceWithConfigure = &alertsDataSource{}

func NewAlertsDataSource() datasource.DataSource {
	return &alertsDataSource{}
}

// alertsDataSource defines the data source implementation.
type alertsDataSource struct {
	PlatformClient platform.ClientWithResponsesInterface
	OrganizationId string
}

func (d *alertsDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_alerts"
}

func (d *alertsDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Alerts data source",
		Attributes:          schemas.AlertsDataSourceSchemaAttributes(),
	}
}

func (d *alertsDataSource) Configure(
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

func (d *alertsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data models.Alerts

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := &platform.ListAlertsParams{
		Limit: lo.ToPtr(1000),
	}
	var diags diag.Diagnostics

	alertIds, diags := utils.TypesSetToStringSlice(ctx, data.AlertIds)
	if len(alertIds) > 0 {
		params.AlertIds = &alertIds
	}
	if diags.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read alerts, got error %v", diags.Errors()[0].Summary()))
		resp.Diagnostics.Append(diags...)
		return
	}

	alertTypes, diags := utils.TypesSetToStringSlice(ctx, data.AlertTypes)
	if len(alertTypes) > 0 {
		alertTypes := lo.Map(alertTypes, func(t string, _ int) platform.ListAlertsParamsAlertTypes {
			return platform.ListAlertsParamsAlertTypes(t)
		})
		params.AlertTypes = &alertTypes
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

	deploymentIds, diags := utils.TypesSetToStringSlice(ctx, data.DeploymentIds)
	if len(deploymentIds) > 0 {
		params.DeploymentIds = &deploymentIds
	}
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	entityType := data.EntityType.ValueString()
	if entityType != "" {
		params.EntityType = lo.ToPtr(platform.ListAlertsParamsEntityType(entityType))
	}

	var alerts []platform.Alert
	offset := 0
	for {
		params.Offset = &offset
		alertsResp, err := d.PlatformClient.ListAlertsWithResponse(
			ctx,
			d.OrganizationId,
			params,
		)
		if err != nil {
			tflog.Error(ctx, "failed to list alerts", map[string]interface{}{"error": err})
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to read alerts, got error: %s", err),
			)
			return
		}
		_, diagnostic := clients.NormalizeAPIError(ctx, alertsResp.HTTPResponse, alertsResp.Body)
		if diagnostic != nil {
			resp.Diagnostics.Append(diagnostic)
			return
		}
		if alertsResp.JSON200 == nil {
			tflog.Error(ctx, "failed to list alerts", map[string]interface{}{"error": "nil response"})
			resp.Diagnostics.AddError("Client Error", "Unable to read alerts, got nil response")
			return
		}

		alerts = append(alerts, alertsResp.JSON200.Alerts...)

		if alertsResp.JSON200.TotalCount <= offset {
			break
		}

		offset += 1000
	}

	// Populate the model with the response data
	diags = data.ReadFromResponse(ctx, alerts)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
