package resources

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/samber/lo"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients"
	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/models"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/schemas"
	"github.com/astronomer/astronomer-terraform-provider/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &standardDeploymentResource{}
var _ resource.ResourceWithImportState = &standardDeploymentResource{}
var _ resource.ResourceWithConfigure = &standardDeploymentResource{}
var _ resource.ResourceWithValidateConfig = &standardDeploymentResource{}

func NewStandardDeploymentResource() resource.Resource {
	return &standardDeploymentResource{}
}

// standardDeploymentResource defines the resource implementation.
type standardDeploymentResource struct {
	platformClient *platform.ClientWithResponses
	organizationId string
}

func (r *standardDeploymentResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_standard_deployment"
}

func (r *standardDeploymentResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Standard Deployment resource",
		Attributes:          schemas.StandardDeploymentResourceSchemaAttributes(),
	}
}

func (r *standardDeploymentResource) Configure(
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

func (r *standardDeploymentResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.StandardDeploymentResource

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	deploymentOptions, err := r.platformClient.GetDeploymentOptionsWithResponse(ctx, r.organizationId, &platform.GetDeploymentOptionsParams{
		DeploymentType: lo.ToPtr(platform.GetDeploymentOptionsParamsDeploymentTypeSTANDARD),
		Executor:       lo.ToPtr(platform.GetDeploymentOptionsParamsExecutor(data.Executor.ValueString())),
		CloudProvider:  lo.ToPtr(platform.GetDeploymentOptionsParamsCloudProvider(data.CloudProvider.ValueString())),
	})
	if err != nil {
		tflog.Error(ctx, "failed to get deployment options", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get deployment options for deployment creation, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, deploymentOptions.HTTPResponse, deploymentOptions.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}
	if deploymentOptions.JSON200 == nil || len(deploymentOptions.JSON200.RuntimeReleases) == 0 {
		resp.Diagnostics.AddError(
			"Client Error",
			"Unable to get runtime releases for deployment creation, got empty runtime releases",
		)
		return
	}

	createStandardDeploymentRequest := platform.CreateStandardDeploymentRequest{
		AstroRuntimeVersion:  deploymentOptions.JSON200.RuntimeReleases[0].Version,
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
		WorkloadIdentity:     data.WorkloadIdentity.ValueStringPointer(),
		WorkspaceId:          data.WorkspaceId.ValueString(),
	}

	var diags diag.Diagnostics

	// contact emails
	createStandardDeploymentRequest.ContactEmails, diags = utils.TypesListToStringSlicePtr(ctx, data.ContactEmails)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// env vars
	envVars, diags := RequestDeploymentEnvironmentVariables(ctx, data.EnvironmentVariables)
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

	var createDeploymentRequest platform.CreateDeploymentRequest
	err = createDeploymentRequest.FromCreateStandardDeploymentRequest(createStandardDeploymentRequest)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("failed to create deployment error: %v", err))
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create hybrid deployment request body, got error: %s", err),
		)
		return
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
	_, diagnostic = clients.NormalizeAPIError(ctx, deployment.HTTPResponse, deployment.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	diags = data.ReadFromResponse(ctx, deployment.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created a deployment resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *standardDeploymentResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.StandardDeploymentResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
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

	diags := data.ReadFromResponse(ctx, deployment.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read a standard deployment resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *standardDeploymentResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data models.StandardDeploymentResource
	var prevData models.StandardDeploymentResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &prevData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// update request
	updateStandardDeploymentRequest := platform.UpdateStandardDeploymentRequest{
		DefaultTaskPodCpu:    data.DefaultTaskPodCpu.ValueString(),
		DefaultTaskPodMemory: data.DefaultTaskPodMemory.ValueString(),
		Description:          data.Description.ValueStringPointer(),
		Executor:             platform.UpdateStandardDeploymentRequestExecutor(data.Executor.ValueString()),
		IsCicdEnforced:       data.IsCicdEnforced.ValueBool(),
		IsDagDeployEnabled:   data.IsDagDeployEnabled.ValueBool(),
		// TODO: Uncomment once this https://github.com/astronomer/astro/pull/19471 is merged
		// IsDevelopmentMode:    data.IsDevelopmentMode.ValueBoolPointer(),
		IsHighAvailability:  data.IsHighAvailability.ValueBool(),
		Name:                data.Name.ValueString(),
		ResourceQuotaCpu:    data.ResourceQuotaCpu.ValueString(),
		ResourceQuotaMemory: data.ResourceQuotaMemory.ValueString(),
		SchedulerSize:       platform.UpdateStandardDeploymentRequestSchedulerSize(data.SchedulerSize.ValueString()),
		Type:                platform.UpdateStandardDeploymentRequestTypeSTANDARD,
		WorkloadIdentity:    data.WorkloadIdentity.ValueStringPointer(),
		WorkspaceId:         data.WorkspaceId.ValueString(),
	}

	var diags diag.Diagnostics
	// contact emails
	updateStandardDeploymentRequest.ContactEmails, diags = utils.TypesListToStringSlicePtr(ctx, data.ContactEmails)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// env vars
	updateStandardDeploymentRequest.EnvironmentVariables, diags = RequestDeploymentEnvironmentVariables(ctx, data.EnvironmentVariables)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

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

	var updateDeploymentRequest platform.UpdateDeploymentRequest
	err := updateDeploymentRequest.FromUpdateStandardDeploymentRequest(updateStandardDeploymentRequest)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("failed to update deployment error: %v", err))
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update hybrid deployment request body, got error: %s", err),
		)
		return
	}

	deployment, err := r.platformClient.UpdateDeploymentWithResponse(
		ctx,
		r.organizationId,
		data.Id.ValueString(),
		updateDeploymentRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to update standard deployment", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update standard deployment, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, deployment.HTTPResponse, deployment.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	diags = data.ReadFromResponse(ctx, deployment.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated a standard deployment resource: %v", data.Id.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *standardDeploymentResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.StandardDeploymentResource

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
		tflog.Error(ctx, "failed to delete standard deployment", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete standard deployment, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, deployment.HTTPResponse, deployment.Body)
	// It is recommended to ignore 404 Resource Not Found errors when deleting a resource
	if statusCode != http.StatusNotFound && diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted a standard deployment resource: %v", data.Id.ValueString()))
}

func (r *standardDeploymentResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *standardDeploymentResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var data models.StandardDeploymentResource

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

	// Need to check that is_development_mode is only for small schedulers with high_availability set to false
	if data.IsDevelopmentMode.ValueBool() && (data.SchedulerSize.ValueString() != string(platform.DeploymentSchedulerSizeSMALL) || data.IsHighAvailability.ValueBool()) {
		resp.Diagnostics.AddError(
			"is_development_mode is only supported for small schedulers with high_availability set to false",
			"Either change the scheduler size to 'SMALL' and high_availability to false or set is_development_mode to true",
		)
	}

	// Need to check that scaling_spec is only for is_development_mode set to true
	if !data.IsDevelopmentMode.ValueBool() && !data.ScalingSpec.IsNull() {
		resp.Diagnostics.AddError(
			"scaling_spec (hibernation) is only supported for is_development_mode set to true",
			"Either set is_development_mode to true or remove scaling_spec",
		)
	}
}
