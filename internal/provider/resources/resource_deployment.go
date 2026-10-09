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
	platform_v1 "github.com/astronomer/terraform-provider-astro/internal/clients/platform_v1"
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
	platformV1Client *platform_v1.ClientWithResponses
	platformClient   *platform.ClientWithResponses
	organizationId   string
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

	r.platformV1Client = apiClients.PlatformV1Client
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
	var createDeploymentRequest platform_v1.CreateDeploymentRequest
	var envVars []platform_v1.DeploymentEnvironmentVariableRequest

	originalAstroRuntimeVersion := data.OriginalAstroRuntimeVersion.ValueString()
	if len(originalAstroRuntimeVersion) == 0 {
		var diagnostic diag.Diagnostic
		originalAstroRuntimeVersion, diagnostic = r.GetLatestAstroRuntimeVersion(ctx, &data)
		if diagnostic != nil {
			resp.Diagnostics.Append(diagnostic)
			return

		}
	}

	desiredWorkloadIdentity := data.DesiredWorkloadIdentity.ValueString()

	switch data.Type.ValueString() {
	case string(platform_v1.DeploymentTypeSTANDARD):
		createStandardDeploymentRequest := platform_v1.CreateStandardDeploymentRequest{
			AstroRuntimeVersion:  &originalAstroRuntimeVersion,
			CloudProvider:        (*platform_v1.CreateStandardDeploymentRequestCloudProvider)(data.CloudProvider.ValueStringPointer()),
			DefaultTaskPodCpu:    data.DefaultTaskPodCpu.ValueStringPointer(),
			DefaultTaskPodMemory: data.DefaultTaskPodMemory.ValueStringPointer(),
			Description:          data.Description.ValueStringPointer(),
			Executor:             lo.ToPtr(platform_v1.CreateStandardDeploymentRequestExecutor(data.Executor.ValueString())),
			IsCicdEnforced:       data.IsCicdEnforced.ValueBoolPointer(),
			IsDagDeployEnabled:   data.IsDagDeployEnabled.ValueBoolPointer(),
			IsDevelopmentMode:    data.IsDevelopmentMode.ValueBoolPointer(),
			IsHighAvailability:   data.IsHighAvailability.ValueBoolPointer(),
			Name:                 data.Name.ValueString(),
			Region:               data.Region.ValueStringPointer(),
			ResourceQuotaCpu:     data.ResourceQuotaCpu.ValueStringPointer(),
			ResourceQuotaMemory:  data.ResourceQuotaMemory.ValueStringPointer(),
			SchedulerSize:        lo.ToPtr(platform_v1.CreateStandardDeploymentRequestSchedulerSize(data.SchedulerSize.ValueString())),
			Type:                 lo.ToPtr(platform_v1.CreateStandardDeploymentRequestTypeSTANDARD),
			WorkspaceId:          data.WorkspaceId.ValueString(),
		}
		if desiredWorkloadIdentity != "" {
			createStandardDeploymentRequest.WorkloadIdentity = &desiredWorkloadIdentity
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

	case string(platform_v1.DeploymentTypeDEDICATED):
		createDedicatedDeploymentRequest := platform_v1.CreateDedicatedDeploymentRequest{
			AstroRuntimeVersion:  &originalAstroRuntimeVersion,
			ClusterId:            data.ClusterId.ValueStringPointer(),
			DefaultTaskPodCpu:    data.DefaultTaskPodCpu.ValueStringPointer(),
			DefaultTaskPodMemory: data.DefaultTaskPodMemory.ValueStringPointer(),
			Description:          data.Description.ValueStringPointer(),
			Executor:             lo.ToPtr(platform_v1.CreateDedicatedDeploymentRequestExecutor(data.Executor.ValueString())),
			IsCicdEnforced:       data.IsCicdEnforced.ValueBoolPointer(),
			IsDagDeployEnabled:   data.IsDagDeployEnabled.ValueBoolPointer(),
			IsDevelopmentMode:    data.IsDevelopmentMode.ValueBoolPointer(),
			IsHighAvailability:   data.IsHighAvailability.ValueBoolPointer(),
			Name:                 data.Name.ValueString(),
			ResourceQuotaCpu:     data.ResourceQuotaCpu.ValueStringPointer(),
			ResourceQuotaMemory:  data.ResourceQuotaMemory.ValueStringPointer(),
			SchedulerSize:        lo.ToPtr(platform_v1.CreateDedicatedDeploymentRequestSchedulerSize(data.SchedulerSize.ValueString())),
			Type:                 lo.ToPtr(platform_v1.CreateDedicatedDeploymentRequestTypeDEDICATED),
			WorkspaceId:          data.WorkspaceId.ValueString(),
		}
		if desiredWorkloadIdentity != "" {
			createDedicatedDeploymentRequest.WorkloadIdentity = &desiredWorkloadIdentity
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

		// remote execution
		createDedicatedDeploymentRequest.RemoteExecution, diags = RequestRemoteExecution(ctx, data.RemoteExecution)
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

	case string(platform_v1.DeploymentTypeHYBRID):
		createHybridDeploymentRequest := platform_v1.CreateHybridDeploymentRequest{
			AstroRuntimeVersion: &originalAstroRuntimeVersion,
			ClusterId:           data.ClusterId.ValueStringPointer(),
			Description:         data.Description.ValueStringPointer(),
			Executor:            lo.ToPtr(platform_v1.CreateHybridDeploymentRequestExecutor(data.Executor.ValueString())),
			IsCicdEnforced:      data.IsCicdEnforced.ValueBoolPointer(),
			IsDagDeployEnabled:  data.IsDagDeployEnabled.ValueBoolPointer(),
			Name:                data.Name.ValueString(),
			Scheduler: &platform_v1.CreateDeploymentInstanceSpecRequest{
				Au:       lo.ToPtr(int(data.SchedulerAu.ValueInt64())),
				Replicas: lo.ToPtr(int(data.SchedulerReplicas.ValueInt64())),
			},
			TaskPodNodePoolId: data.TaskPodNodePoolId.ValueStringPointer(),
			Type:              lo.ToPtr(platform_v1.CreateHybridDeploymentRequestTypeHYBRID),
			WorkspaceId:       data.WorkspaceId.ValueString(),
		}

		if desiredWorkloadIdentity != "" {
			createHybridDeploymentRequest.WorkloadIdentity = &desiredWorkloadIdentity
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

	deployment, err := r.platformV1Client.CreateDeploymentWithResponse(
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
	_, diagnostic := clients.NormalizeAPIResponseWithBody(ctx, deployment.HTTPResponse, deployment.Body, deployment.JSON200, "create deployment")
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
	deployment, err := r.platformV1Client.GetDeploymentWithResponse(
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
	statusCode, diagnostic := clients.NormalizeAPIResponseWithBody(ctx, deployment.HTTPResponse, deployment.Body, deployment.JSON200, "read deployment")
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
	var updateDeploymentRequest platform_v1.UpdateDeploymentRequest
	var envVars []platform_v1.DeploymentEnvironmentVariableRequest

	desiredWorkloadIdentity := data.DesiredWorkloadIdentity.ValueString()

	switch data.Type.ValueString() {
	case string(platform_v1.DeploymentTypeSTANDARD):
		updateStandardDeploymentRequest := platform_v1.UpdateStandardDeploymentRequest{
			DefaultTaskPodCpu:    data.DefaultTaskPodCpu.ValueStringPointer(),
			DefaultTaskPodMemory: data.DefaultTaskPodMemory.ValueStringPointer(),
			Description:          data.Description.ValueStringPointer(),
			Executor:             platform_v1.UpdateStandardDeploymentRequestExecutor(data.Executor.ValueString()),
			IsCicdEnforced:       data.IsCicdEnforced.ValueBool(),
			IsDagDeployEnabled:   data.IsDagDeployEnabled.ValueBool(),
			IsDevelopmentMode:    data.IsDevelopmentMode.ValueBoolPointer(),
			IsHighAvailability:   data.IsHighAvailability.ValueBool(),
			Name:                 data.Name.ValueString(),
			ResourceQuotaCpu:     data.ResourceQuotaCpu.ValueStringPointer(),
			ResourceQuotaMemory:  data.ResourceQuotaMemory.ValueStringPointer(),
			SchedulerSize:        platform_v1.UpdateStandardDeploymentRequestSchedulerSize(data.SchedulerSize.ValueString()),
			Type:                 platform_v1.UpdateStandardDeploymentRequestType(platform_v1.DeploymentTypeSTANDARD),
			WorkspaceId:          data.WorkspaceId.ValueString(),
		}

		if desiredWorkloadIdentity != "" {
			updateStandardDeploymentRequest.WorkloadIdentity = &desiredWorkloadIdentity
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
		updateStandardDeploymentRequest.WorkerQueues, diags = RequestHostedUpdateWorkerQueues(ctx, data.WorkerQueues)
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

	case string(platform_v1.DeploymentTypeDEDICATED):
		updateDedicatedDeploymentRequest := platform_v1.UpdateDedicatedDeploymentRequest{
			DefaultTaskPodCpu:    data.DefaultTaskPodCpu.ValueStringPointer(),
			DefaultTaskPodMemory: data.DefaultTaskPodMemory.ValueStringPointer(),
			Description:          data.Description.ValueStringPointer(),
			Executor:             platform_v1.UpdateDedicatedDeploymentRequestExecutor(data.Executor.ValueString()),
			IsCicdEnforced:       data.IsCicdEnforced.ValueBool(),
			IsDagDeployEnabled:   data.IsDagDeployEnabled.ValueBool(),
			IsDevelopmentMode:    data.IsDevelopmentMode.ValueBoolPointer(),
			IsHighAvailability:   data.IsHighAvailability.ValueBool(),
			Name:                 data.Name.ValueString(),
			ResourceQuotaCpu:     data.ResourceQuotaCpu.ValueStringPointer(),
			ResourceQuotaMemory:  data.ResourceQuotaMemory.ValueStringPointer(),
			SchedulerSize:        platform_v1.UpdateDedicatedDeploymentRequestSchedulerSize(data.SchedulerSize.ValueString()),
			Type:                 platform_v1.UpdateDedicatedDeploymentRequestTypeDEDICATED,
			WorkspaceId:          data.WorkspaceId.ValueString(),
		}

		if desiredWorkloadIdentity != "" {
			updateDedicatedDeploymentRequest.WorkloadIdentity = &desiredWorkloadIdentity
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
		updateDedicatedDeploymentRequest.WorkerQueues, diags = RequestHostedUpdateWorkerQueues(ctx, data.WorkerQueues)
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

		// remote execution
		updateDedicatedDeploymentRequest.RemoteExecution, diags = RequestRemoteExecution(ctx, data.RemoteExecution)
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

	case string(platform_v1.DeploymentTypeHYBRID):
		updateHybridDeploymentRequest := platform_v1.UpdateHybridDeploymentRequest{
			Description:        data.Description.ValueStringPointer(),
			Executor:           platform_v1.UpdateHybridDeploymentRequestExecutor(data.Executor.ValueString()),
			IsCicdEnforced:     data.IsCicdEnforced.ValueBool(),
			IsDagDeployEnabled: data.IsDagDeployEnabled.ValueBool(),
			Name:               data.Name.ValueString(),
			Scheduler: platform_v1.UpdateDeploymentInstanceSpecRequest{
				Au:       int(data.SchedulerAu.ValueInt64()),
				Replicas: int(data.SchedulerReplicas.ValueInt64()),
			},
			TaskPodNodePoolId: data.TaskPodNodePoolId.ValueStringPointer(),
			Type:              platform_v1.UpdateHybridDeploymentRequestTypeHYBRID,
			WorkspaceId:       data.WorkspaceId.ValueString(),
		}

		if desiredWorkloadIdentity != "" {
			updateHybridDeploymentRequest.WorkloadIdentity = &desiredWorkloadIdentity
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
		updateHybridDeploymentRequest.WorkerQueues, diags = RequestHybridUpdateWorkerQueues(ctx, data.WorkerQueues)
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

	deployment, err := r.platformV1Client.UpdateDeploymentWithResponse(
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
	_, diagnostic := clients.NormalizeAPIResponseWithBody(ctx, deployment.HTTPResponse, deployment.Body, deployment.JSON200, "update deployment")
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

	deployment, deleteErr := r.platformV1Client.DeleteDeploymentWithResponse(
		ctx,
		r.organizationId,
		data.Id.ValueString(),
	)
	if deleteErr != nil {
		tflog.Error(ctx, "failed to delete deployment", map[string]interface{}{"error": deleteErr})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete deployment, got error: %s", deleteErr),
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

	// Block ASTRO executor for HYBRID deployments
	if data.Executor.ValueString() == string(platform_v1.DeploymentExecutorASTRO) &&
		data.Type.ValueString() == string(platform_v1.DeploymentTypeHYBRID) {
		resp.Diagnostics.AddAttributeError(
			path.Root("executor"),
			"ASTRO executor is not allowed for HYBRID deployments",
			"The 'ASTRO' executor cannot be used with deployment type 'HYBRID'.",
		)
		return
	}

	if data.Executor.ValueString() == string(platform_v1.DeploymentExecutorCELERY) &&
		data.RemoteExecution.IsNull() && !data.WorkerQueues.IsUnknown() &&
		len(data.WorkerQueues.Elements()) == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("worker_queues"),
			"worker_queues is required for 'CELERY' executor",
			"Please provide at least one worker_queue for CELERY executor.",
		)
	}

	// Type specific validation
	switch platform_v1.DeploymentType(data.Type.ValueString()) {
	case platform_v1.DeploymentTypeSTANDARD:
		resp.Diagnostics.Append(validateStandardConfig(ctx, &data)...)
		resp.Diagnostics.Append(validateHostedConfig(ctx, &data)...)
	case platform_v1.DeploymentTypeDEDICATED:
		resp.Diagnostics.Append(validateHostedConfig(ctx, &data)...)
		resp.Diagnostics.Append(validateClusterIdConfig(ctx, &data)...)
	case platform_v1.DeploymentTypeHYBRID:
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
	if !data.RemoteExecution.IsNull() {
		diags.AddError(
			"remote_execution is not allowed for 'HYBRID' deployment",
			"Please remove remote_execution",
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
			// Hybrid worker queue pod sizing comes from the node pool's instance type, and the
			// v1 create payload (HybridWorkerQueueRequest) has no ephemeral storage field.
			if !workerQueue.PodEphemeralStorage.IsNull() && !workerQueue.PodEphemeralStorage.IsUnknown() {
				diags.AddError(
					"pod_ephemeral_storage is not allowed for 'HYBRID' worker_queues",
					"Please remove pod_ephemeral_storage",
				)
			}
		}

		// Check for duplicate worker_queue names
		duplicateWorkerQueueNames := lo.FindDuplicatesBy(workerQueues, func(wq models.WorkerQueueResource) string {
			return wq.Name.ValueString()
		})
		if len(duplicateWorkerQueueNames) > 0 {
			diags.AddError(
				"worker_queue names must be unique",
				fmt.Sprintf("The following worker_queue names are duplicated: %v", duplicateWorkerQueueNames),
			)
		}
	}

	if data.Executor.ValueString() == string(platform_v1.DeploymentExecutorKUBERNETES) && data.TaskPodNodePoolId.IsNull() {
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
	if !data.RemoteExecution.IsNull() {
		diags.AddError(
			"remote_execution is not allowed for 'STANDARD' deployment",
			"Please remove remote_execution",
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
	if data.ResourceQuotaCpu.IsNull() && data.RemoteExecution.IsNull() {
		diags.AddError(
			"resource_quota_cpu is required for 'STANDARD' and 'DEDICATED' deployment",
			"Please provide a resource_quota_cpu",
		)
	}
	if data.ResourceQuotaMemory.IsNull() && data.RemoteExecution.IsNull() {
		diags.AddError(
			"resource_quota_memory is required for 'STANDARD' and 'DEDICATED' deployment",
			"Please provide a resource_quota_memory",
		)
	}
	if data.DefaultTaskPodCpu.IsNull() && data.RemoteExecution.IsNull() {
		diags.AddError(
			"default_task_pod_cpu is required for 'STANDARD' and 'DEDICATED' deployment",
			"Please provide a default_task_pod_cpu",
		)
	}
	if data.DefaultTaskPodMemory.IsNull() && data.RemoteExecution.IsNull() {
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
	if !data.RemoteExecution.IsNull() && data.Executor.ValueString() != string(platform_v1.DeploymentExecutorASTRO) {
		diags.AddError(
			"remote_execution is only allowed for 'ASTRO' executor",
			"Please remove remote_execution or change executor to 'ASTRO'",
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

		// Check for duplicate worker_queue names
		workerQueuesNames := lo.Map(workerQueues, func(wq models.WorkerQueueResource, _ int) string {
			return wq.Name.ValueString()
		})
		duplicateWorkerQueueNames := lo.FindDuplicates(workerQueuesNames)
		if len(duplicateWorkerQueueNames) > 0 {
			diags.AddError(
				"worker_queue names must be unique",
				fmt.Sprintf("The following worker_queue names are duplicated: %v", duplicateWorkerQueueNames),
			)
		}
	}

	// For ASTRO executor without Remote Execution, require at least one worker_queue named 'default'
	if data.Executor.ValueString() == string(platform_v1.DeploymentExecutorASTRO) && data.RemoteExecution.IsNull() {
		// Skip validation if worker_queues is unknown (e.g., provided via a local or variable).
		// The actual value will be validated at plan/apply time when it's resolved.
		if data.WorkerQueues.IsUnknown() {
			return diags
		}
		if len(data.WorkerQueues.Elements()) == 0 {
			diags.AddError(
				"worker_queues is required for 'ASTRO' executor",
				"Please provide at least one worker_queue for ASTRO executor.",
			)
		} else {
			var workerQueues []models.WorkerQueueResource
			diags = append(diags, data.WorkerQueues.ElementsAs(ctx, &workerQueues, false)...)
			foundDefault := false
			for _, wq := range workerQueues {
				if wq.Name.ValueString() == "default" {
					foundDefault = true
					break
				}
			}
			if !foundDefault {
				diags.AddError(
					"'ASTRO' executor requires a worker_queue named 'default'",
					"Please add a worker_queue with name 'default' for ASTRO executor.",
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

// RequestScalingSpec converts a Terraform object to a platform_v1.DeploymentScalingSpecRequest to be used in create and update requests
func RequestScalingSpec(ctx context.Context, scalingSpecObj types.Object) (*platform_v1.DeploymentScalingSpecRequest, diag.Diagnostics) {
	if scalingSpecObj.IsNull() {
		// If the scaling spec is not set, return an empty scaling spec for the request
		return &platform_v1.DeploymentScalingSpecRequest{}, nil
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

	platformScalingSpec := &platform_v1.DeploymentScalingSpecRequest{}
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
	platformScalingSpec.HibernationSpec = &platform_v1.DeploymentHibernationSpecRequest{}

	if hibernationSpec.Override.IsNull() && hibernationSpec.Schedules.IsNull() {
		// If the hibernation spec is set but both override and schedules are not set, return an error
		return platformScalingSpec, diag.Diagnostics{diag.NewErrorDiagnostic("scaling_spec.hibernation_spec must have either override or schedules", "Please provide either override or schedules in 'scaling_spec.hibernation_spec")}
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
		platformScalingSpec.HibernationSpec.Override = &platform_v1.DeploymentHibernationOverrideRequest{
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
		requestSchedules := lo.Map(schedules, func(schedule models.HibernationSchedule, _ int) platform_v1.DeploymentHibernationSchedule {
			return platform_v1.DeploymentHibernationSchedule{
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

// RequestRemoteExecution converts a Terraform object to a platform_v1.RemoteExecutionRequest to be used in create and update requests
func RequestRemoteExecution(ctx context.Context, remoteExecutionObj types.Object) (*platform_v1.DeploymentRemoteExecutionRequest, diag.Diagnostics) {
	if remoteExecutionObj.IsNull() {
		return nil, nil
	}
	var remoteExecution models.RemoteExecution
	diags := remoteExecutionObj.As(ctx, &remoteExecution, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		tflog.Error(ctx, "failed to convert remote execution", map[string]interface{}{"error": diags})
		return nil, diags
	}

	platformRemoteExecution := &platform_v1.DeploymentRemoteExecutionRequest{
		Enabled:           remoteExecution.Enabled.ValueBool(),
		TaskLogBucket:     remoteExecution.TaskLogBucket.ValueStringPointer(),
		TaskLogUrlPattern: remoteExecution.TaskLogUrlPattern.ValueStringPointer(),
	}

	if !remoteExecution.AllowedIpAddressRanges.IsNull() {
		var allowedIpAddressRanges []string
		diags = remoteExecution.AllowedIpAddressRanges.ElementsAs(ctx, &allowedIpAddressRanges, false)
		if diags.HasError() {
			tflog.Error(ctx, "failed to convert allowed IP address ranges", map[string]interface{}{"error": diags})
			return nil, diags
		}
		platformRemoteExecution.AllowedIpAddressRanges = &allowedIpAddressRanges
	}

	return platformRemoteExecution, nil
}

// RequestHostedWorkerQueues converts a Terraform set to a list of platform_v1.WorkerQueueRequest to be used in create and update requests
func RequestHostedWorkerQueues(ctx context.Context, workerQueuesObjSet types.Set) (*[]platform_v1.WorkerQueueRequest, diag.Diagnostics) {
	workerQueues, diags := workerQueueResources(ctx, workerQueuesObjSet)
	if workerQueues == nil || diags.HasError() {
		return nil, diags
	}
	platformWorkerQueues := lo.Map(workerQueues, func(workerQueue models.WorkerQueueResource, _ int) platform_v1.WorkerQueueRequest {
		return platform_v1.WorkerQueueRequest{
			AstroMachine:        platform_v1.WorkerQueueRequestAstroMachine(workerQueue.AstroMachine.ValueString()),
			IsDefault:           workerQueue.IsDefault.ValueBool(),
			MaxWorkerCount:      int(workerQueue.MaxWorkerCount.ValueInt64()),
			MinWorkerCount:      int(workerQueue.MinWorkerCount.ValueInt64()),
			Name:                workerQueue.Name.ValueString(),
			PodEphemeralStorage: configuredPodEphemeralStorage(workerQueue),
			WorkerConcurrency:   int(workerQueue.WorkerConcurrency.ValueInt64()),
		}
	})
	return &platformWorkerQueues, nil
}

// RequestHybridWorkerQueues converts a Terraform set to a list of platform_v1.WorkerQueueRequest to be used in create and update requests
func RequestHybridWorkerQueues(ctx context.Context, workerQueuesObjSet types.Set) (*[]platform_v1.HybridWorkerQueueRequest, diag.Diagnostics) {
	workerQueues, diags := workerQueueResources(ctx, workerQueuesObjSet)
	if workerQueues == nil || diags.HasError() {
		return nil, diags
	}
	platformWorkerQueues := lo.Map(workerQueues, func(workerQueue models.WorkerQueueResource, _ int) platform_v1.HybridWorkerQueueRequest {
		return platform_v1.HybridWorkerQueueRequest{
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

func RequestHostedUpdateWorkerQueues(ctx context.Context, workerQueuesObjSet types.Set) (*[]platform_v1.UpdateWorkerQueueRequest, diag.Diagnostics) {
	workerQueues, diags := workerQueueResources(ctx, workerQueuesObjSet)
	if workerQueues == nil || diags.HasError() {
		return nil, diags
	}
	platformWorkerQueues := lo.Map(workerQueues, func(workerQueue models.WorkerQueueResource, _ int) platform_v1.UpdateWorkerQueueRequest {
		return platform_v1.UpdateWorkerQueueRequest{
			AstroMachine:        lo.ToPtr(platform_v1.UpdateWorkerQueueRequestAstroMachine(workerQueue.AstroMachine.ValueString())),
			IsDefault:           workerQueue.IsDefault.ValueBool(),
			MaxWorkerCount:      int(workerQueue.MaxWorkerCount.ValueInt64()),
			MinWorkerCount:      int(workerQueue.MinWorkerCount.ValueInt64()),
			Name:                workerQueue.Name.ValueString(),
			PodEphemeralStorage: configuredPodEphemeralStorage(workerQueue),
			WorkerConcurrency:   int(workerQueue.WorkerConcurrency.ValueInt64()),
		}
	})
	return &platformWorkerQueues, nil
}

func RequestHybridUpdateWorkerQueues(ctx context.Context, workerQueuesObjSet types.Set) (*[]platform_v1.UpdateWorkerQueueRequest, diag.Diagnostics) {
	workerQueues, diags := workerQueueResources(ctx, workerQueuesObjSet)
	if workerQueues == nil || diags.HasError() {
		return nil, diags
	}
	platformWorkerQueues := lo.Map(workerQueues, func(workerQueue models.WorkerQueueResource, _ int) platform_v1.UpdateWorkerQueueRequest {
		return platform_v1.UpdateWorkerQueueRequest{
			IsDefault:         workerQueue.IsDefault.ValueBool(),
			MaxWorkerCount:    int(workerQueue.MaxWorkerCount.ValueInt64()),
			MinWorkerCount:    int(workerQueue.MinWorkerCount.ValueInt64()),
			Name:              workerQueue.Name.ValueString(),
			NodePoolId:        lo.ToPtr(workerQueue.NodePoolId.ValueString()),
			WorkerConcurrency: int(workerQueue.WorkerConcurrency.ValueInt64()),
		}
	})
	return &platformWorkerQueues, nil
}

// configuredPodEphemeralStorage returns the worker queue's pod_ephemeral_storage only when the
// user actually set it. The attribute is Optional+Computed, so an omitted value arrives as
// unknown on create and as the platform default read back from the API on update; sending
// either would turn a platform-managed default into a value the provider pins.
func configuredPodEphemeralStorage(workerQueue models.WorkerQueueResource) *string {
	if workerQueue.PodEphemeralStorage.IsNull() || workerQueue.PodEphemeralStorage.IsUnknown() {
		return nil
	}
	return workerQueue.PodEphemeralStorage.ValueStringPointer()
}

func workerQueueResources(ctx context.Context, workerQueuesObjSet types.Set) ([]models.WorkerQueueResource, diag.Diagnostics) {
	if len(workerQueuesObjSet.Elements()) == 0 {
		return nil, nil
	}

	var workerQueues []models.WorkerQueueResource
	diags := workerQueuesObjSet.ElementsAs(ctx, &workerQueues, false)
	if diags.HasError() {
		return nil, diags
	}
	return workerQueues, nil
}

// RequestDeploymentEnvironmentVariables converts a Terraform set to a list of platform_v1.DeploymentEnvironmentVariableRequest to be used in create and update requests
func RequestDeploymentEnvironmentVariables(ctx context.Context, environmentVariablesObjSet types.Set) ([]platform_v1.DeploymentEnvironmentVariableRequest, diag.Diagnostics) {
	if len(environmentVariablesObjSet.Elements()) == 0 {
		return []platform_v1.DeploymentEnvironmentVariableRequest{}, nil
	}

	var envVars []models.DeploymentEnvironmentVariable
	diags := environmentVariablesObjSet.ElementsAs(ctx, &envVars, false)
	if diags.HasError() {
		return nil, diags
	}
	platformEnvVars := lo.Map(envVars, func(envVar models.DeploymentEnvironmentVariable, _ int) platform_v1.DeploymentEnvironmentVariableRequest {
		return platform_v1.DeploymentEnvironmentVariableRequest{
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
	_, diagnostic := clients.NormalizeAPIResponseWithBody(ctx, deploymentOptions.HTTPResponse, deploymentOptions.Body, deploymentOptions.JSON200, "read deployment options")
	if diagnostic != nil {
		return "", diagnostic
	}
	if len(deploymentOptions.JSON200.RuntimeReleases) == 0 {
		return "", diag.NewErrorDiagnostic(
			"Client Error",
			"Unable to get runtime releases for deployment creation, got empty runtime releases",
		)
	}

	return LatestStableRuntimeVersion(ctx, deploymentOptions.JSON200.RuntimeReleases), nil
}

// stableRuntimeReleaseChannel is the channel Core marks generally-available runtime releases
// with. Deployment options also advertise pre-release channels (alpha, beta, nightly) that the
// create deployment endpoint then rejects as invalid runtime versions, so taking the first
// release off the list is not safe.
const stableRuntimeReleaseChannel = "stable"

// LatestStableRuntimeVersion picks the newest stable release out of the deployment options.
// The list arrives sorted newest-first, so the first stable entry is the latest one. If no
// release is marked stable the newest release is returned regardless, preserving the previous
// behaviour rather than failing a create that might still have succeeded.
func LatestStableRuntimeVersion(ctx context.Context, runtimeReleases []platform.RuntimeRelease) string {
	for _, runtimeRelease := range runtimeReleases {
		if runtimeRelease.Channel == stableRuntimeReleaseChannel {
			return runtimeRelease.Version
		}
	}

	tflog.Warn(ctx, "no stable Astro Runtime release in deployment options, falling back to the newest release", map[string]interface{}{
		"count":   len(runtimeReleases),
		"version": runtimeReleases[0].Version,
		"channel": runtimeReleases[0].Channel,
	})
	return runtimeReleases[0].Version
}
