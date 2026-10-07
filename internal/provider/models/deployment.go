package models

import (
	"context"
	"time"

	"github.com/samber/lo"

	platform_v1 "github.com/astronomer/terraform-provider-astro/internal/clients/platform_v1"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type DeploymentResource struct {
	// Common fields
	Id                          types.String `tfsdk:"id"`
	Name                        types.String `tfsdk:"name"`
	Description                 types.String `tfsdk:"description"`
	CreatedAt                   types.String `tfsdk:"created_at"`
	UpdatedAt                   types.String `tfsdk:"updated_at"`
	CreatedBy                   types.Object `tfsdk:"created_by"`
	UpdatedBy                   types.Object `tfsdk:"updated_by"`
	WorkspaceId                 types.String `tfsdk:"workspace_id"`
	Type                        types.String `tfsdk:"type"`
	Region                      types.String `tfsdk:"region"`
	CloudProvider               types.String `tfsdk:"cloud_provider"`
	OriginalAstroRuntimeVersion types.String `tfsdk:"original_astro_runtime_version"`
	AstroRuntimeVersion         types.String `tfsdk:"astro_runtime_version"`
	AirflowVersion              types.String `tfsdk:"airflow_version"`
	Namespace                   types.String `tfsdk:"namespace"`
	ContactEmails               types.Set    `tfsdk:"contact_emails"`
	Executor                    types.String `tfsdk:"executor"`
	SchedulerCpu                types.String `tfsdk:"scheduler_cpu"`
	SchedulerMemory             types.String `tfsdk:"scheduler_memory"`
	SchedulerAu                 types.Int64  `tfsdk:"scheduler_au"`
	SchedulerReplicas           types.Int64  `tfsdk:"scheduler_replicas"`
	ImageTag                    types.String `tfsdk:"image_tag"`
	ImageRepository             types.String `tfsdk:"image_repository"`
	ImageVersion                types.String `tfsdk:"image_version"`
	EnvironmentVariables        types.Set    `tfsdk:"environment_variables"`
	WebserverIngressHostname    types.String `tfsdk:"webserver_ingress_hostname"`
	WebserverUrl                types.String `tfsdk:"webserver_url"`
	WebserverAirflowApiUrl      types.String `tfsdk:"webserver_airflow_api_url"`
	Status                      types.String `tfsdk:"status"`
	StatusReason                types.String `tfsdk:"status_reason"`
	DagTarballVersion           types.String `tfsdk:"dag_tarball_version"`
	DesiredDagTarballVersion    types.String `tfsdk:"desired_dag_tarball_version"`
	IsCicdEnforced              types.Bool   `tfsdk:"is_cicd_enforced"`
	IsDagDeployEnabled          types.Bool   `tfsdk:"is_dag_deploy_enabled"`
	DesiredWorkloadIdentity     types.String `tfsdk:"desired_workload_identity"`
	WorkloadIdentity            types.String `tfsdk:"workload_identity"`
	ExternalIps                 types.Set    `tfsdk:"external_ips"`
	OidcIssuerUrl               types.String `tfsdk:"oidc_issuer_url"`
	WorkerQueues                types.Set    `tfsdk:"worker_queues"`

	// Hybrid and dedicated specific fields
	ClusterId types.String `tfsdk:"cluster_id"`

	// Hybrid deployment specific fields
	TaskPodNodePoolId types.String `tfsdk:"task_pod_node_pool_id"`

	// Hosted (standard and dedicated) deployment specific fields
	ResourceQuotaCpu     types.String `tfsdk:"resource_quota_cpu"`
	ResourceQuotaMemory  types.String `tfsdk:"resource_quota_memory"`
	DefaultTaskPodCpu    types.String `tfsdk:"default_task_pod_cpu"`
	DefaultTaskPodMemory types.String `tfsdk:"default_task_pod_memory"`
	SchedulerSize        types.String `tfsdk:"scheduler_size"`
	IsDevelopmentMode    types.Bool   `tfsdk:"is_development_mode"`
	IsHighAvailability   types.Bool   `tfsdk:"is_high_availability"`
	ScalingStatus        types.Object `tfsdk:"scaling_status"`
	ScalingSpec          types.Object `tfsdk:"scaling_spec"`
	RemoteExecution      types.Object `tfsdk:"remote_execution"`
}

type DeploymentDataSource struct {
	// Common fields
	Id                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	CreatedAt                types.String `tfsdk:"created_at"`
	UpdatedAt                types.String `tfsdk:"updated_at"`
	CreatedBy                types.Object `tfsdk:"created_by"`
	UpdatedBy                types.Object `tfsdk:"updated_by"`
	WorkspaceId              types.String `tfsdk:"workspace_id"`
	Type                     types.String `tfsdk:"type"`
	Region                   types.String `tfsdk:"region"`
	CloudProvider            types.String `tfsdk:"cloud_provider"`
	AstroRuntimeVersion      types.String `tfsdk:"astro_runtime_version"`
	AirflowVersion           types.String `tfsdk:"airflow_version"`
	Namespace                types.String `tfsdk:"namespace"`
	ContactEmails            types.Set    `tfsdk:"contact_emails"`
	Executor                 types.String `tfsdk:"executor"`
	SchedulerCpu             types.String `tfsdk:"scheduler_cpu"`
	SchedulerMemory          types.String `tfsdk:"scheduler_memory"`
	SchedulerAu              types.Int64  `tfsdk:"scheduler_au"`
	SchedulerReplicas        types.Int64  `tfsdk:"scheduler_replicas"`
	ImageTag                 types.String `tfsdk:"image_tag"`
	ImageRepository          types.String `tfsdk:"image_repository"`
	ImageVersion             types.String `tfsdk:"image_version"`
	EnvironmentVariables     types.Set    `tfsdk:"environment_variables"`
	WebserverIngressHostname types.String `tfsdk:"webserver_ingress_hostname"`
	WebserverUrl             types.String `tfsdk:"webserver_url"`
	WebserverAirflowApiUrl   types.String `tfsdk:"webserver_airflow_api_url"`
	Status                   types.String `tfsdk:"status"`
	StatusReason             types.String `tfsdk:"status_reason"`
	DagTarballVersion        types.String `tfsdk:"dag_tarball_version"`
	DesiredDagTarballVersion types.String `tfsdk:"desired_dag_tarball_version"`
	IsCicdEnforced           types.Bool   `tfsdk:"is_cicd_enforced"`
	IsDagDeployEnabled       types.Bool   `tfsdk:"is_dag_deploy_enabled"`
	WorkloadIdentity         types.String `tfsdk:"workload_identity"`
	ExternalIps              types.Set    `tfsdk:"external_ips"`
	OidcIssuerUrl            types.String `tfsdk:"oidc_issuer_url"`
	WorkerQueues             types.Set    `tfsdk:"worker_queues"`

	// Hybrid and dedicated specific fields
	ClusterId types.String `tfsdk:"cluster_id"`

	// Hybrid deployment specific fields
	TaskPodNodePoolId types.String `tfsdk:"task_pod_node_pool_id"`

	// Hosted deployment specific fields
	ResourceQuotaCpu     types.String `tfsdk:"resource_quota_cpu"`
	ResourceQuotaMemory  types.String `tfsdk:"resource_quota_memory"`
	DefaultTaskPodCpu    types.String `tfsdk:"default_task_pod_cpu"`
	DefaultTaskPodMemory types.String `tfsdk:"default_task_pod_memory"`
	ScalingStatus        types.Object `tfsdk:"scaling_status"`
	ScalingSpec          types.Object `tfsdk:"scaling_spec"`
	RemoteExecution      types.Object `tfsdk:"remote_execution"`
	SchedulerSize        types.String `tfsdk:"scheduler_size"`
	IsDevelopmentMode    types.Bool   `tfsdk:"is_development_mode"`
	IsHighAvailability   types.Bool   `tfsdk:"is_high_availability"`
}

func (data *DeploymentResource) ReadFromResponse(
	ctx context.Context,
	deployment *platform_v1.Deployment,
	originalAstroRuntimeVersion *string,
	requestEnvVars *[]platform_v1.DeploymentEnvironmentVariableRequest,
) diag.Diagnostics {
	// Read common fields
	data.Id = types.StringValue(deployment.Id)
	data.Name = types.StringValue(deployment.Name)
	// If the description is nil, set it to an empty string since the terraform state/config for this resource
	// cannot have a null value for a string.
	if deployment.Description != nil {
		data.Description = types.StringValue(*deployment.Description)
	} else {
		data.Description = types.StringValue("")
	}
	data.CreatedAt = types.StringValue(deployment.CreatedAt.String())
	data.UpdatedAt = types.StringValue(deployment.UpdatedAt.String())
	var diags diag.Diagnostics
	data.CreatedBy, diags = SubjectProfileTypesObject(ctx, deployment.CreatedBy)
	if diags.HasError() {
		return diags
	}
	data.UpdatedBy, diags = SubjectProfileTypesObject(ctx, deployment.UpdatedBy)
	if diags.HasError() {
		return diags
	}
	data.WorkspaceId = types.StringValue(deployment.WorkspaceId)
	data.Type = types.StringPointerValue((*string)(deployment.Type))
	data.Region = types.StringPointerValue(deployment.Region)
	data.CloudProvider = types.StringPointerValue((*string)(deployment.CloudProvider))

	// OriginalAstroRuntimeVersion is the version of the Astro runtime that was used to create the deployment
	// If originalAstroRuntimeVersion is nil (e.g., during import), fall back to the current AstroRuntimeVersion
	if originalAstroRuntimeVersion == nil || *originalAstroRuntimeVersion == "" {
		data.OriginalAstroRuntimeVersion = types.StringValue(deployment.AstroRuntimeVersion)
	} else {
		data.OriginalAstroRuntimeVersion = types.StringPointerValue(originalAstroRuntimeVersion)
	}
	data.AstroRuntimeVersion = types.StringValue(deployment.AstroRuntimeVersion)

	data.AirflowVersion = types.StringValue(deployment.AirflowVersion)
	data.Namespace = types.StringValue(deployment.Namespace)
	data.ContactEmails, diags = utils.StringSet(deployment.ContactEmails)
	if diags.HasError() {
		return diags
	}
	data.Executor = types.StringPointerValue((*string)(deployment.Executor))
	data.SchedulerCpu = types.StringValue(deployment.SchedulerCpu)
	data.SchedulerMemory = types.StringValue(deployment.SchedulerMemory)
	if deployment.SchedulerAu != nil {
		deploymentSchedulerAu := int64(*deployment.SchedulerAu)
		data.SchedulerAu = types.Int64Value(deploymentSchedulerAu)
	} else {
		// scheduler_au is Optional+Computed - when the API doesn't return a value (e.g. STANDARD/DEDICATED
		// deployments), it must be explicitly set to null. Otherwise it's left as whatever the prior
		// plan/state value was (often unknown), which Terraform rejects as an invalid apply result.
		data.SchedulerAu = types.Int64Null()
	}
	data.SchedulerReplicas = types.Int64Value(int64(deployment.SchedulerReplicas))
	data.ImageTag = types.StringValue(deployment.ImageTag)
	data.ImageRepository = types.StringValue(deployment.ImageRepository)
	data.ImageVersion = types.StringPointerValue(deployment.ImageVersion)

	// Environment variables are a special case
	// Since terraform wants to know the values of the secret values in the request at all times, and our API does not send back the secret values in the response
	// We must use the request value and set it in the Terraform response to keep Terraform from emitting errors
	// Since the value is marked as sensitive, Terraform will not output the actual value in the plan/apply output
	var envVarsPtr *[]platform_v1.DeploymentEnvironmentVariable
	if deployment.EnvironmentVariables != nil {
		envVars := *deployment.EnvironmentVariables
		if requestEnvVars != nil {
			requestEnvVarsMap := lo.SliceToMap(*requestEnvVars, func(envVar platform_v1.DeploymentEnvironmentVariableRequest) (string, platform_v1.DeploymentEnvironmentVariable) {
				return envVar.Key, platform_v1.DeploymentEnvironmentVariable{
					Key:      envVar.Key,
					Value:    envVar.Value,
					IsSecret: envVar.IsSecret,
				}
			})
			for i, envVar := range envVars {
				if envVar.IsSecret {
					if requestEnvVar, ok := requestEnvVarsMap[envVar.Key]; ok {
						// If the envVar has a secret value, update the value in the response
						envVars[i].Value = requestEnvVar.Value
					}
				}
			}
		}
		envVarsPtr = &envVars
	}
	data.EnvironmentVariables, diags = utils.ObjectSet(ctx, envVarsPtr, schemas.DeploymentEnvironmentVariableAttributeTypes(), DeploymentEnvironmentVariableTypesObject)
	if diags.HasError() {
		return diags
	}
	data.WebserverIngressHostname = types.StringValue(deployment.WebServerIngressHostname)
	data.WebserverUrl = types.StringValue(deployment.WebServerUrl)
	data.WebserverAirflowApiUrl = types.StringValue(deployment.WebServerAirflowApiUrl)
	data.Status = types.StringValue(string(deployment.Status))
	data.StatusReason = types.StringPointerValue(deployment.StatusReason)
	data.DagTarballVersion = types.StringPointerValue(deployment.DagTarballVersion)
	data.DesiredDagTarballVersion = types.StringPointerValue(deployment.DesiredDagTarballVersion)
	data.IsCicdEnforced = types.BoolValue(deployment.IsCicdEnforced)
	data.IsDagDeployEnabled = types.BoolValue(deployment.IsDagDeployEnabled)
	data.WorkloadIdentity = types.StringPointerValue(deployment.EffectiveWorkloadIdentity)
	data.ExternalIps, diags = utils.StringSet(deployment.ExternalIPs)
	if diags.HasError() {
		return diags
	}
	data.OidcIssuerUrl = types.StringPointerValue(deployment.OidcIssuerUrl)
	data.WorkerQueues, diags = utils.ObjectSet(ctx, deployment.WorkerQueues, schemas.WorkerQueueResourceAttributeTypes(), WorkerQueueResourceTypesObject)
	if diags.HasError() {
		return diags
	}

	// Read hybrid and dedicated specific fields
	data.ClusterId = types.StringPointerValue(deployment.ClusterId)

	// Read hybrid deployment specific fields
	data.TaskPodNodePoolId = types.StringPointerValue(deployment.TaskPodNodePoolId)

	// Read hosted (standard and dedicated) deployment specific fields
	data.ResourceQuotaCpu = types.StringPointerValue(deployment.ResourceQuotaCpu)
	data.ResourceQuotaMemory = types.StringPointerValue(deployment.ResourceQuotaMemory)
	data.DefaultTaskPodCpu = types.StringPointerValue(deployment.DefaultTaskPodCpu)
	data.DefaultTaskPodMemory = types.StringPointerValue(deployment.DefaultTaskPodMemory)
	data.SchedulerSize = types.StringPointerValue((*string)(deployment.SchedulerSize))
	data.IsDevelopmentMode = types.BoolPointerValue(deployment.IsDevelopmentMode)
	data.IsHighAvailability = types.BoolPointerValue(deployment.IsHighAvailability)
	data.ScalingStatus, diags = ScalingStatusTypesObject(ctx, deployment.ScalingStatus)
	if diags.HasError() {
		return diags
	}
	configuredOverrideUntil := ConfiguredOverrideUntil(ctx, data.ScalingSpec)
	data.ScalingSpec, diags = ScalingSpecTypesObject(ctx, deployment.ScalingSpec, configuredOverrideUntil)
	if diags.HasError() {
		return diags
	}
	data.RemoteExecution, diags = RemoteExecutionTypesObject(ctx, deployment.RemoteExecution)
	if diags.HasError() {
		return diags
	}

	return nil
}

func (data *DeploymentDataSource) ReadFromResponse(
	ctx context.Context,
	deployment *platform_v1.Deployment,
) diag.Diagnostics {
	// Read common fields
	data.Id = types.StringValue(deployment.Id)
	data.Name = types.StringValue(deployment.Name)
	// If the description is nil, set it to an empty string since the terraform state/config for this resource
	// cannot have a null value for a string.
	if deployment.Description != nil {
		data.Description = types.StringValue(*deployment.Description)
	} else {
		data.Description = types.StringValue("")
	}
	data.CreatedAt = types.StringValue(deployment.CreatedAt.String())
	data.UpdatedAt = types.StringValue(deployment.UpdatedAt.String())
	var diags diag.Diagnostics
	data.CreatedBy, diags = SubjectProfileTypesObject(ctx, deployment.CreatedBy)
	if diags.HasError() {
		return diags
	}
	data.UpdatedBy, diags = SubjectProfileTypesObject(ctx, deployment.UpdatedBy)
	if diags.HasError() {
		return diags
	}
	data.WorkspaceId = types.StringValue(deployment.WorkspaceId)
	data.Type = types.StringPointerValue((*string)(deployment.Type))
	data.Region = types.StringPointerValue(deployment.Region)
	data.CloudProvider = types.StringPointerValue((*string)(deployment.CloudProvider))
	data.AstroRuntimeVersion = types.StringValue(deployment.AstroRuntimeVersion)
	data.AirflowVersion = types.StringValue(deployment.AirflowVersion)
	data.Namespace = types.StringValue(deployment.Namespace)
	data.ContactEmails, diags = utils.StringSet(deployment.ContactEmails)
	if diags.HasError() {
		return diags
	}
	data.Executor = types.StringPointerValue((*string)(deployment.Executor))
	data.SchedulerCpu = types.StringValue(deployment.SchedulerCpu)
	data.SchedulerMemory = types.StringValue(deployment.SchedulerMemory)
	if deployment.SchedulerAu != nil {
		deploymentSchedulerAu := int64(*deployment.SchedulerAu)
		data.SchedulerAu = types.Int64Value(deploymentSchedulerAu)
	} else {
		data.SchedulerAu = types.Int64Null()
	}
	data.SchedulerReplicas = types.Int64Value(int64(deployment.SchedulerReplicas))
	data.ImageTag = types.StringValue(deployment.ImageTag)
	data.ImageRepository = types.StringValue(deployment.ImageRepository)
	data.ImageVersion = types.StringPointerValue(deployment.ImageVersion)
	data.EnvironmentVariables, diags = utils.ObjectSet(ctx, deployment.EnvironmentVariables, schemas.DeploymentEnvironmentVariableAttributeTypes(), DeploymentEnvironmentVariableTypesObject)
	if diags.HasError() {
		return diags
	}
	data.WebserverIngressHostname = types.StringValue(deployment.WebServerIngressHostname)
	data.WebserverUrl = types.StringValue(deployment.WebServerUrl)
	data.WebserverAirflowApiUrl = types.StringValue(deployment.WebServerAirflowApiUrl)
	data.Status = types.StringValue(string(deployment.Status))
	data.StatusReason = types.StringPointerValue(deployment.StatusReason)
	data.DagTarballVersion = types.StringPointerValue(deployment.DagTarballVersion)
	data.DesiredDagTarballVersion = types.StringPointerValue(deployment.DesiredDagTarballVersion)
	data.IsCicdEnforced = types.BoolValue(deployment.IsCicdEnforced)
	data.IsDagDeployEnabled = types.BoolValue(deployment.IsDagDeployEnabled)
	data.WorkloadIdentity = types.StringPointerValue(deployment.EffectiveWorkloadIdentity)
	data.ExternalIps, diags = utils.StringSet(deployment.ExternalIPs)
	if diags.HasError() {
		return diags
	}
	data.OidcIssuerUrl = types.StringPointerValue(deployment.OidcIssuerUrl)
	data.WorkerQueues, diags = utils.ObjectSet(ctx, deployment.WorkerQueues, schemas.WorkerQueueDataSourceAttributeTypes(), WorkerQueueDataSourceTypesObject)
	if diags.HasError() {
		return diags
	}

	// Read hybrid and dedicated specific fields
	data.ClusterId = types.StringPointerValue(deployment.ClusterId)

	// Read hybrid deployment specific fields
	data.TaskPodNodePoolId = types.StringPointerValue(deployment.TaskPodNodePoolId)

	// Read hosted (standard and dedicated) deployment specific fields
	data.ResourceQuotaCpu = types.StringPointerValue(deployment.ResourceQuotaCpu)
	data.ResourceQuotaMemory = types.StringPointerValue(deployment.ResourceQuotaMemory)
	data.DefaultTaskPodCpu = types.StringPointerValue(deployment.DefaultTaskPodCpu)
	data.DefaultTaskPodMemory = types.StringPointerValue(deployment.DefaultTaskPodMemory)
	data.SchedulerSize = types.StringPointerValue((*string)(deployment.SchedulerSize))
	data.IsDevelopmentMode = types.BoolPointerValue(deployment.IsDevelopmentMode)
	data.IsHighAvailability = types.BoolPointerValue(deployment.IsHighAvailability)
	data.ScalingStatus, diags = ScalingStatusTypesObject(ctx, deployment.ScalingStatus)
	if diags.HasError() {
		return diags
	}
	data.ScalingSpec, diags = ScalingSpecTypesObject(ctx, deployment.ScalingSpec, types.StringNull())
	if diags.HasError() {
		return diags
	}
	data.RemoteExecution, diags = RemoteExecutionTypesObject(ctx, deployment.RemoteExecution)
	if diags.HasError() {
		return diags
	}

	return nil
}

type DeploymentEnvironmentVariable struct {
	Key       types.String `tfsdk:"key"`
	Value     types.String `tfsdk:"value"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	IsSecret  types.Bool   `tfsdk:"is_secret"`
}

type WorkerQueueDataSource struct {
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	AstroMachine      types.String `tfsdk:"astro_machine"`
	IsDefault         types.Bool   `tfsdk:"is_default"`
	MaxWorkerCount    types.Int64  `tfsdk:"max_worker_count"`
	MinWorkerCount    types.Int64  `tfsdk:"min_worker_count"`
	NodePoolId        types.String `tfsdk:"node_pool_id"`
	PodCpu            types.String `tfsdk:"pod_cpu"`
	PodMemory         types.String `tfsdk:"pod_memory"`
	WorkerConcurrency types.Int64  `tfsdk:"worker_concurrency"`
}

type WorkerQueueResource struct {
	Name              types.String `tfsdk:"name"`
	AstroMachine      types.String `tfsdk:"astro_machine"`
	IsDefault         types.Bool   `tfsdk:"is_default"`
	MaxWorkerCount    types.Int64  `tfsdk:"max_worker_count"`
	MinWorkerCount    types.Int64  `tfsdk:"min_worker_count"`
	NodePoolId        types.String `tfsdk:"node_pool_id"`
	PodCpu            types.String `tfsdk:"pod_cpu"`
	PodMemory         types.String `tfsdk:"pod_memory"`
	WorkerConcurrency types.Int64  `tfsdk:"worker_concurrency"`
}

func DeploymentEnvironmentVariableTypesObject(
	ctx context.Context,
	envVar platform_v1.DeploymentEnvironmentVariable,
) (types.Object, diag.Diagnostics) {
	obj := DeploymentEnvironmentVariable{
		Key:       types.StringValue(envVar.Key),
		Value:     types.StringPointerValue(envVar.Value),
		UpdatedAt: types.StringValue(envVar.UpdatedAt.String()),
		IsSecret:  types.BoolValue(envVar.IsSecret),
	}

	return types.ObjectValueFrom(ctx, schemas.DeploymentEnvironmentVariableAttributeTypes(), obj)
}

func WorkerQueueResourceTypesObject(
	ctx context.Context,
	workerQueue platform_v1.WorkerQueue,
) (types.Object, diag.Diagnostics) {
	obj := WorkerQueueResource{
		Name:              types.StringValue(workerQueue.Name),
		AstroMachine:      types.StringPointerValue(workerQueue.AstroMachine),
		IsDefault:         types.BoolValue(workerQueue.IsDefault),
		MaxWorkerCount:    types.Int64Value(int64(workerQueue.MaxWorkerCount)),
		MinWorkerCount:    types.Int64Value(int64(workerQueue.MinWorkerCount)),
		NodePoolId:        types.StringPointerValue(workerQueue.NodePoolId),
		PodCpu:            types.StringValue(workerQueue.PodCpu),
		PodMemory:         types.StringValue(workerQueue.PodMemory),
		WorkerConcurrency: types.Int64Value(int64(workerQueue.WorkerConcurrency)),
	}

	return types.ObjectValueFrom(ctx, schemas.WorkerQueueResourceAttributeTypes(), obj)
}

func WorkerQueueDataSourceTypesObject(
	ctx context.Context,
	workerQueue platform_v1.WorkerQueue,
) (types.Object, diag.Diagnostics) {
	obj := WorkerQueueDataSource{
		Id:                types.StringValue(workerQueue.Id),
		Name:              types.StringValue(workerQueue.Name),
		AstroMachine:      types.StringPointerValue(workerQueue.AstroMachine),
		IsDefault:         types.BoolValue(workerQueue.IsDefault),
		MaxWorkerCount:    types.Int64Value(int64(workerQueue.MaxWorkerCount)),
		MinWorkerCount:    types.Int64Value(int64(workerQueue.MinWorkerCount)),
		NodePoolId:        types.StringPointerValue(workerQueue.NodePoolId),
		PodCpu:            types.StringValue(workerQueue.PodCpu),
		PodMemory:         types.StringValue(workerQueue.PodMemory),
		WorkerConcurrency: types.Int64Value(int64(workerQueue.WorkerConcurrency)),
	}

	return types.ObjectValueFrom(ctx, schemas.WorkerQueueDataSourceAttributeTypes(), obj)
}

type DeploymentScalingSpec struct {
	HibernationSpec types.Object `tfsdk:"hibernation_spec"`
}

type DeploymentStatus struct {
	HibernationStatus types.Object `tfsdk:"hibernation_status"`
}

type HibernationStatus struct {
	IsHibernating types.Bool   `tfsdk:"is_hibernating"`
	NextEventType types.String `tfsdk:"next_event_type"`
	NextEventAt   types.String `tfsdk:"next_event_at"`
	Reason        types.String `tfsdk:"reason"`
}

type HibernationSpec struct {
	Override  types.Object `tfsdk:"override"`
	Schedules types.Set    `tfsdk:"schedules"`
}

type HibernationSpecOverride struct {
	IsHibernating types.Bool   `tfsdk:"is_hibernating"`
	OverrideUntil types.String `tfsdk:"override_until"`
	IsActive      types.Bool   `tfsdk:"is_active"`
}

type HibernationSchedule struct {
	Description     types.String `tfsdk:"description"`
	HibernateAtCron types.String `tfsdk:"hibernate_at_cron"`
	IsEnabled       types.Bool   `tfsdk:"is_enabled"`
	WakeAtCron      types.String `tfsdk:"wake_at_cron"`
}

func HibernationStatusTypesObject(
	ctx context.Context,
	hibernationStatus *platform_v1.DeploymentHibernationStatus,
) (types.Object, diag.Diagnostics) {
	if hibernationStatus == nil {
		return types.ObjectNull(schemas.HibernationStatusAttributeTypes()), nil
	}

	obj := HibernationStatus{
		IsHibernating: types.BoolValue(hibernationStatus.IsHibernating),
		NextEventType: types.StringPointerValue((*string)(hibernationStatus.NextEventType)),
		NextEventAt:   types.StringPointerValue(hibernationStatus.NextEventAt),
		Reason:        types.StringPointerValue(hibernationStatus.Reason),
	}
	return types.ObjectValueFrom(ctx, schemas.HibernationStatusAttributeTypes(), obj)
}

func HibernationOverrideTypesObject(
	ctx context.Context,
	hibernationOverride *platform_v1.DeploymentHibernationOverride,
	configuredOverrideUntil types.String,
) (types.Object, diag.Diagnostics) {
	if hibernationOverride == nil {
		return types.ObjectNull(schemas.HibernationOverrideAttributeTypes()), nil
	}
	obj := HibernationSpecOverride{
		IsHibernating: types.BoolPointerValue(hibernationOverride.IsHibernating),
		IsActive:      types.BoolPointerValue(hibernationOverride.IsActive),
	}
	if hibernationOverride.OverrideUntil != nil {
		obj.OverrideUntil = types.StringValue(hibernationOverride.OverrideUntil.Format(time.RFC3339))

		// The API normalizes override_until to UTC, so a value configured in another offset
		// comes back as a different string for the same instant ("2075-04-25T12:58:00+05:30"
		// -> "2075-04-25T07:28:00Z"). override_until is Optional, so Terraform compares the
		// applied string against the configured one and rejects the apply as an inconsistent
		// result. Keep the configured spelling whenever it denotes the same instant.
		if !configuredOverrideUntil.IsNull() && !configuredOverrideUntil.IsUnknown() {
			configured, err := time.Parse(time.RFC3339, configuredOverrideUntil.ValueString())
			if err == nil && configured.Equal(*hibernationOverride.OverrideUntil) {
				obj.OverrideUntil = configuredOverrideUntil
			}
		}
	}
	return types.ObjectValueFrom(ctx, schemas.HibernationOverrideAttributeTypes(), obj)
}

// ConfiguredOverrideUntil digs scaling_spec.hibernation_spec.override.override_until out of a
// scaling spec object as it was supplied in configuration, so the API response can be
// reconciled against it. Returns a null string when any level is absent.
func ConfiguredOverrideUntil(ctx context.Context, scalingSpecObj types.Object) types.String {
	objectAsOptions := basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	}

	if scalingSpecObj.IsNull() || scalingSpecObj.IsUnknown() {
		return types.StringNull()
	}
	var scalingSpec DeploymentScalingSpec
	if diags := scalingSpecObj.As(ctx, &scalingSpec, objectAsOptions); diags.HasError() {
		return types.StringNull()
	}

	if scalingSpec.HibernationSpec.IsNull() || scalingSpec.HibernationSpec.IsUnknown() {
		return types.StringNull()
	}
	var hibernationSpec HibernationSpec
	if diags := scalingSpec.HibernationSpec.As(ctx, &hibernationSpec, objectAsOptions); diags.HasError() {
		return types.StringNull()
	}

	if hibernationSpec.Override.IsNull() || hibernationSpec.Override.IsUnknown() {
		return types.StringNull()
	}
	var override HibernationSpecOverride
	if diags := hibernationSpec.Override.As(ctx, &override, objectAsOptions); diags.HasError() {
		return types.StringNull()
	}

	return override.OverrideUntil
}

func HibernationScheduleTypesObject(
	ctx context.Context,
	schedule platform_v1.DeploymentHibernationSchedule,
) (types.Object, diag.Diagnostics) {
	obj := HibernationSchedule{
		Description:     types.StringPointerValue(schedule.Description),
		HibernateAtCron: types.StringValue(schedule.HibernateAtCron),
		IsEnabled:       types.BoolValue(schedule.IsEnabled),
		WakeAtCron:      types.StringValue(schedule.WakeAtCron),
	}
	return types.ObjectValueFrom(ctx, schemas.HibernationScheduleAttributeTypes(), obj)
}

func HibernationSpecTypesObject(
	ctx context.Context,
	hibernationSpec *platform_v1.DeploymentHibernationSpec,
	configuredOverrideUntil types.String,
) (types.Object, diag.Diagnostics) {
	if hibernationSpec == nil || (hibernationSpec.Override == nil && hibernationSpec.Schedules == nil) {
		return types.ObjectNull(schemas.HibernationSpecAttributeTypes()), nil
	}

	override, diags := HibernationOverrideTypesObject(ctx, hibernationSpec.Override, configuredOverrideUntil)
	if diags.HasError() {
		return types.ObjectNull(schemas.HibernationSpecAttributeTypes()), diags
	}
	schedules, diags := utils.ObjectSet(ctx, hibernationSpec.Schedules, schemas.HibernationScheduleAttributeTypes(), HibernationScheduleTypesObject)
	if diags.HasError() {
		return types.ObjectNull(schemas.HibernationSpecAttributeTypes()), diags
	}
	obj := HibernationSpec{
		Override:  override,
		Schedules: schedules,
	}
	return types.ObjectValueFrom(ctx, schemas.HibernationSpecAttributeTypes(), obj)
}

func ScalingStatusTypesObject(
	ctx context.Context,
	scalingStatus *platform_v1.DeploymentScalingStatus,
) (types.Object, diag.Diagnostics) {
	if scalingStatus == nil {
		return types.ObjectNull(schemas.ScalingStatusAttributeTypes()), nil
	}

	hibernationStatus, diags := HibernationStatusTypesObject(ctx, scalingStatus.HibernationStatus)
	if diags.HasError() {
		return types.ObjectNull(schemas.ScalingStatusAttributeTypes()), diags
	}
	obj := DeploymentStatus{
		HibernationStatus: hibernationStatus,
	}
	return types.ObjectValueFrom(ctx, schemas.ScalingStatusAttributeTypes(), obj)
}

func ScalingSpecTypesObject(
	ctx context.Context,
	scalingSpec *platform_v1.DeploymentScalingSpec,
	configuredOverrideUntil types.String,
) (types.Object, diag.Diagnostics) {
	if scalingSpec == nil {
		return types.ObjectNull(schemas.ScalingSpecAttributeTypes()), nil
	}

	hibernationSpec, diags := HibernationSpecTypesObject(ctx, scalingSpec.HibernationSpec, configuredOverrideUntil)
	if diags.HasError() {
		return types.ObjectNull(schemas.ScalingSpecAttributeTypes()), diags
	}
	obj := DeploymentScalingSpec{
		HibernationSpec: hibernationSpec,
	}
	return types.ObjectValueFrom(ctx, schemas.ScalingSpecAttributeTypes(), obj)
}

type RemoteExecution struct {
	Enabled                types.Bool   `tfsdk:"enabled"`
	AllowedIpAddressRanges types.Set    `tfsdk:"allowed_ip_address_ranges"`
	RemoteApiUrl           types.String `tfsdk:"remote_api_url"`
	TaskLogBucket          types.String `tfsdk:"task_log_bucket"`
	TaskLogUrlPattern      types.String `tfsdk:"task_log_url_pattern"`
}

func RemoteExecutionTypesObject(
	ctx context.Context,
	remoteExecution *platform_v1.DeploymentRemoteExecution,
) (types.Object, diag.Diagnostics) {
	if remoteExecution == nil {
		return types.ObjectNull(schemas.RemoteExecutionAttributeTypes()), nil
	}
	allowedIpAddressRanges, diags := utils.StringSet(&remoteExecution.AllowedIpAddressRanges)
	if diags.HasError() {
		return types.ObjectNull(schemas.RemoteExecutionAttributeTypes()), diags
	}
	obj := RemoteExecution{
		Enabled:                types.BoolValue(remoteExecution.Enabled),
		AllowedIpAddressRanges: allowedIpAddressRanges,
		RemoteApiUrl:           types.StringValue(remoteExecution.RemoteApiUrl),
		TaskLogBucket:          types.StringPointerValue(remoteExecution.TaskLogBucket),
		TaskLogUrlPattern:      types.StringPointerValue(remoteExecution.TaskLogUrlPattern),
	}
	return types.ObjectValueFrom(ctx, schemas.RemoteExecutionAttributeTypes(), obj)
}
