package resources

import (
	"context"
	"fmt"
	"net/http"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &notificationChannelResource{}
var _ resource.ResourceWithImportState = &notificationChannelResource{}
var _ resource.ResourceWithConfigure = &notificationChannelResource{}

func NewNotificationChannelResource() resource.Resource {
	return &notificationChannelResource{}
}

// notificationChannelResource defines the resource implementation.
type notificationChannelResource struct {
	platformClient *platform.ClientWithResponses
	organizationId string
}

func (r *notificationChannelResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_channel"
}

func (r *notificationChannelResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Notification Channel resource",
		Attributes:          schemas.NotificationChannelResourceSchemaAttributes(),
	}
}

func (r *notificationChannelResource) Configure(
	ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	apiClients, ok := req.ProviderData.(models.ApiClientsModel)
	if !ok {
		utils.ResourceApiClientConfigureError(ctx, req, resp)
		return
	}

	r.platformClient = apiClients.PlatformClient
	r.organizationId = apiClients.OrganizationId
}

func (r *notificationChannelResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.NotificationChannelResource

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var createNotificationChannelRequest platform.CreateNotificationChannelJSONRequestBody

	// Build create request based on notification channel type
	switch data.Type.ValueString() {
	case string(platform.AlertNotificationChannelTypeEMAIL):
		var definition models.NotificationChannelDefinition
		if errList := data.Definition.As(ctx, &definition, basetypes.ObjectAsOptions{}); errList.HasError() {
			resp.Diagnostics.Append(errList...)
			return
		}

		var recipients []string
		if errList := definition.Recipients.ElementsAs(ctx, &recipients, false); errList.HasError() {
			resp.Diagnostics.Append(errList...)
			return
		}

		createEmailNotificationChannelRequest := platform.CreateEmailNotificationChannelRequest{
			Name: data.Name.ValueString(),
			Definition: platform.EmailNotificationChannelDefinition{
				Recipients: recipients,
			},
			Type:       platform.CreateEmailNotificationChannelRequestType(data.Type.ValueString()),
			EntityId:   data.EntityId.ValueString(),
			EntityType: platform.CreateEmailNotificationChannelRequestEntityType(data.EntityType.ValueString()),
			IsShared:   data.IsShared.ValueBoolPointer(),
		}

		err := createNotificationChannelRequest.FromCreateEmailNotificationChannelRequest(createEmailNotificationChannelRequest)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build EMAIL Notification Channel request: %s", err))
			return
		}

	case string(platform.AlertNotificationChannelTypeSLACK):
		var definition models.NotificationChannelDefinition
		if errList := data.Definition.As(ctx, &definition, basetypes.ObjectAsOptions{}); errList.HasError() {
			resp.Diagnostics.Append(errList...)
			return
		}

		createSlackNotificationChannelRequest := platform.CreateSlackNotificationChannelRequest{
			Name: data.Name.ValueString(),
			Definition: platform.SlackNotificationChannelDefinition{
				WebhookUrl: definition.WebhookUrl.ValueString(),
			},
			Type:       platform.CreateSlackNotificationChannelRequestType(data.Type.ValueString()),
			EntityId:   data.EntityId.ValueString(),
			EntityType: platform.CreateSlackNotificationChannelRequestEntityType(data.EntityType.ValueString()),
			IsShared:   data.IsShared.ValueBoolPointer(),
		}

		err := createNotificationChannelRequest.FromCreateSlackNotificationChannelRequest(createSlackNotificationChannelRequest)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build SLACK Notification Channel request: %s", err))
			return
		}

	case string(platform.AlertNotificationChannelTypeDAGTRIGGER):
		var definition models.NotificationChannelDefinition
		if errList := data.Definition.As(ctx, &definition, basetypes.ObjectAsOptions{}); errList.HasError() {
			resp.Diagnostics.Append(errList...)
			return
		}

		createDagTriggerNotificationChannelRequest := platform.CreateDagTriggerNotificationChannelRequest{
			Name: data.Name.ValueString(),
			Definition: platform.DagTriggerNotificationChannelDefinition{
				DagId:              definition.DagId.ValueString(),
				DeploymentApiToken: definition.DeploymentApiToken.ValueString(),
				DeploymentId:       definition.DeploymentId.ValueString(),
			},
			Type:       platform.CreateDagTriggerNotificationChannelRequestType(data.Type.ValueString()),
			EntityId:   data.EntityId.ValueString(),
			EntityType: platform.CreateDagTriggerNotificationChannelRequestEntityType(data.EntityType.ValueString()),
			IsShared:   data.IsShared.ValueBoolPointer(),
		}

		err := createNotificationChannelRequest.FromCreateDagTriggerNotificationChannelRequest(createDagTriggerNotificationChannelRequest)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build DAGTRIGGER Notification Channel request: %s", err))
			return
		}

	case string(platform.AlertNotificationChannelTypePAGERDUTY):
		var definition models.NotificationChannelDefinition
		if errList := data.Definition.As(ctx, &definition, basetypes.ObjectAsOptions{}); errList.HasError() {
			resp.Diagnostics.Append(errList...)
			return
		}

		createPagerDutyNotificationChannelRequest := platform.CreatePagerDutyNotificationChannelRequest{
			Name: data.Name.ValueString(),
			Definition: platform.PagerDutyNotificationChannelDefinition{
				IntegrationKey: definition.IntegrationKey.ValueString(),
			},
			Type:       platform.CreatePagerDutyNotificationChannelRequestType(data.Type.ValueString()),
			EntityId:   data.EntityId.ValueString(),
			EntityType: platform.CreatePagerDutyNotificationChannelRequestEntityType(data.EntityType.ValueString()),
			IsShared:   data.IsShared.ValueBoolPointer(),
		}

		err := createNotificationChannelRequest.FromCreatePagerDutyNotificationChannelRequest(createPagerDutyNotificationChannelRequest)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build PAGERDUTY Notification Channel request: %s", err))
			return
		}

	case string(platform.AlertNotificationChannelTypeOPSGENIE):
		var definition models.NotificationChannelDefinition
		if errList := data.Definition.As(ctx, &definition, basetypes.ObjectAsOptions{}); errList.HasError() {
			resp.Diagnostics.Append(errList...)
			return
		}

		createOpsgenieNotificationChannelRequest := platform.CreateOpsgenieNotificationChannelRequest{
			Name: data.Name.ValueString(),
			Definition: platform.OpsgenieNotificationChannelDefinition{
				ApiKey: definition.ApiKey.ValueString(),
			},
			Type:       platform.CreateOpsgenieNotificationChannelRequestType(data.Type.ValueString()),
			EntityId:   data.EntityId.ValueString(),
			EntityType: platform.CreateOpsgenieNotificationChannelRequestEntityType(data.EntityType.ValueString()),
			IsShared:   data.IsShared.ValueBoolPointer(),
		}

		err := createNotificationChannelRequest.FromCreateOpsgenieNotificationChannelRequest(createOpsgenieNotificationChannelRequest)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build OPSGENIE Notification Channel request: %s", err))
			return
		}

	default:
		resp.Diagnostics.AddError("Invalid notification channel type", fmt.Sprintf("Unsupported notification channel type: %s", data.Type.ValueString()))
		return
	}

	// Call platform to create
	notificationChannelResp, err := r.platformClient.CreateNotificationChannelWithResponse(ctx, r.organizationId, createNotificationChannelRequest)
	if err != nil {
		tflog.Error(ctx, "failed to create notification channel", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create notification channel: %s", err))
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, notificationChannelResp.HTTPResponse, notificationChannelResp.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	// Map response into state
	diags := data.ReadFromResponse(ctx, notificationChannelResp.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created notification channel resource %s", data.Id.ValueString()))

	// Save to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *notificationChannelResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.NotificationChannelResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// get request
	notificationChannel, err := r.platformClient.GetNotificationChannelWithResponse(
		ctx,
		r.organizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to get notificationChannel", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get notificationChannel, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, notificationChannel.HTTPResponse, notificationChannel.Body)
	// If the resource no longer exists, it is recommended to ignore the errors
	// and call RemoveResource to remove the resource from the state. The next Terraform plan will recreate the resource.
	if statusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	diags := data.ReadFromResponse(ctx, notificationChannel.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read a notificationChannel resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *notificationChannelResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data models.NotificationChannelResource

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build update request based on notification channel type
	var updateBody platform.UpdateNotificationChannelJSONRequestBody

	switch data.Type.ValueString() {
	case string(platform.AlertNotificationChannelTypeEMAIL):
		var definition models.NotificationChannelDefinition
		if errList := data.Definition.As(ctx, &definition, basetypes.ObjectAsOptions{}); errList.HasError() {
			resp.Diagnostics.Append(errList...)
			return
		}

		var recipients []string
		if errList := definition.Recipients.ElementsAs(ctx, &recipients, false); errList.HasError() {
			resp.Diagnostics.Append(errList...)
			return
		}

		name := data.Name.ValueString()
		channelType := platform.UpdateEmailNotificationChannelRequestType(data.Type.ValueString())
		updateEmailNotificationChannelRequest := platform.UpdateEmailNotificationChannelRequest{
			Name: &name,
			Definition: &platform.EmailNotificationChannelDefinition{
				Recipients: recipients,
			},
			Type:     &channelType,
			IsShared: data.IsShared.ValueBoolPointer(),
		}

		err := updateBody.FromUpdateEmailNotificationChannelRequest(updateEmailNotificationChannelRequest)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build EMAIL Notification Channel update request: %s", err))
			return
		}

	case string(platform.AlertNotificationChannelTypeSLACK):
		var definition models.NotificationChannelDefinition
		if errList := data.Definition.As(ctx, &definition, basetypes.ObjectAsOptions{}); errList.HasError() {
			resp.Diagnostics.Append(errList...)
			return
		}

		name := data.Name.ValueString()
		channelType := platform.UpdateSlackNotificationChannelRequestType(data.Type.ValueString())
		updateSlackNotificationChannelRequest := platform.UpdateSlackNotificationChannelRequest{
			Name: &name,
			Definition: &platform.SlackNotificationChannelDefinition{
				WebhookUrl: definition.WebhookUrl.ValueString(),
			},
			Type:     &channelType,
			IsShared: data.IsShared.ValueBoolPointer(),
		}

		err := updateBody.FromUpdateSlackNotificationChannelRequest(updateSlackNotificationChannelRequest)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build SLACK Notification Channel update request: %s", err))
			return
		}

	case string(platform.AlertNotificationChannelTypeDAGTRIGGER):
		var definition models.NotificationChannelDefinition
		if errList := data.Definition.As(ctx, &definition, basetypes.ObjectAsOptions{}); errList.HasError() {
			resp.Diagnostics.Append(errList...)
			return
		}

		name := data.Name.ValueString()
		channelType := platform.UpdateDagTriggerNotificationChannelRequestType(data.Type.ValueString())
		updateDagTriggerNotificationChannelRequest := platform.UpdateDagTriggerNotificationChannelRequest{
			Name: &name,
			Definition: &platform.DagTriggerNotificationChannelDefinition{
				DagId:              definition.DagId.ValueString(),
				DeploymentApiToken: definition.DeploymentApiToken.ValueString(),
				DeploymentId:       definition.DeploymentId.ValueString(),
			},
			Type:     &channelType,
			IsShared: data.IsShared.ValueBoolPointer(),
		}

		err := updateBody.FromUpdateDagTriggerNotificationChannelRequest(updateDagTriggerNotificationChannelRequest)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build DAGTRIGGER Notification Channel update request: %s", err))
			return
		}

	case string(platform.AlertNotificationChannelTypePAGERDUTY):
		var definition models.NotificationChannelDefinition
		if errList := data.Definition.As(ctx, &definition, basetypes.ObjectAsOptions{}); errList.HasError() {
			resp.Diagnostics.Append(errList...)
			return
		}

		name := data.Name.ValueString()
		channelType := platform.UpdatePagerDutyNotificationChannelRequestType(data.Type.ValueString())
		updatePagerDutyNotificationChannelRequest := platform.UpdatePagerDutyNotificationChannelRequest{
			Name: &name,
			Definition: &platform.PagerDutyNotificationChannelDefinition{
				IntegrationKey: definition.IntegrationKey.ValueString(),
			},
			Type:     &channelType,
			IsShared: data.IsShared.ValueBoolPointer(),
		}

		err := updateBody.FromUpdatePagerDutyNotificationChannelRequest(updatePagerDutyNotificationChannelRequest)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build PAGERDUTY Notification Channel update request: %s", err))
			return
		}

	case string(platform.AlertNotificationChannelTypeOPSGENIE):
		var definition models.NotificationChannelDefinition
		if errList := data.Definition.As(ctx, &definition, basetypes.ObjectAsOptions{}); errList.HasError() {
			resp.Diagnostics.Append(errList...)
			return
		}

		name := data.Name.ValueString()
		channelType := platform.UpdateOpsgenieNotificationChannelRequestType(data.Type.ValueString())
		updateOpsgenieNotificationChannelRequest := platform.UpdateOpsgenieNotificationChannelRequest{
			Name: &name,
			Definition: &platform.OpsgenieNotificationChannelDefinition{
				ApiKey: definition.ApiKey.ValueString(),
			},
			Type:     &channelType,
			IsShared: data.IsShared.ValueBoolPointer(),
		}

		err := updateBody.FromUpdateOpsgenieNotificationChannelRequest(updateOpsgenieNotificationChannelRequest)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build OPSGENIE Notification Channel update request: %s", err))
			return
		}

	default:
		resp.Diagnostics.AddError("Invalid notification channel type", fmt.Sprintf("Unsupported notification channel type: %s", data.Type.ValueString()))
		return
	}

	// Call the API to update the notification channel
	notificationChannelResp, err := r.platformClient.UpdateNotificationChannelWithResponse(ctx, r.organizationId, data.Id.ValueString(), updateBody)
	if err != nil {
		tflog.Error(ctx, "failed to update notification channel", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to update notification channel: %s", err))
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, notificationChannelResp.HTTPResponse, notificationChannelResp.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	// Map updated response
	diags := data.ReadFromResponse(ctx, notificationChannelResp.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated notification channel resource %s", data.Id.ValueString()))

	// Save to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *notificationChannelResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.NotificationChannelResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// delete request
	notificationChannel, err := r.platformClient.DeleteNotificationChannelWithResponse(
		ctx,
		r.organizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to delete notificationChannel", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete notificationChannel, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, notificationChannel.HTTPResponse, notificationChannel.Body)
	// It is recommended to ignore 404 Resource Not Found errors when deleting a resource
	if statusCode != http.StatusNotFound && diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted a notificationChannel resource: %v", data.Id.ValueString()))
}

func (r *notificationChannelResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
