package models

import (
	"context"
	"time"

	"github.com/samber/lo"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Deployment struct {
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
	SchedulerSize        types.String `tfsdk:"scheduler_size"`
	IsDevelopmentMode    types.Bool   `tfsdk:"is_development_mode"`
	IsHighAvailability   types.Bool   `tfsdk:"is_high_availability"`
}

func (data *Deployment) ReadFromResponse(
	ctx context.Context,
	deployment *platform.Deployment,
	isResource bool,
	requestEnvVars *[]platform.DeploymentEnvironmentVariableRequest,
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
	if deployment.SchedulerAu != nil {
		deploymentSchedulerAu := int64(*deployment.SchedulerAu)
		data.SchedulerAu = types.Int64Value(deploymentSchedulerAu)
	}
	data.SchedulerCpu = types.StringValue(deployment.SchedulerCpu)
	data.SchedulerMemory = types.StringValue(deployment.SchedulerMemory)
	data.SchedulerReplicas = types.Int64Value(int64(deployment.SchedulerReplicas))
	data.ImageTag = types.StringValue(deployment.ImageTag)
	data.ImageRepository = types.StringValue(deployment.ImageRepository)
	data.ImageVersion = types.StringPointerValue(deployment.ImageVersion)

	// Environment variables are a special case
	// Since terraform wants to know the values of the secret values in the request at all times, and our API does not send back the secret values in the response
	// We must use the request value and set it in the Terraform response to keep Terraform from emitting errors
	// Since the value is marked as sensitive, Terraform will not output the actual value in the plan/apply output
	envVars := *deployment.EnvironmentVariables
	if requestEnvVars != nil && deployment.EnvironmentVariables != nil {
		requestEnvVarsMap := lo.SliceToMap(*requestEnvVars, func(envVar platform.DeploymentEnvironmentVariableRequest) (string, platform.DeploymentEnvironmentVariable) {
			return envVar.Key, platform.DeploymentEnvironmentVariable{
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
	data.EnvironmentVariables, diags = utils.ObjectSet(ctx, &envVars, schemas.DeploymentEnvironmentVariableAttributeTypes(), DeploymentEnvironmentVariableTypesObject)
	if diags.HasError() {
		return diags
	}
	data.WebserverIngressHostname = types.StringValue(deployment.WebServerIngressHostname)
	data.WebserverUrl = types.StringValue(deployment.WebServerUrl)
	data.WebserverAirflowApiUrl = types.StringValue(deployment.WebServerAirflowApiUrl)
	data.Status = types.StringValue(string(deployment.Status))
	data.StatusReason = types.StringPointerValue(deployment.StatusReason)
	data.Type = types.StringPointerValue((*string)(deployment.Type))
	data.DagTarballVersion = types.StringPointerValue(deployment.DagTarballVersion)
	data.DesiredDagTarballVersion = types.StringPointerValue(deployment.DesiredDagTarballVersion)
	if isResource {
		data.WorkerQueues, diags = utils.ObjectSet(ctx, deployment.WorkerQueues, schemas.WorkerQueueResourceAttributeTypes(), WorkerQueueResourceTypesObject)
		if diags.HasError() {
			return diags
		}
	} else {
		data.WorkerQueues, diags = utils.ObjectSet(ctx, deployment.WorkerQueues, schemas.WorkerQueueDataSourceAttributeTypes(), WorkerQueueDataSourceTypesObject)
		if diags.HasError() {
			return diags
		}
	}
	data.IsCicdEnforced = types.BoolValue(deployment.IsCicdEnforced)
	data.IsDagDeployEnabled = types.BoolValue(deployment.IsDagDeployEnabled)

	data.WorkloadIdentity = types.StringPointerValue(deployment.WorkloadIdentity)
	data.ExternalIps, diags = utils.StringSet(deployment.ExternalIPs)
	if diags.HasError() {
		return diags
	}
	data.OidcIssuerUrl = types.StringPointerValue(deployment.OidcIssuerUrl)

	// Read hybrid and dedicated specific fields
	data.ClusterId = types.StringPointerValue(deployment.ClusterId)

	// Read hybrid deployment specific fields
	data.TaskPodNodePoolId = types.StringPointerValue(deployment.TaskPodNodePoolId)

	// Read hosted deployment specific fields
	data.ResourceQuotaCpu = types.StringPointerValue(deployment.ResourceQuotaCpu)
	data.ResourceQuotaMemory = types.StringPointerValue(deployment.ResourceQuotaMemory)
	data.DefaultTaskPodCpu = types.StringPointerValue(deployment.DefaultTaskPodCpu)
	data.DefaultTaskPodMemory = types.StringPointerValue(deployment.DefaultTaskPodMemory)
	data.SchedulerSize = types.StringPointerValue((*string)(deployment.SchedulerSize))
	data.IsHighAvailability = types.BoolPointerValue(deployment.IsHighAvailability)
	data.IsDevelopmentMode = types.BoolPointerValue(deployment.IsDevelopmentMode)

	// Currently, the scaling status and spec are only available in development mode
	// However, there is a bug in the API where the scaling status and spec are returned even if the deployment is not in development mode for updated deployments
	// This is a workaround to handle the bug until the API is fixed
	// Issue here: https://github.com/astronomer/astro/issues/21073
	if deployment.IsDevelopmentMode != nil && *deployment.IsDevelopmentMode {
		data.ScalingStatus, diags = ScalingStatusTypesObject(ctx, deployment.ScalingStatus)
		if diags.HasError() {
			return diags
		}
		data.ScalingSpec, diags = ScalingSpecTypesObject(ctx, deployment.ScalingSpec)
		if diags.HasError() {
			return diags
		}
	} else {
		data.ScalingStatus = types.ObjectNull(schemas.ScalingStatusAttributeTypes())
		data.ScalingSpec = types.ObjectNull(schemas.ScalingSpecAttributeTypes())
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
	envVar platform.DeploymentEnvironmentVariable,
) (types.Object, diag.Diagnostics) {
	obj := DeploymentEnvironmentVariable{
		Key:       types.StringValue(envVar.Key),
		UpdatedAt: types.StringValue(envVar.UpdatedAt),
		Value:     types.StringPointerValue(envVar.Value),
		IsSecret:  types.BoolValue(envVar.IsSecret),
	}

	return types.ObjectValueFrom(ctx, schemas.DeploymentEnvironmentVariableAttributeTypes(), obj)
}

func WorkerQueueResourceTypesObject(
	ctx context.Context,
	workerQueue platform.WorkerQueue,
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
	workerQueue platform.WorkerQueue,
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
	HibernationSpec HibernationSpec `tfsdk:"hibernation_spec"`
}

type DeploymentStatus struct {
	HibernationStatus HibernationStatus `tfsdk:"hibernation_status"`
}

type HibernationStatus struct {
	IsHibernating types.Bool   `tfsdk:"is_hibernating"`
	NextEventType types.String `tfsdk:"next_event_type"`
	NextEventAt   types.String `tfsdk:"next_event_at"`
	Reason        types.String `tfsdk:"reason"`
}

type HibernationSpec struct {
	Override  HibernationSpecOverride `tfsdk:"override"`
	Schedules []HibernationSchedule   `tfsdk:"schedules"`
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

func ScalingStatusTypesObject(
	ctx context.Context,
	scalingStatus *platform.DeploymentScalingStatus,
) (types.Object, diag.Diagnostics) {
	if scalingStatus != nil && scalingStatus.HibernationStatus != nil {
		obj := DeploymentStatus{
			HibernationStatus: HibernationStatus{
				IsHibernating: types.BoolValue(scalingStatus.HibernationStatus.IsHibernating),
				NextEventType: types.StringPointerValue((*string)(scalingStatus.HibernationStatus.NextEventType)),
				NextEventAt:   types.StringPointerValue(scalingStatus.HibernationStatus.NextEventAt),
				Reason:        types.StringPointerValue(scalingStatus.HibernationStatus.Reason),
			},
		}
		return types.ObjectValueFrom(ctx, schemas.ScalingStatusAttributeTypes(), obj)
	}
	return types.ObjectNull(schemas.ScalingStatusAttributeTypes()), nil
}

func ScalingSpecTypesObject(
	ctx context.Context,
	scalingSpec *platform.DeploymentScalingSpec,
) (types.Object, diag.Diagnostics) {
	if scalingSpec != nil && scalingSpec.HibernationSpec != nil && (scalingSpec.HibernationSpec.Override != nil || scalingSpec.HibernationSpec.Schedules != nil) {
		obj := DeploymentScalingSpec{
			HibernationSpec: HibernationSpec{},
		}
		if scalingSpec.HibernationSpec.Override != nil {
			obj.HibernationSpec.Override = HibernationSpecOverride{
				IsHibernating: types.BoolPointerValue(scalingSpec.HibernationSpec.Override.IsHibernating),
				IsActive:      types.BoolPointerValue(scalingSpec.HibernationSpec.Override.IsActive),
			}
			if scalingSpec.HibernationSpec.Override.OverrideUntil != nil {
				obj.HibernationSpec.Override.OverrideUntil = types.StringValue(scalingSpec.HibernationSpec.Override.OverrideUntil.Format(time.RFC3339))
			}
		}
		if scalingSpec.HibernationSpec.Schedules != nil {
			schedules := make([]HibernationSchedule, 0, len(*scalingSpec.HibernationSpec.Schedules))
			for _, schedule := range *scalingSpec.HibernationSpec.Schedules {
				schedules = append(schedules, HibernationSchedule{
					Description:     types.StringPointerValue(schedule.Description),
					HibernateAtCron: types.StringValue(schedule.HibernateAtCron),
					IsEnabled:       types.BoolValue(schedule.IsEnabled),
					WakeAtCron:      types.StringValue(schedule.WakeAtCron),
				})
			}
			obj.HibernationSpec.Schedules = schedules
		}
		return types.ObjectValueFrom(ctx, schemas.ScalingSpecAttributeTypes(), obj)
	}
	return types.ObjectNull(schemas.ScalingSpecAttributeTypes()), nil
}
