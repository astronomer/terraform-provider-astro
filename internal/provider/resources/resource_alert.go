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
var _ resource.Resource = &alertResource{}
var _ resource.ResourceWithImportState = &alertResource{}
var _ resource.ResourceWithConfigure = &alertResource{}

func NewAlertResource() resource.Resource {
	return &alertResource{}
}

// alertResource defines the resource implementation.
type alertResource struct {
	platformClient *platform.ClientWithResponses
	organizationId string
}

func (r *alertResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_alert"
}

func (r *alertResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Alert resource",
		Attributes:          schemas.AlertResourceSchemaAttributes(),
	}
}

func (r *alertResource) Configure(
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

func (r *alertResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.AlertResource

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var createAlertRequest platform.CreateAlertJSONRequestBody
	notificationChannelIds, diags := utils.TypesSetToStringSlice(ctx, data.NotificationChannelIds)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Build create request based on alert type
	switch data.Type.ValueString() {
	case string(platform.AlertTypeDAGFAILURE):
		var alertRulesInput models.ResourceAlertRulesInput
		diags := data.Rules.As(ctx, &alertRulesInput, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		// turn those into the API's PatternMatchRequest
		pmReqs := make([]platform.PatternMatchRequest, len(alertRulesInput.PatternMatches))
		for i, pm := range alertRulesInput.PatternMatches {
			pmReqs[i] = platform.PatternMatchRequest{
				EntityType:   platform.PatternMatchRequestEntityType(pm.EntityType),
				OperatorType: platform.PatternMatchRequestOperatorType(pm.OperatorType),
				Values:       pm.Values,
			}
		}

		createDagFailureAlertRequest := platform.CreateDagFailureAlertRequest{
			EntityId:               data.EntityId.ValueString(),
			EntityType:             platform.CreateDagFailureAlertRequestEntityType(data.EntityType.ValueString()),
			Name:                   data.Name.ValueString(),
			NotificationChannelIds: notificationChannelIds,
			Severity:               platform.CreateDagFailureAlertRequestSeverity(data.Severity.ValueString()),
			Type:                   platform.CreateDagFailureAlertRequestType(data.Type.ValueString()),
			Rules: platform.CreateDagFailureAlertRules{
				PatternMatches: pmReqs,
				Properties: platform.CreateDagFailureAlertProperties{
					DeploymentId: alertRulesInput.Properties.DeploymentId,
				},
			},
		}

		err := createAlertRequest.FromCreateDagFailureAlertRequest(createDagFailureAlertRequest)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build DAG_FAILURE request: %s", err))
			return
		}

	case string(platform.AlertTypeDAGSUCCESS):
		// decode the Terraform `rules` nested block
		var alertRulesInput models.ResourceAlertRulesInput
		diags := data.Rules.As(ctx, &alertRulesInput, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		// turn those into the API's PatternMatchRequest
		pmReqs := make([]platform.PatternMatchRequest, len(alertRulesInput.PatternMatches))
		for i, pm := range alertRulesInput.PatternMatches {
			pmReqs[i] = platform.PatternMatchRequest{
				EntityType:   platform.PatternMatchRequestEntityType(pm.EntityType),
				OperatorType: platform.PatternMatchRequestOperatorType(pm.OperatorType),
				Values:       pm.Values,
			}
		}

		createDagSuccessAlertRequest := platform.CreateDagSuccessAlertRequest{
			EntityId:               data.EntityId.ValueString(),
			EntityType:             platform.CreateDagSuccessAlertRequestEntityType(data.EntityType.ValueString()),
			Name:                   data.Name.ValueString(),
			NotificationChannelIds: notificationChannelIds,
			Severity:               platform.CreateDagSuccessAlertRequestSeverity(data.Severity.ValueString()),
			Type:                   platform.CreateDagSuccessAlertRequestType(data.Type.ValueString()),
			Rules: platform.CreateDagSuccessAlertRules{
				PatternMatches: pmReqs,
				Properties: platform.CreateDagSuccessAlertProperties{
					DeploymentId: alertRulesInput.Properties.DeploymentId,
				},
			},
		}
		err := createAlertRequest.FromCreateDagSuccessAlertRequest(createDagSuccessAlertRequest)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build DAG_SUCCESS request: %s", err))
			return
		}

	case string(platform.AlertTypeDAGDURATION):
		var alertRulesInput models.ResourceAlertRulesInput
		diags := data.Rules.As(ctx, &alertRulesInput, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		pmReqs := make([]platform.PatternMatchRequest, len(alertRulesInput.PatternMatches))
		for i, pm := range alertRulesInput.PatternMatches {
			pmReqs[i] = platform.PatternMatchRequest{
				EntityType:   platform.PatternMatchRequestEntityType(pm.EntityType),
				OperatorType: platform.PatternMatchRequestOperatorType(pm.OperatorType),
				Values:       pm.Values,
			}
		}

		createDagDurationAlertRequest := platform.CreateDagDurationAlertRequest{
			EntityId:               data.EntityId.ValueString(),
			EntityType:             platform.CreateDagDurationAlertRequestEntityType(data.EntityType.ValueString()),
			Name:                   data.Name.ValueString(),
			NotificationChannelIds: notificationChannelIds,
			Severity:               platform.CreateDagDurationAlertRequestSeverity(data.Severity.ValueString()),
			Type:                   platform.CreateDagDurationAlertRequestType(data.Type.ValueString()),
			Rules: platform.CreateDagDurationAlertRules{
				PatternMatches: pmReqs,
				Properties: platform.CreateDagDurationAlertProperties{
					DeploymentId:       alertRulesInput.Properties.DeploymentId,
					DagDurationSeconds: int(alertRulesInput.Properties.DagDurationSeconds),
				},
			},
		}
		err := createAlertRequest.FromCreateDagDurationAlertRequest(createDagDurationAlertRequest)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build DAG_DURATION request: %s", err))
			return
		}

	case string(platform.AlertTypeDAGTIMELINESS):
		var alertRulesInput models.ResourceAlertRulesInput
		diags := data.Rules.As(ctx, &alertRulesInput, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		pmReqs := make([]platform.PatternMatchRequest, len(alertRulesInput.PatternMatches))
		for i, pm := range alertRulesInput.PatternMatches {
			pmReqs[i] = platform.PatternMatchRequest{
				EntityType:   platform.PatternMatchRequestEntityType(pm.EntityType),
				OperatorType: platform.PatternMatchRequestOperatorType(pm.OperatorType),
				Values:       pm.Values,
			}
		}

		createDagTimelinessAlertRequest := platform.CreateDagTimelinessAlertRequest{
			EntityId:               data.EntityId.ValueString(),
			EntityType:             platform.CreateDagTimelinessAlertRequestEntityType(data.EntityType.ValueString()),
			Name:                   data.Name.ValueString(),
			NotificationChannelIds: notificationChannelIds,
			Severity:               platform.CreateDagTimelinessAlertRequestSeverity(data.Severity.ValueString()),
			Type:                   platform.CreateDagTimelinessAlertRequestType(data.Type.ValueString()),
			Rules: platform.CreateDagTimelinessAlertRules{
				PatternMatches: pmReqs,
				Properties: platform.CreateDagTimelinessAlertProperties{
					DeploymentId:          alertRulesInput.Properties.DeploymentId,
					DagDeadline:           alertRulesInput.Properties.DagDeadline,
					DaysOfWeek:            alertRulesInput.Properties.DaysOfWeek,
					LookBackPeriodSeconds: int(alertRulesInput.Properties.LookBackPeriodSeconds),
				},
			},
		}
		err := createAlertRequest.FromCreateDagTimelinessAlertRequest(createDagTimelinessAlertRequest)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build DAG_TIMELINESS request: %s", err))
			return
		}

	case string(platform.AlertTypeTASKFAILURE):
		var alertRulesInput models.ResourceAlertRulesInput
		diags := data.Rules.As(ctx, &alertRulesInput, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		pmReqs := make([]platform.PatternMatchRequest, len(alertRulesInput.PatternMatches))
		for i, pm := range alertRulesInput.PatternMatches {
			pmReqs[i] = platform.PatternMatchRequest{
				EntityType:   platform.PatternMatchRequestEntityType(pm.EntityType),
				OperatorType: platform.PatternMatchRequestOperatorType(pm.OperatorType),
				Values:       pm.Values,
			}
		}

		createTaskFailureAlertRequest := platform.CreateTaskFailureAlertRequest{
			EntityId:               data.EntityId.ValueString(),
			EntityType:             platform.CreateTaskFailureAlertRequestEntityType(data.EntityType.ValueString()),
			Name:                   data.Name.ValueString(),
			NotificationChannelIds: notificationChannelIds,
			Severity:               platform.CreateTaskFailureAlertRequestSeverity(data.Severity.ValueString()),
			Type:                   platform.CreateTaskFailureAlertRequestType(data.Type.ValueString()),
			Rules: platform.CreateTaskFailureAlertRules{
				PatternMatches: pmReqs,
				Properties: platform.CreateTaskFailureAlertProperties{
					DeploymentId: alertRulesInput.Properties.DeploymentId,
				},
			},
		}
		err := createAlertRequest.FromCreateTaskFailureAlertRequest(createTaskFailureAlertRequest)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build TASK_FAILURE request: %s", err))
			return
		}

	case string(platform.AlertTypeTASKDURATION):
		var alertRulesInput models.ResourceAlertRulesInput
		diags := data.Rules.As(ctx, &alertRulesInput, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		pmReqs := make([]platform.PatternMatchRequest, len(alertRulesInput.PatternMatches))
		for i, pm := range alertRulesInput.PatternMatches {
			pmReqs[i] = platform.PatternMatchRequest{
				EntityType:   platform.PatternMatchRequestEntityType(pm.EntityType),
				OperatorType: platform.PatternMatchRequestOperatorType(pm.OperatorType),
				Values:       pm.Values,
			}
		}

		createTaskDurationAlertRequest := platform.CreateTaskDurationAlertRequest{
			EntityId:               data.EntityId.ValueString(),
			EntityType:             platform.CreateTaskDurationAlertRequestEntityType(data.EntityType.ValueString()),
			Name:                   data.Name.ValueString(),
			NotificationChannelIds: notificationChannelIds,
			Severity:               platform.CreateTaskDurationAlertRequestSeverity(data.Severity.ValueString()),
			Type:                   platform.CreateTaskDurationAlertRequestType(data.Type.ValueString()),
			Rules: platform.CreateTaskDurationAlertRules{
				PatternMatches: pmReqs,
				Properties: platform.CreateTaskDurationAlertProperties{
					DeploymentId:        alertRulesInput.Properties.DeploymentId,
					TaskDurationSeconds: int(alertRulesInput.Properties.TaskDurationSeconds),
				},
			},
		}
		err := createAlertRequest.FromCreateTaskDurationAlertRequest(createTaskDurationAlertRequest)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build TASK_DURATION request: %s", err))
			return
		}

	default:
		resp.Diagnostics.AddError("Invalid alert type", fmt.Sprintf("Unsupported alert type: %s", data.Type.ValueString()))
		return
	}

	// Call platform to create
	alertResp, err := r.platformClient.CreateAlertWithResponse(ctx, r.organizationId, createAlertRequest)
	if err != nil {
		tflog.Error(ctx, "failed to create alert", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create alert: %s", err))
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, alertResp.HTTPResponse, alertResp.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	// Map response into state
	diags = data.ReadFromResponse(ctx, alertResp.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created alert resource %s", data.Id.ValueString()))

	// Save to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *alertResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.AlertResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// get request
	alert, err := r.platformClient.GetAlertWithResponse(
		ctx,
		r.organizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to get alert", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get alert, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, alert.HTTPResponse, alert.Body)
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

	diags := data.ReadFromResponse(ctx, alert.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read a alert resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *alertResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data models.AlertResource

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var updateBody platform.UpdateAlertJSONRequestBody
	// Build notification channel IDs slice
	ncIds, diags := utils.TypesSetToStringSlice(ctx, data.NotificationChannelIds)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Build update request based on alert type
	switch data.Type.ValueString() {
	case string(platform.AlertTypeDAGFAILURE):
		var alertRulesInput models.ResourceAlertRulesInput
		diags := data.Rules.As(ctx, &alertRulesInput, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		pmReqs := make([]platform.PatternMatchRequest, len(alertRulesInput.PatternMatches))
		for i, pm := range alertRulesInput.PatternMatches {
			pmReqs[i] = platform.PatternMatchRequest{
				EntityType:   platform.PatternMatchRequestEntityType(pm.EntityType),
				OperatorType: platform.PatternMatchRequestOperatorType(pm.OperatorType),
				Values:       pm.Values,
			}
		}

		name := data.Name.ValueString()
		sev := platform.UpdateDagFailureAlertRequestSeverity(data.Severity.ValueString())

		reqModel := platform.UpdateDagFailureAlertRequest{
			Name:                   &name,
			Severity:               &sev,
			NotificationChannelIds: &ncIds,
			Rules: &platform.UpdateDagFailureAlertRules{
				PatternMatches: &pmReqs,
			},
		}
		err := updateBody.FromUpdateDagFailureAlertRequest(reqModel)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build update for DAG_FAILURE: %s", err))
			return
		}

	case string(platform.AlertTypeDAGSUCCESS):
		var alertRulesInput models.ResourceAlertRulesInput
		diags := data.Rules.As(ctx, &alertRulesInput, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		pmReqs := make([]platform.PatternMatchRequest, len(alertRulesInput.PatternMatches))
		for i, pm := range alertRulesInput.PatternMatches {
			pmReqs[i] = platform.PatternMatchRequest{
				EntityType:   platform.PatternMatchRequestEntityType(pm.EntityType),
				OperatorType: platform.PatternMatchRequestOperatorType(pm.OperatorType),
				Values:       pm.Values,
			}
		}

		name := data.Name.ValueString()
		sev := platform.UpdateDagSuccessAlertRequestSeverity(data.Severity.ValueString())
		reqModel := platform.UpdateDagSuccessAlertRequest{
			Name:                   &name,
			NotificationChannelIds: &ncIds,
			Severity:               &sev,
			Rules: &platform.UpdateDagSuccessAlertRules{
				PatternMatches: &pmReqs,
			},
		}
		err := updateBody.FromUpdateDagSuccessAlertRequest(reqModel)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build update for DAG_SUCCESS: %s", err))
			return
		}

	case string(platform.AlertTypeDAGDURATION):
		var alertRulesInput models.ResourceAlertRulesInput
		diags := data.Rules.As(ctx, &alertRulesInput, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		pmReqs := make([]platform.PatternMatchRequest, len(alertRulesInput.PatternMatches))
		for i, pm := range alertRulesInput.PatternMatches {
			pmReqs[i] = platform.PatternMatchRequest{
				EntityType:   platform.PatternMatchRequestEntityType(pm.EntityType),
				OperatorType: platform.PatternMatchRequestOperatorType(pm.OperatorType),
				Values:       pm.Values,
			}
		}

		name := data.Name.ValueString()
		sev := platform.UpdateDagDurationAlertRequestSeverity(data.Severity.ValueString())
		dagDurationSeconds := int(alertRulesInput.Properties.DagDurationSeconds)

		reqModel := platform.UpdateDagDurationAlertRequest{
			Name:                   &name,
			NotificationChannelIds: &ncIds,
			Severity:               &sev,
			Rules: &platform.UpdateDagDurationAlertRules{
				PatternMatches: &pmReqs,
				Properties: &platform.UpdateDagDurationAlertProperties{
					DagDurationSeconds: &dagDurationSeconds,
				},
			},
		}
		err := updateBody.FromUpdateDagDurationAlertRequest(reqModel)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build update for DAG_DURATION: %s", err))
			return
		}

	case string(platform.AlertTypeDAGTIMELINESS):
		var alertRulesInput models.ResourceAlertRulesInput
		diags := data.Rules.As(ctx, &alertRulesInput, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		pmReqs := make([]platform.PatternMatchRequest, len(alertRulesInput.PatternMatches))
		for i, pm := range alertRulesInput.PatternMatches {
			pmReqs[i] = platform.PatternMatchRequest{
				EntityType:   platform.PatternMatchRequestEntityType(pm.EntityType),
				OperatorType: platform.PatternMatchRequestOperatorType(pm.OperatorType),
				Values:       pm.Values,
			}
		}

		name := data.Name.ValueString()
		sev := platform.UpdateDagTimelinessAlertRequestSeverity(data.Severity.ValueString())
		lookBackPeriodSeconds := int(alertRulesInput.Properties.LookBackPeriodSeconds)

		reqModel := platform.UpdateDagTimelinessAlertRequest{
			Name:                   &name,
			NotificationChannelIds: &ncIds,
			Severity:               &sev,
			Rules: &platform.UpdateDagTimelinessAlertRules{
				PatternMatches: &pmReqs,
				Properties: &platform.UpdateDagTimelinessAlertProperties{
					DagDeadline:           &alertRulesInput.Properties.DagDeadline,
					DaysOfWeek:            &alertRulesInput.Properties.DaysOfWeek,
					LookBackPeriodSeconds: &lookBackPeriodSeconds,
				},
			},
		}
		err := updateBody.FromUpdateDagTimelinessAlertRequest(reqModel)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build update for DAG_TIMELINESS: %s", err))
			return
		}

	case string(platform.AlertTypeTASKFAILURE):
		var alertRulesInput models.ResourceAlertRulesInput
		diags := data.Rules.As(ctx, &alertRulesInput, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		pmReqs := make([]platform.PatternMatchRequest, len(alertRulesInput.PatternMatches))
		for i, pm := range alertRulesInput.PatternMatches {
			pmReqs[i] = platform.PatternMatchRequest{
				EntityType:   platform.PatternMatchRequestEntityType(pm.EntityType),
				OperatorType: platform.PatternMatchRequestOperatorType(pm.OperatorType),
				Values:       pm.Values,
			}
		}

		name := data.Name.ValueString()
		sev := platform.UpdateTaskFailureAlertRequestSeverity(data.Severity.ValueString())

		reqModel := platform.UpdateTaskFailureAlertRequest{
			Name:                   &name,
			NotificationChannelIds: &ncIds,
			Severity:               &sev,
			Rules: &platform.UpdateTaskFailureAlertRules{
				PatternMatches: &pmReqs,
			},
		}
		err := updateBody.FromUpdateTaskFailureAlertRequest(reqModel)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build update for TASK_FAILURE: %s", err))
			return
		}

	case string(platform.AlertTypeTASKDURATION):
		var alertRulesInput models.ResourceAlertRulesInput
		diags := data.Rules.As(ctx, &alertRulesInput, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		pmReqs := make([]platform.PatternMatchRequest, len(alertRulesInput.PatternMatches))
		for i, pm := range alertRulesInput.PatternMatches {
			pmReqs[i] = platform.PatternMatchRequest{
				EntityType:   platform.PatternMatchRequestEntityType(pm.EntityType),
				OperatorType: platform.PatternMatchRequestOperatorType(pm.OperatorType),
				Values:       pm.Values,
			}
		}

		name := data.Name.ValueString()
		sev := platform.UpdateTaskDurationAlertRequestSeverity(data.Severity.ValueString())
		taskDurationSeconds := int(alertRulesInput.Properties.TaskDurationSeconds)

		reqModel := platform.UpdateTaskDurationAlertRequest{
			Name:                   &name,
			NotificationChannelIds: &ncIds,
			Severity:               &sev,
			Rules: &platform.UpdateTaskDurationAlertRules{
				PatternMatches: &pmReqs,
				Properties: &platform.UpdateTaskDurationAlertProperties{
					TaskDurationSeconds: &taskDurationSeconds,
				},
			},
		}
		err := updateBody.FromUpdateTaskDurationAlertRequest(reqModel)
		if err != nil {
			resp.Diagnostics.AddError("Internal Error", fmt.Sprintf("failed to build update for TASK_DURATION: %s", err))
			return
		}

	default:
		resp.Diagnostics.AddError("Invalid alert type", fmt.Sprintf("Unsupported alert type: %s", data.Type.ValueString()))
		return
	}

	// Call platform update
	alertResp, err := r.platformClient.UpdateAlertWithResponse(ctx, r.organizationId, data.Id.ValueString(), updateBody)
	if err != nil {
		tflog.Error(ctx, "failed to update alert", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update alert: %s", err))
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, alertResp.HTTPResponse, alertResp.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	// Map updated response
	diags = data.ReadFromResponse(ctx, alertResp.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated alert resource %s", data.Id.ValueString()))

	// Save to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *alertResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.AlertResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// delete request
	alert, err := r.platformClient.DeleteAlertWithResponse(
		ctx,
		r.organizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to delete alert", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete alert, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, alert.HTTPResponse, alert.Body)
	// It is recommended to ignore 404 Resource Not Found errors when deleting a resource
	if statusCode != http.StatusNotFound && diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted a alert resource: %v", data.Id.ValueString()))
}

func (r *alertResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
