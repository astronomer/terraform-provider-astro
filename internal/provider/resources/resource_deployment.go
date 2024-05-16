package resources

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/samber/lo"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DeploymentResource{}
var _ resource.ResourceWithImportState = &DeploymentResource{}
var _ resource.ResourceWithConfigure = &DeploymentResource{}
var _ resource.ResourceWithValidateConfig = &DeploymentResource{}

func NewDeploymentResource() resource.Resource {
	return &DeploymentResource{}
}

// DeploymentResource defines the resource implementation.
type DeploymentResource struct {
	platformClient *platform.ClientWithResponses
	organizationId string
}

func (r *DeploymentResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (r *DeploymentResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Deployment resource",
		Attributes:          schemas.DeploymentResourceSchemaAttributes(),
	}
}

func (r *DeploymentResource) Configure(
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

func (r *DeploymentResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.DeploymentResource

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var diags diag.Diagnostics
	var createDeploymentRequest platform.CreateDeploymentRequest
	var envVars []platform.DeploymentEnvironmentVariableRequest

	originalAstroRuntimeVersion := data.OriginalAstroRuntimeVersion.ValueString()
	if len(originalAstroRuntimeVersion) == 0 {
		var diagnostic diag.Diagnostic
		originalAstroRuntimeVersion, diagnostic = r.GetLatestAstroRuntimeVersion(ctx, &data)
		if diagnostic != nil {
			resp.Diagnostics.Append(diagnostic)
			return

		}
	}

	switch data.Type.ValueString() {
	case string(platform.DeploymentTypeSTANDARD):
		createStandardDeploymentRequest := platform.CreateStandardDeploymentRequest{
			AstroRuntimeVersion:  originalAstroRuntimeVersion,
			CloudProvider:        (*platform.CreateStandardDeploymentRequestCloudProvider)(data.CloudProvider.ValueStringPointer()),
			DefaultTaskPodCpu:    data.DefaultTaskPodCpu.ValueString(),
			DefaultTaskPodMemory: data.DefaultTaskPodMemory.ValueString(),
			Description:          data.Description.ValueStringPointer(),
			Executor:             platform.CreateStandardDeploymentRequestExecutor(data.Executor.ValueString()),
			IsCicdEnforced:       data.IsCicdEnforced.ValueBool(),
			IsDagDeployEnabled:   data.IsDagDeployEnabled.ValueBool(),
			IsDevelopmentMode:    data.IsDevelopmentMode.ValueBoolPointer(),
			IsHighAvailability:   data.IsHighAvailability.ValueBool(),
			Name:                 data.Name.ValueString(),
			Region:               data.Region.ValueStringPointer(),
			ResourceQuotaCpu:     data.ResourceQuotaCpu.ValueString(),
			ResourceQuotaMemory:  data.ResourceQuotaMemory.ValueString(),
			SchedulerSize:        platform.CreateStandardDeploymentRequestSchedulerSize(data.SchedulerSize.ValueString()),
			Type:                 platform.CreateStandardDeploymentRequestTypeSTANDARD,
			WorkspaceId:          data.WorkspaceId.ValueString(),
		}

		// contact emails
		contactEmails, diags := utils.TypesSetToStringSlice(ctx, data.ContactEmails)
		createStandardDeploymentRequest.ContactEmails = &contactEmails
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		// env vars
		envVars, diags = RequestDeploymentEnvironmentVariables(ctx, data.EnvironmentVariables)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		createStandardDeploymentRequest.EnvironmentVariables = &envVars

		// worker queues
		createStandardDeploymentRequest.WorkerQueues, diags = RequestHostedWorkerQueues(ctx, data.WorkerQueues)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		// scaling spec
		createStandardDeploymentRequest.ScalingSpec, diags = RequestScalingSpec(ctx, data.ScalingSpec)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		err := createDeploymentRequest.FromCreateStandardDeploymentRequest(createStandardDeploymentRequest)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("failed to create standard deployment error: %v", err))
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to create standard deployment request body, got error: %s", err),
			)
			return
		}

	case string(platform.DeploymentTypeDEDICATED):
		createDedicatedDeploymentRequest := platform.CreateDedicatedDeploymentRequest{
			AstroRuntimeVersion:  originalAstroRuntimeVersion,
			ClusterId:            data.ClusterId.ValueString(),
			DefaultTaskPodCpu:    data.DefaultTaskPodCpu.ValueString(),
			DefaultTaskPodMemory: data.DefaultTaskPodMemory.ValueString(),
			Description:          data.Description.ValueStringPointer(),
			Executor:             platform.CreateDedicatedDeploymentRequestExecutor(data.Executor.ValueString()),
			IsCicdEnforced:       data.IsCicdEnforced.ValueBool(),
			IsDagDeployEnabled:   data.IsDagDeployEnabled.ValueBool(),
			IsDevelopmentMode:    data.IsDevelopmentMode.ValueBoolPointer(),
			IsHighAvailability:   data.IsHighAvailability.ValueBool(),
			Name:                 data.Name.ValueString(),
			ResourceQuotaCpu:     data.ResourceQuotaCpu.ValueString(),
			ResourceQuotaMemory:  data.ResourceQuotaMemory.ValueString(),
			SchedulerSize:        platform.CreateDedicatedDeploymentRequestSchedulerSize(data.SchedulerSize.ValueString()),
			Type:                 platform.CreateDedicatedDeploymentRequestTypeDEDICATED,
			WorkspaceId:          data.WorkspaceId.ValueString(),
		}

		// contact emails
		contactEmails, diags := utils.TypesSetToStringSlice(ctx, data.ContactEmails)
		createDedicatedDeploymentRequest.ContactEmails = &contactEmails
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		// env vars
		envVars, diags = RequestDeploymentEnvironmentVariables(ctx, data.EnvironmentVariables)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		createDedicatedDeploymentRequest.EnvironmentVariables = &envVars

		// worker queues
		createDedicatedDeploymentRequest.WorkerQueues, diags = RequestHostedWorkerQueues(ctx, data.WorkerQueues)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		// scaling spec
		createDedicatedDeploymentRequest.ScalingSpec, diags = RequestScalingSpec(ctx, data.ScalingSpec)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		err := createDeploymentRequest.FromCreateDedicatedDeploymentRequest(createDedicatedDeploymentRequest)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("failed to create dedicated deployment error: %v", err))
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to create dedicated deployment request body, got error: %s", err),
			)
			return
		}

	case string(platform.DeploymentTypeHYBRID):
		createHybridDeploymentRequest := platform.CreateHybridDeploymentRequest{
			AstroRuntimeVersion: originalAstroRuntimeVersion,
			ClusterId:           data.ClusterId.ValueString(),
			Description:         data.Description.ValueStringPointer(),
			Executor:            platform.CreateHybridDeploymentRequestExecutor(data.Executor.ValueString()),
			IsCicdEnforced:      data.IsCicdEnforced.ValueBool(),
			IsDagDeployEnabled:  data.IsDagDeployEnabled.ValueBool(),
			Name:                data.Name.ValueString(),
			Scheduler: platform.DeploymentInstanceSpecRequest{
				Au:       int(data.SchedulerAu.ValueInt64()),
				Replicas: int(data.SchedulerReplicas.ValueInt64()),
			},
			TaskPodNodePoolId: data.TaskPodNodePoolId.ValueStringPointer(),
			Type:              platform.CreateHybridDeploymentRequestTypeHYBRID,
			WorkspaceId:       data.WorkspaceId.ValueString(),
		}

		// contact emails
		contactEmails, diags := utils.TypesSetToStringSlice(ctx, data.ContactEmails)
		createHybridDeploymentRequest.ContactEmails = &contactEmails
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		// env vars
		envVars, diags = RequestDeploymentEnvironmentVariables(ctx, data.EnvironmentVariables)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		createHybridDeploymentRequest.EnvironmentVariables = &envVars

		// worker queues
		createHybridDeploymentRequest.WorkerQueues, diags = RequestHybridWorkerQueues(ctx, data.WorkerQueues)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		err := createDeploymentRequest.FromCreateHybridDeploymentRequest(createHybridDeploymentRequest)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("failed to create hybrid deployment error: %v", err))
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to create hybrid deployment request body, got error: %s", err),
			)
			return
		}
	}

	deployment, err := r.platformClient.CreateDeploymentWithResponse(
		ctx,
		r.organizationId,
		createDeploymentRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to create deployment", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create deployment, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, deployment.HTTPResponse, deployment.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	diags = data.ReadFromResponse(ctx, deployment.JSON200, data.OriginalAstroRuntimeVersion.ValueStringPointer(), &envVars)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created a deployment resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.DeploymentResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	envVars, diags := RequestDeploymentEnvironmentVariables(ctx, data.EnvironmentVariables)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// get request
	deployment, err := r.platformClient.GetDeploymentWithResponse(
		ctx,
		r.organizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to get deployment", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get deployment, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, deployment.HTTPResponse, deployment.Body)
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

	diags = data.ReadFromResponse(ctx, deployment.JSON200, data.OriginalAstroRuntimeVersion.ValueStringPointer(), &envVars)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read a deployment resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data models.DeploymentResource

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// update request
	diags := make(diag.Diagnostics, 0)
	var updateDeploymentRequest platform.UpdateDeploymentRequest
	var envVars []platform.DeploymentEnvironmentVariableRequest

	switch data.Type.ValueString() {
	case string(platform.DeploymentTypeSTANDARD):
		updateStandardDeploymentRequest := platform.UpdateStandardDeploymentRequest{
			DefaultTaskPodCpu:    data.DefaultTaskPodCpu.ValueString(),
			DefaultTaskPodMemory: data.DefaultTaskPodMemory.ValueString(),
			Description:          data.Description.ValueStringPointer(),
			Executor:             platform.UpdateStandardDeploymentRequestExecutor(data.Executor.ValueString()),
			IsCicdEnforced:       data.IsCicdEnforced.ValueBool(),
			IsDagDeployEnabled:   data.IsDagDeployEnabled.ValueBool(),
			IsDevelopmentMode:    data.IsDevelopmentMode.ValueBoolPointer(),
			IsHighAvailability:   data.IsHighAvailability.ValueBool(),
			Name:                 data.Name.ValueString(),
			ResourceQuotaCpu:     data.ResourceQuotaCpu.ValueString(),
			ResourceQuotaMemory:  data.ResourceQuotaMemory.ValueString(),
			SchedulerSize:        platform.UpdateStandardDeploymentRequestSchedulerSize(data.SchedulerSize.ValueString()),
			Type:                 platform.UpdateStandardDeploymentRequestTypeSTANDARD,
			WorkspaceId:          data.WorkspaceId.ValueString(),
		}

		// contact emails
		contactEmails, diags := utils.TypesSetToStringSlice(ctx, data.ContactEmails)
		updateStandardDeploymentRequest.ContactEmails = &contactEmails
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		// env vars
		envVars, diags = RequestDeploymentEnvironmentVariables(ctx, data.EnvironmentVariables)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		updateStandardDeploymentRequest.EnvironmentVariables = envVars

		// worker queues
		updateStandardDeploymentRequest.WorkerQueues, diags = RequestHostedWorkerQueues(ctx, data.WorkerQueues)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		// scaling spec
		updateStandardDeploymentRequest.ScalingSpec, diags = RequestScalingSpec(ctx, data.ScalingSpec)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		err := updateDeploymentRequest.FromUpdateStandardDeploymentRequest(updateStandardDeploymentRequest)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("failed to update standard deployment error: %v", err))
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to update standard deployment request body, got error: %s", err),
			)
			return
		}

	case string(platform.DeploymentTypeDEDICATED):
		updateDedicatedDeploymentRequest := platform.UpdateDedicatedDeploymentRequest{
			DefaultTaskPodCpu:    data.DefaultTaskPodCpu.ValueString(),
			DefaultTaskPodMemory: data.DefaultTaskPodMemory.ValueString(),
			Description:          data.Description.ValueStringPointer(),
			Executor:             platform.UpdateDedicatedDeploymentRequestExecutor(data.Executor.ValueString()),
			IsCicdEnforced:       data.IsCicdEnforced.ValueBool(),
			IsDagDeployEnabled:   data.IsDagDeployEnabled.ValueBool(),
			IsDevelopmentMode:    data.IsDevelopmentMode.ValueBoolPointer(),
			IsHighAvailability:   data.IsHighAvailability.ValueBool(),
			Name:                 data.Name.ValueString(),
			ResourceQuotaCpu:     data.ResourceQuotaCpu.ValueString(),
			ResourceQuotaMemory:  data.ResourceQuotaMemory.ValueString(),
			SchedulerSize:        platform.UpdateDedicatedDeploymentRequestSchedulerSize(data.SchedulerSize.ValueString()),
			Type:                 platform.UpdateDedicatedDeploymentRequestTypeDEDICATED,
			WorkspaceId:          data.WorkspaceId.ValueString(),
		}

		// contact emails
		contactEmails, diags := utils.TypesSetToStringSlice(ctx, data.ContactEmails)
		updateDedicatedDeploymentRequest.ContactEmails = &contactEmails
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		// env vars
		envVars, diags = RequestDeploymentEnvironmentVariables(ctx, data.EnvironmentVariables)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		updateDedicatedDeploymentRequest.EnvironmentVariables = envVars

		// worker queues
		updateDedicatedDeploymentRequest.WorkerQueues, diags = RequestHostedWorkerQueues(ctx, data.WorkerQueues)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		// scaling spec
		updateDedicatedDeploymentRequest.ScalingSpec, diags = RequestScalingSpec(ctx, data.ScalingSpec)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		err := updateDeploymentRequest.FromUpdateDedicatedDeploymentRequest(updateDedicatedDeploymentRequest)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("failed to update dedicated deployment error: %v", err))
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to update dedicated deployment request body, got error: %s", err),
			)
			return
		}

	case string(platform.DeploymentTypeHYBRID):
		updateHybridDeploymentRequest := platform.UpdateHybridDeploymentRequest{
			Description:        data.Description.ValueStringPointer(),
			Executor:           platform.UpdateHybridDeploymentRequestExecutor(data.Executor.ValueString()),
			IsCicdEnforced:     data.IsCicdEnforced.ValueBool(),
			IsDagDeployEnabled: data.IsDagDeployEnabled.ValueBool(),
			Name:               data.Name.ValueString(),
			Scheduler: platform.DeploymentInstanceSpecRequest{
				Au:       int(data.SchedulerAu.ValueInt64()),
				Replicas: int(data.SchedulerReplicas.ValueInt64()),
			},
			TaskPodNodePoolId: data.TaskPodNodePoolId.ValueStringPointer(),
			Type:              platform.UpdateHybridDeploymentRequestTypeHYBRID,
			WorkspaceId:       data.WorkspaceId.ValueString(),
		}

		// contact emails
		contactEmails, diags := utils.TypesSetToStringSlice(ctx, data.ContactEmails)
		updateHybridDeploymentRequest.ContactEmails = &contactEmails
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		// env vars
		envVars, diags = RequestDeploymentEnvironmentVariables(ctx, data.EnvironmentVariables)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		updateHybridDeploymentRequest.EnvironmentVariables = envVars

		// worker queues
		updateHybridDeploymentRequest.WorkerQueues, diags = RequestHybridWorkerQueues(ctx, data.WorkerQueues)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		err := updateDeploymentRequest.FromUpdateHybridDeploymentRequest(updateHybridDeploymentRequest)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("failed to create hybrid deployment error: %v", err))
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to create hybrid deployment request body, got error: %s", err),
			)
			return
		}
	}

	deployment, err := r.platformClient.UpdateDeploymentWithResponse(
		ctx,
		r.organizationId,
		data.Id.ValueString(),
		updateDeploymentRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to update deployment", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update deployment, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, deployment.HTTPResponse, deployment.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	diags = data.ReadFromResponse(ctx, deployment.JSON200, data.OriginalAstroRuntimeVersion.ValueStringPointer(), &envVars)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated a deployment resource: %v", data.Id.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.DeploymentResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// delete request
	deployment, err := r.platformClient.DeleteDeploymentWithResponse(
		ctx,
		r.organizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to delete deployment", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete deployment, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, deployment.HTTPResponse, deployment.Body)
	// It is recommended to ignore 404 Resource Not Found errors when deleting a resource
	if statusCode != http.StatusNotFound && diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted a deployment resource: %v", data.Id.ValueString()))
}

func (r *DeploymentResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ValidateConfig validates the configuration of the resource as a whole before any operations are performed.
// This is a good place to check for any conflicting settings.
func (r *DeploymentResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var data models.DeploymentResource

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Need to do dynamic validation based on the executor and worker queues
	if data.Executor.ValueString() == string(platform.DeploymentExecutorKUBERNETES) && len(data.WorkerQueues.Elements()) > 0 {
		resp.Diagnostics.AddError(
			"worker_queues are not supported for 'KUBERNETES' executor",
			"Either change the executor to 'CELERY' or remove worker_queues",
		)
	}
	if data.Executor.ValueString() == string(platform.DeploymentExecutorCELERY) && (data.WorkerQueues.IsNull() || len(data.WorkerQueues.Elements()) == 0) {
		resp.Diagnostics.AddError(
			"worker_queues must be included for 'CELERY' executor",
			"Either change the executor to 'KUBERNETES' or include worker_queues",
		)
	}

	// Type specific validation
	switch platform.DeploymentType(data.Type.ValueString()) {
	case platform.DeploymentTypeSTANDARD:
		resp.Diagnostics.Append(validateStandardConfig(ctx, &data)...)
		resp.Diagnostics.Append(validateHostedConfig(ctx, &data)...)
	case platform.DeploymentTypeDEDICATED:
		resp.Diagnostics.Append(validateHostedConfig(ctx, &data)...)
		resp.Diagnostics.Append(validateClusterIdConfig(ctx, &data)...)
	case platform.DeploymentTypeHYBRID:
		resp.Diagnostics.Append(validateHybridConfig(ctx, &data)...)
		resp.Diagnostics.Append(validateClusterIdConfig(ctx, &data)...)
	}
}

func validateHybridConfig(ctx context.Context, data *models.DeploymentResource) diag.Diagnostics {
	diags := make(diag.Diagnostics, 0)
	// Required hybrid values
	if data.SchedulerAu.IsNull() {
		diags.AddError(
			"scheduler_au is required for 'HYBRID' deployment",
			"Please provide a scheduler_au",
		)
	}
	if data.SchedulerReplicas.IsNull() {
		diags.AddError(
			"scheduler_replicas is required for 'HYBRID' deployment",
			"Please provide a scheduler_replicas",
		)
	}

	// Unallowed values
	if !data.SchedulerSize.IsNull() {
		diags.AddError(
			"scheduler_size is not allowed for 'HYBRID' deployment",
			"Please remove scheduler_size",
		)
	}
	if !data.ScalingSpec.IsNull() {
		diags.AddError(
			"scaling_spec is not allowed for 'HYBRID' deployment",
			"Please remove scaling_spec",
		)
	}
	if !data.IsDevelopmentMode.IsNull() {
		diags.AddError(
			"is_development_mode is not allowed for 'HYBRID' deployment",
			"Please remove is_development_mode",
		)
	}
	if !data.IsHighAvailability.IsNull() {
		diags.AddError(
			"is_high_availability is not allowed for 'HYBRID' deployment",
			"Please remove is_high_availability",
		)
	}
	if !data.ResourceQuotaCpu.IsNull() {
		diags.AddError(
			"resource_quota_cpu is not allowed for 'HYBRID' deployment",
			"Please remove resource_quota_cpu",
		)
	}
	if !data.ResourceQuotaMemory.IsNull() {
		diags.AddError(
			"resource_quota_memory is not allowed for 'HYBRID' deployment",
			"Please remove resource_quota_memory",
		)
	}
	if !data.DefaultTaskPodCpu.IsNull() {
		diags.AddError(
			"default_task_pod_cpu is not allowed for 'HYBRID' deployment",
			"Please remove default_task_pod_cpu",
		)
	}
	if !data.DefaultTaskPodMemory.IsNull() {
		diags.AddError(
			"default_task_pod_memory is not allowed for 'HYBRID' deployment",
			"Please remove default_task_pod_memory",
		)
	}

	// Need to check worker_queues for hybrid deployments have `node_pool_id` and do not have `astro_machine`
	if len(data.WorkerQueues.Elements()) > 0 {
		var workerQueues []models.WorkerQueueResource
		diags = append(diags, data.WorkerQueues.ElementsAs(ctx, &workerQueues, false)...)
		for _, workerQueue := range workerQueues {
			if !workerQueue.AstroMachine.IsNull() {
				diags.AddError(
					"astro_machine is not allowed for 'HYBRID' worker_queues",
					"Please remove astro_machine",
				)
			}
			if workerQueue.NodePoolId.IsNull() {
				diags.AddError(
					"node_pool_id is required for 'HYBRID' worker_queues",
					"Please provide a node_pool_id",
				)
			}
		}
	}

	if data.Executor.ValueString() == string(platform.DeploymentExecutorKUBERNETES) && data.TaskPodNodePoolId.IsNull() {
		diags.AddError(
			"task_node_pool_id is required for 'KUBERNETES' executor in 'HYBRID' deployment",
			"Please provide a task_node_pool_id",
		)
	}

	return diags
}

func validateStandardConfig(ctx context.Context, data *models.DeploymentResource) diag.Diagnostics {
	diags := make(diag.Diagnostics, 0)
	// Required standard values
	if data.Region.IsNull() {
		diags.AddError(
			"region is required for 'STANDARD' deployment",
			"Please provide a region",
		)
	}
	if data.CloudProvider.IsNull() {
		diags.AddError(
			"cloud_provider is required for 'STANDARD' deployment",
			"Please provide a cloud_provider",
		)
	}

	// Unallowed values
	if !data.ClusterId.IsNull() {
		diags.AddError(
			"cluster_id is not allowed for 'STANDARD' deployment",
			"Please remove cluster_id",
		)
	}
	return diags
}

func validateHostedConfig(ctx context.Context, data *models.DeploymentResource) diag.Diagnostics {
	// Required hosted values
	diags := make(diag.Diagnostics, 0)
	if data.SchedulerSize.IsNull() {
		diags.AddError(
			"scheduler_size is required for 'STANDARD' and 'DEDICATED' deployment",
			"Please provide a scheduler_size",
		)
	}
	if data.IsHighAvailability.IsNull() {
		diags.AddError(
			"is_high_availability is required for 'STANDARD' and 'DEDICATED' deployment",
			"Please provide is_high_availability",
		)
	}
	if data.IsDevelopmentMode.IsNull() {
		diags.AddError(
			"is_development_mode is required for 'STANDARD' and 'DEDICATED' deployment",
			"Please provide is_development_mode",
		)
	}
	if data.ResourceQuotaCpu.IsNull() {
		diags.AddError(
			"resource_quota_cpu is required for 'STANDARD' and 'DEDICATED' deployment",
			"Please provide a resource_quota_cpu",
		)
	}
	if data.ResourceQuotaMemory.IsNull() {
		diags.AddError(
			"resource_quota_memory is required for 'STANDARD' and 'DEDICATED' deployment",
			"Please provide a resource_quota_memory",
		)
	}
	if data.DefaultTaskPodCpu.IsNull() {
		diags.AddError(
			"default_task_pod_cpu is required for 'STANDARD' and 'DEDICATED' deployment",
			"Please provide a default_task_pod_cpu",
		)
	}
	if data.DefaultTaskPodMemory.IsNull() {
		diags.AddError(
			"default_task_pod_memory is required for 'STANDARD' and 'DEDICATED' deployment",
			"Please provide a default_task_pod_memory",
		)
	}

	// Unallowed values
	if !data.SchedulerAu.IsNull() {
		diags.AddError(
			"scheduler_au is not allowed for 'STANDARD' and 'DEDICATED' deployment",
			"Please remove scheduler_au",
		)
	}
	if !data.SchedulerReplicas.IsNull() {
		diags.AddError(
			"scheduler_replicas is not allowed for 'STANDARD' and 'DEDICATED' deployment",
			"Please remove scheduler_replicas",
		)
	}
	if !data.TaskPodNodePoolId.IsNull() {
		diags.AddError(
			"task_node_pool_id is not allowed for 'STANDARD' and 'DEDICATED' deployment",
			"Please remove task_node_pool_id",
		)
	}

	// Need to check that is_development_mode is only for small schedulers with high_availability set to false
	if data.IsDevelopmentMode.ValueBool() && (data.SchedulerSize.ValueString() != string(platform.DeploymentSchedulerSizeSMALL) || data.IsHighAvailability.ValueBool()) {
		diags.AddError(
			"is_development_mode is only supported for small schedulers with high_availability set to false",
			"Either change the scheduler size to 'SMALL' and high_availability to false or set is_development_mode to true",
		)
	}

	// Need to check that scaling_spec is only for is_development_mode set to true
	if !data.IsDevelopmentMode.ValueBool() && !data.ScalingSpec.IsNull() {
		diags.AddError(
			"scaling_spec (hibernation) is only supported for is_development_mode set to true",
			"Either set is_development_mode to true or remove scaling_spec",
		)
	}

	// Need to check that scaling_spec has either override or schedules
	if !data.ScalingSpec.IsNull() {
		var scalingSpec models.DeploymentScalingSpec
		diags = append(diags, data.ScalingSpec.As(ctx, &scalingSpec, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})...)
		if diags.HasError() {
			tflog.Error(ctx, "failed to convert scaling spec", map[string]interface{}{"error": diags})
			return diags
		}

		// scalingSpec.HibernationSpec is required if ScalingSpec is set via schemas/deployment.go
		var hibernationSpec models.HibernationSpec
		diags = scalingSpec.HibernationSpec.As(ctx, &hibernationSpec, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diags.HasError() {
			tflog.Error(ctx, "failed to convert hibernation spec", map[string]interface{}{"error": diags})
			return diags
		}
		if hibernationSpec.Override.IsNull() && hibernationSpec.Schedules.IsNull() {
			diags.AddError(
				"scaling_spec (hibernation) must have either override or schedules",
				"Please provide either override or schedules in 'scaling_spec.hibernation_spec'",
			)
			return diags
		}
	}

	// Need to check worker_queues for hosted deployments have `astro_machine` and do not have `node_pool_id`
	if len(data.WorkerQueues.Elements()) > 0 {
		var workerQueues []models.WorkerQueueResource
		diags = append(diags, data.WorkerQueues.ElementsAs(ctx, &workerQueues, false)...)
		for _, workerQueue := range workerQueues {
			if workerQueue.AstroMachine.IsNull() {
				diags.AddError(
					"astro_machine is required for 'STANDARD' and 'DEDICATED' worker_queues",
					"Please provide an astro_machine",
				)
			}
			if !workerQueue.NodePoolId.IsNull() {
				diags.AddError(
					"node_pool_id is not allowed for 'STANDARD' and 'DEDICATED' worker_queues",
					"Please remove node_pool_id",
				)
			}
		}
	}

	return diags
}

func validateClusterIdConfig(ctx context.Context, data *models.DeploymentResource) diag.Diagnostics {
	diags := make(diag.Diagnostics, 0)
	// Required clusterId value
	if data.ClusterId.IsNull() {
		diags.AddError(
			"cluster_id is required for 'DEDICATED' and 'HYBRID' deployment",
			"Please provide a cluster_id",
		)
	}

	// Unallowed values
	if !data.CloudProvider.IsNull() {
		diags.AddError(
			"cloud_provider is not allowed for 'DEDICATED' and 'HYBRID' deployment",
			"Please remove cloud_provider",
		)
	}
	if !data.Region.IsNull() {
		diags.AddError(
			"region is not allowed for 'DEDICATED' and 'HYBRID' deployment",
			"Please remove region",
		)
	}
	return diags
}

// RequestScalingSpec converts a Terraform object to a platform.DeploymentScalingSpecRequest to be used in create and update requests
func RequestScalingSpec(ctx context.Context, scalingSpecObj types.Object) (*platform.DeploymentScalingSpecRequest, diag.Diagnostics) {
	if scalingSpecObj.IsNull() {
		// If the scaling spec is not set, return nil for the request
		return nil, nil
	}
	var scalingSpec models.DeploymentScalingSpec
	diags := scalingSpecObj.As(ctx, &scalingSpec, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		tflog.Error(ctx, "failed to convert scaling spec", map[string]interface{}{"error": diags})
		return nil, diags
	}

	platformScalingSpec := &platform.DeploymentScalingSpecRequest{}
	if scalingSpec.HibernationSpec.IsNull() {
		// If the hibernation spec is not set, return a scaling spec without hibernation spec for the request
		platformScalingSpec.HibernationSpec = nil
		return platformScalingSpec, nil
	}
	var hibernationSpec models.HibernationSpec
	diags = scalingSpec.HibernationSpec.As(ctx, &hibernationSpec, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		tflog.Error(ctx, "failed to convert hibernation spec", map[string]interface{}{"error": diags})
		return nil, diags
	}
	platformScalingSpec.HibernationSpec = &platform.DeploymentHibernationSpecRequest{}

	if hibernationSpec.Override.IsNull() && hibernationSpec.Schedules.IsNull() {
		// If the hibernation spec is set but both override and schedules are not set, return an empty hibernation spec for the request
		return platformScalingSpec, nil
	}
	if !hibernationSpec.Override.IsNull() {
		var override models.HibernationSpecOverride
		diags = hibernationSpec.Override.As(ctx, &override, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diags.HasError() {
			tflog.Error(ctx, "failed to convert hibernation override", map[string]interface{}{"error": diags})
			return nil, diags
		}
		platformScalingSpec.HibernationSpec.Override = &platform.DeploymentHibernationOverrideRequest{
			IsHibernating: override.IsHibernating.ValueBoolPointer(),
			OverrideUntil: override.OverrideUntil.ValueStringPointer(),
		}
	}
	if !hibernationSpec.Schedules.IsNull() {
		var schedules []models.HibernationSchedule
		diags = hibernationSpec.Schedules.ElementsAs(ctx, &schedules, false)
		if diags.HasError() {
			tflog.Error(ctx, "failed to convert hibernation schedules", map[string]interface{}{"error": diags})
			return nil, diags
		}
		requestSchedules := lo.Map(schedules, func(schedule models.HibernationSchedule, _ int) platform.DeploymentHibernationSchedule {
			return platform.DeploymentHibernationSchedule{
				Description:     schedule.Description.ValueStringPointer(),
				HibernateAtCron: schedule.HibernateAtCron.ValueString(),
				IsEnabled:       schedule.IsEnabled.ValueBool(),
				WakeAtCron:      schedule.WakeAtCron.ValueString(),
			}
		})
		platformScalingSpec.HibernationSpec.Schedules = &requestSchedules
	}

	return platformScalingSpec, nil
}

// RequestHostedWorkerQueues converts a Terraform set to a list of platform.WorkerQueueRequest to be used in create and update requests
func RequestHostedWorkerQueues(ctx context.Context, workerQueuesObjSet types.Set) (*[]platform.WorkerQueueRequest, diag.Diagnostics) {
	if len(workerQueuesObjSet.Elements()) == 0 {
		return nil, nil
	}

	var workerQueues []models.WorkerQueueResource
	diags := workerQueuesObjSet.ElementsAs(ctx, &workerQueues, false)
	if diags.HasError() {
		return nil, diags
	}
	platformWorkerQueues := lo.Map(workerQueues, func(workerQueue models.WorkerQueueResource, _ int) platform.WorkerQueueRequest {
		return platform.WorkerQueueRequest{
			AstroMachine:      platform.WorkerQueueRequestAstroMachine(workerQueue.AstroMachine.ValueString()),
			IsDefault:         workerQueue.IsDefault.ValueBool(),
			MaxWorkerCount:    int(workerQueue.MaxWorkerCount.ValueInt64()),
			MinWorkerCount:    int(workerQueue.MinWorkerCount.ValueInt64()),
			Name:              workerQueue.Name.ValueString(),
			WorkerConcurrency: int(workerQueue.WorkerConcurrency.ValueInt64()),
		}
	})
	return &platformWorkerQueues, nil
}

// RequestHybridWorkerQueues converts a Terraform set to a list of platform.WorkerQueueRequest to be used in create and update requests
func RequestHybridWorkerQueues(ctx context.Context, workerQueuesObjSet types.Set) (*[]platform.HybridWorkerQueueRequest, diag.Diagnostics) {
	if len(workerQueuesObjSet.Elements()) == 0 {
		return nil, nil
	}

	var workerQueues []models.WorkerQueueResource
	diags := workerQueuesObjSet.ElementsAs(ctx, &workerQueues, false)
	if diags.HasError() {
		return nil, diags
	}
	platformWorkerQueues := lo.Map(workerQueues, func(workerQueue models.WorkerQueueResource, _ int) platform.HybridWorkerQueueRequest {
		return platform.HybridWorkerQueueRequest{
			IsDefault:         workerQueue.IsDefault.ValueBool(),
			MaxWorkerCount:    int(workerQueue.MaxWorkerCount.ValueInt64()),
			MinWorkerCount:    int(workerQueue.MinWorkerCount.ValueInt64()),
			Name:              workerQueue.Name.ValueString(),
			NodePoolId:        workerQueue.NodePoolId.ValueString(),
			WorkerConcurrency: int(workerQueue.WorkerConcurrency.ValueInt64()),
		}
	})
	return &platformWorkerQueues, nil
}

// RequestDeploymentEnvironmentVariables converts a Terraform set to a list of platform.DeploymentEnvironmentVariableRequest to be used in create and update requests
func RequestDeploymentEnvironmentVariables(ctx context.Context, environmentVariablesObjSet types.Set) ([]platform.DeploymentEnvironmentVariableRequest, diag.Diagnostics) {
	if len(environmentVariablesObjSet.Elements()) == 0 {
		return []platform.DeploymentEnvironmentVariableRequest{}, nil
	}

	var envVars []models.DeploymentEnvironmentVariable
	diags := environmentVariablesObjSet.ElementsAs(ctx, &envVars, false)
	if diags.HasError() {
		return nil, diags
	}
	platformEnvVars := lo.Map(envVars, func(envVar models.DeploymentEnvironmentVariable, _ int) platform.DeploymentEnvironmentVariableRequest {
		return platform.DeploymentEnvironmentVariableRequest{
			IsSecret: envVar.IsSecret.ValueBool(),
			Key:      envVar.Key.ValueString(),
			Value:    envVar.Value.ValueStringPointer(),
		}
	})
	return platformEnvVars, nil
}

func (r *DeploymentResource) GetLatestAstroRuntimeVersion(ctx context.Context, data *models.DeploymentResource) (string, diag.Diagnostic) {
	deploymentOptions, err := r.platformClient.GetDeploymentOptionsWithResponse(ctx, r.organizationId, &platform.GetDeploymentOptionsParams{
		DeploymentType: lo.ToPtr(platform.GetDeploymentOptionsParamsDeploymentType(data.Type.ValueString())),
		Executor:       lo.ToPtr(platform.GetDeploymentOptionsParamsExecutor(data.Executor.ValueString())),
		CloudProvider:  lo.ToPtr(platform.GetDeploymentOptionsParamsCloudProvider(data.CloudProvider.ValueString())),
	})
	if err != nil {
		tflog.Error(ctx, "failed to get deployment options", map[string]interface{}{"error": err})
		return "", diag.NewErrorDiagnostic(
			"Client Error",
			fmt.Sprintf("Unable to get deployment options for deployment creation, got error: %s", err),
		)
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, deploymentOptions.HTTPResponse, deploymentOptions.Body)
	if diagnostic != nil {
		return "", diagnostic
	}
	if deploymentOptions.JSON200 == nil || len(deploymentOptions.JSON200.RuntimeReleases) == 0 {
		return "", diag.NewErrorDiagnostic(
			"Client Error",
			"Unable to get runtime releases for deployment creation, got empty runtime releases",
		)
	}
	return deploymentOptions.JSON200.RuntimeReleases[0].Version, nil
}
