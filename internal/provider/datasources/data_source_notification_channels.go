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
var _ datasource.DataSource = &notificationChannelsDataSource{}
var _ datasource.DataSourceWithConfigure = &notificationChannelsDataSource{}

func NewNotificationChannelsDataSource() datasource.DataSource {
	return &notificationChannelsDataSource{}
}

// notificationChannelDataSource defines the data source implementation.
type notificationChannelsDataSource struct {
	PlatformClient platform.ClientWithResponsesInterface
	OrganizationId string
}

func (d *notificationChannelsDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_channels"
}

func (d *notificationChannelsDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Notification Channels data source",
		Attributes:          schemas.NotificationChannelsDataSourceSchemaAttributes(),
	}
}

func (d *notificationChannelsDataSource) Configure(
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

func (d *notificationChannelsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data models.NotificationChannels

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := &platform.ListNotificationChannelsParams{
		Limit: lo.ToPtr(1000),
	}
	var diags diag.Diagnostics

	notificationChannelIds, diags := utils.TypesSetToStringSlice(ctx, data.NotificationChannelIds)
	if len(notificationChannelIds) > 0 {
		params.NotificationChannelIds = &notificationChannelIds
	}
	if diags.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read notification channels, got error %v", diags.Errors()[0].Summary()))
		resp.Diagnostics.Append(diags...)
		return
	}

	notificationChannelTypes, diags := utils.TypesSetToStringSlice(ctx, data.ChannelTypes)
	if len(notificationChannelTypes) > 0 {
		notificationChannelTypes := lo.Map(notificationChannelTypes, func(t string, _ int) platform.ListNotificationChannelsParamsChannelTypes {
			return platform.ListNotificationChannelsParamsChannelTypes(t)
		})
		params.ChannelTypes = &notificationChannelTypes
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
		params.EntityType = lo.ToPtr(platform.ListNotificationChannelsParamsEntityType(entityType))
	}

	var notificationChannels []platform.NotificationChannel
	offset := 0
	for {
		params.Offset = &offset
		notificationChannelsResp, err := d.PlatformClient.ListNotificationChannelsWithResponse(
			ctx,
			d.OrganizationId,
			params,
		)
		if err != nil {
			tflog.Error(ctx, "failed to list notification channels", map[string]interface{}{"error": err})
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to read notification channels, got error: %s", err),
			)
			return
		}
		_, diagnostic := clients.NormalizeAPIError(ctx, notificationChannelsResp.HTTPResponse, notificationChannelsResp.Body)
		if diagnostic != nil {
			resp.Diagnostics.Append(diagnostic)
			return
		}
		if notificationChannelsResp.JSON200 == nil {
			tflog.Error(ctx, "failed to list notification channels", map[string]interface{}{"error": "nil response"})
			resp.Diagnostics.AddError("Client Error", "Unable to read notification channels, got nil response")
			return
		}

		notificationChannels = append(notificationChannels, notificationChannelsResp.JSON200.NotificationChannels...)

		if notificationChannelsResp.JSON200.TotalCount <= offset {
			break
		}
		offset += 1000
	}

	// Populate the model with the response data
	diags = data.ReadFromResponse(ctx, notificationChannels)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
