package models

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"

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
	data.EnvironmentVariables, diags = utils.ObjectSet(ctx, deployment.EnvironmentVariables, schemas.DeploymentEnvironmentVariableAttributeTypes(), DeploymentEnvironmentVariableTypesObject)
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
	data.ScalingStatus, diags = ScalingStatusTypesObject(ctx, deployment.ScalingStatus)
	if diags.HasError() {
		return diags
	}
	data.ScalingSpec, diags = ScalingSpecTypesObject(ctx, deployment.ScalingSpec)
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
	envVar platform.DeploymentEnvironmentVariable,
) (types.Object, diag.Diagnostics) {
	obj := DeploymentEnvironmentVariable{
		Key:       types.StringValue(envVar.Key),
		Value:     types.StringPointerValue(envVar.Value),
		UpdatedAt: types.StringValue(envVar.UpdatedAt),
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
	hibernationStatus *platform.DeploymentHibernationStatus,
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
	hibernationOverride *platform.DeploymentHibernationOverride,
) (types.Object, diag.Diagnostics) {
	if hibernationOverride == nil {
		return types.ObjectNull(schemas.HibernationOverrideAttributeTypes()), nil
	}
	obj := HibernationSpecOverride{
		IsHibernating: types.BoolPointerValue(hibernationOverride.IsHibernating),
		OverrideUntil: types.StringValue(hibernationOverride.OverrideUntil.Format(time.RFC3339)),
		IsActive:      types.BoolPointerValue(hibernationOverride.IsActive),
	}
	return types.ObjectValueFrom(ctx, schemas.HibernationOverrideAttributeTypes(), obj)
}

func HibernationScheduleTypesObject(
	ctx context.Context,
	schedule platform.DeploymentHibernationSchedule,
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
	hibernationSpec *platform.DeploymentHibernationSpec,
) (types.Object, diag.Diagnostics) {
	if hibernationSpec == nil {
		return types.ObjectNull(schemas.HibernationSpecAttributeTypes()), nil
	}
	override, diags := HibernationOverrideTypesObject(ctx, hibernationSpec.Override)
	if diags.HasError() {
		tflog.Error(ctx, "Failed to create HibernationOverride object", map[string]interface{}{"error": diags})
		return types.ObjectNull(schemas.HibernationSpecAttributeTypes()), diags
	}
	schedules, diags := utils.ObjectSet(ctx, hibernationSpec.Schedules, schemas.HibernationScheduleAttributeTypes(), HibernationScheduleTypesObject)
	if diags.HasError() {
		tflog.Error(ctx, "Failed to create HibernationSchedule object", map[string]interface{}{"error": diags})
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
	scalingStatus *platform.DeploymentScalingStatus,
) (types.Object, diag.Diagnostics) {
	if scalingStatus == nil {
		return types.ObjectNull(schemas.ScalingStatusAttributeTypes()), nil
	}

	hibernationStatus, diags := HibernationStatusTypesObject(ctx, scalingStatus.HibernationStatus)
	if diags.HasError() {
		tflog.Error(ctx, "Failed to create HibernationStatus object", map[string]interface{}{"error": diags})
		return types.ObjectNull(schemas.ScalingStatusAttributeTypes()), diags
	}
	obj := DeploymentStatus{
		HibernationStatus: hibernationStatus,
	}
	return types.ObjectValueFrom(ctx, schemas.ScalingStatusAttributeTypes(), obj)
}

func ScalingSpecTypesObject(
	ctx context.Context,
	scalingSpec *platform.DeploymentScalingSpec,
) (types.Object, diag.Diagnostics) {
	if scalingSpec == nil {
		return types.ObjectNull(schemas.ScalingSpecAttributeTypes()), nil
	}

	hibernationSpec, diags := HibernationSpecTypesObject(ctx, scalingSpec.HibernationSpec)
	if diags.HasError() {
		tflog.Error(ctx, "Failed to create HibernationSpec object", map[string]interface{}{"error": diags})
		return types.ObjectNull(schemas.ScalingSpecAttributeTypes()), diags
	}
	obj := DeploymentScalingSpec{
		HibernationSpec: hibernationSpec,
	}
	return types.ObjectValueFrom(ctx, schemas.ScalingSpecAttributeTypes(), obj)
}
