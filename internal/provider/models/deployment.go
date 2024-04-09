package models

import (
	"context"
	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/schemas"
	"github.com/astronomer/astronomer-terraform-provider/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DeploymentDataSource describes the data source data model.
type DeploymentDataSource struct {
	Id                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	CreatedAt                types.String `tfsdk:"created_at"`
	UpdatedAt                types.String `tfsdk:"updated_at"`
	CreatedBy                types.Object `tfsdk:"created_by"`
	UpdatedBy                types.Object `tfsdk:"updated_by"`
	WorkspaceId              types.String `tfsdk:"workspace_id"`
	WorkspaceName            types.String `tfsdk:"workspace_name"`
	ClusterId                types.String `tfsdk:"cluster_id"`
	ClusterName              types.String `tfsdk:"cluster_name"`
	Region                   types.String `tfsdk:"region"`
	CloudProvider            types.String `tfsdk:"cloud_provider"`
	AstroRuntimeVersion      types.String `tfsdk:"astro_runtime_version"`
	AirflowVersion           types.String `tfsdk:"airflow_version"`
	Namespace                types.String `tfsdk:"namespace"`
	ContactEmails            types.List   `tfsdk:"contact_emails"`
	Executor                 types.String `tfsdk:"executor"`
	SchedulerAu              types.Int64  `tfsdk:"scheduler_au"`
	SchedulerCpu             types.String `tfsdk:"scheduler_cpu"`
	SchedulerMemory          types.String `tfsdk:"scheduler_memory"`
	SchedulerReplicas        types.Int64  `tfsdk:"scheduler_replicas"`
	ImageTag                 types.String `tfsdk:"image_tag"`
	ImageRepository          types.String `tfsdk:"image_repository"`
	ImageVersion             types.String `tfsdk:"image_version"`
	EnvironmentVariables     types.List   `tfsdk:"environment_variables"`
	WebserverIngressHostname types.String `tfsdk:"webserver_ingress_hostname"`
	WebserverUrl             types.String `tfsdk:"webserver_url"`
	WebserverAirflowApiUrl   types.String `tfsdk:"webserver_airflow_api_url"`
	WebserverCpu             types.String `tfsdk:"webserver_cpu"`
	WebserverMemory          types.String `tfsdk:"webserver_memory"`
	WebserverReplicas        types.Int64  `tfsdk:"webserver_replicas"`
	Status                   types.String `tfsdk:"status"`
	StatusReason             types.String `tfsdk:"status_reason"`
	DagTarballVersion        types.String `tfsdk:"dag_tarball_version"`
	DesiredDagTarballVersion types.String `tfsdk:"desired_dag_tarball_version"`
	WorkerQueues             types.List   `tfsdk:"worker_queues"`
	TaskPodNodePoolId        types.String `tfsdk:"task_pod_node_pool_id"`
	IsCicdEnforced           types.Bool   `tfsdk:"is_cicd_enforced"`
	Type                     types.String `tfsdk:"type"`
	IsDagDeployEnabled       types.Bool   `tfsdk:"is_dag_deploy_enabled"`
	SchedulerSize            types.String `tfsdk:"scheduler_size"`
	IsHighAvailability       types.Bool   `tfsdk:"is_high_availability"`
	IsDevelopmentMode        types.Bool   `tfsdk:"is_development_mode"`
	WorkloadIdentity         types.String `tfsdk:"workload_identity"`
	ExternalIps              types.List   `tfsdk:"external_ips"`
	OidcIssuerUrl            types.String `tfsdk:"oidc_issuer_url"`
	ResourceQuotaCpu         types.String `tfsdk:"resource_quota_cpu"`
	ResourceQuotaMemory      types.String `tfsdk:"resource_quota_memory"`
	DefaultTaskPodCpu        types.String `tfsdk:"default_task_pod_cpu"`
	DefaultTaskPodMemory     types.String `tfsdk:"default_task_pod_memory"`
	//ScalingStatus            types.Object `tfsdk:"scaling_status"`
	//ScalingSpec              types.Object `tfsdk:"scaling_spec"`
}

type DeploymentEnvironmentVariable struct {
	Key       types.String `tfsdk:"key"`
	Value     types.String `tfsdk:"value"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	IsSecret  types.Bool   `tfsdk:"is_secret"`
}

type WorkerQueue struct {
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

func (data *DeploymentDataSource) ReadFromResponse(
	ctx context.Context,
	deployment *platform.Deployment,
) diag.Diagnostics {
	data.Id = types.StringValue(deployment.Id)
	data.Name = types.StringValue(deployment.Name)
	data.Description = types.StringPointerValue(deployment.Description)
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
	data.WorkspaceName = types.StringPointerValue(deployment.WorkspaceName)
	data.ClusterId = types.StringPointerValue(deployment.ClusterId)
	data.ClusterName = types.StringPointerValue(deployment.ClusterName)
	data.Region = types.StringPointerValue(deployment.Region)
	data.CloudProvider = types.StringPointerValue((*string)(deployment.CloudProvider))
	data.AstroRuntimeVersion = types.StringValue(deployment.AstroRuntimeVersion)
	data.AirflowVersion = types.StringValue(deployment.AirflowVersion)
	data.Namespace = types.StringValue(deployment.Namespace)
	if deployment.ContactEmails != nil {
		data.ContactEmails, diags = utils.StringList(*deployment.ContactEmails)
		if diags.HasError() {
			return diags
		}
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
	if deployment.EnvironmentVariables != nil {
		data.EnvironmentVariables, diags = utils.ObjectList(ctx, *deployment.EnvironmentVariables, schemas.DeploymentEnvironmentVariableAttributeTypes(), DeploymentEnvironmentVariableTypesObject)
		if diags.HasError() {
			return diags
		}
	}
	data.WebserverIngressHostname = types.StringValue(deployment.WebServerIngressHostname)
	data.WebserverUrl = types.StringValue(deployment.WebServerUrl)
	data.WebserverAirflowApiUrl = types.StringValue(deployment.WebServerAirflowApiUrl)
	data.WebserverCpu = types.StringValue(deployment.WebServerCpu)
	data.WebserverMemory = types.StringValue(deployment.WebServerMemory)
	if deployment.WebServerReplicas != nil {
		data.WebserverReplicas = types.Int64Value(int64(*deployment.WebServerReplicas))
	}
	data.Status = types.StringValue(string(deployment.Status))
	data.StatusReason = types.StringPointerValue(deployment.StatusReason)
	data.DagTarballVersion = types.StringPointerValue(deployment.DagTarballVersion)
	data.DesiredDagTarballVersion = types.StringPointerValue(deployment.DesiredDagTarballVersion)
	if deployment.WorkerQueues != nil {
		data.WorkerQueues, diags = utils.ObjectList(ctx, *deployment.WorkerQueues, schemas.WorkerQueueAttributeTypes(), WorkerQueueTypesObject)
		if diags.HasError() {
			return diags
		}
	}
	data.TaskPodNodePoolId = types.StringPointerValue(deployment.TaskPodNodePoolId)
	data.IsCicdEnforced = types.BoolValue(deployment.IsCicdEnforced)
	data.Type = types.StringPointerValue((*string)(deployment.Type))
	data.IsDagDeployEnabled = types.BoolValue(deployment.IsDagDeployEnabled)
	data.SchedulerSize = types.StringPointerValue((*string)(deployment.SchedulerSize))
	data.IsHighAvailability = types.BoolPointerValue(deployment.IsHighAvailability)
	data.IsDevelopmentMode = types.BoolPointerValue(deployment.IsDevelopmentMode)
	data.WorkloadIdentity = types.StringPointerValue(deployment.WorkloadIdentity)
	if deployment.ExternalIPs != nil {
		data.ExternalIps, diags = utils.StringList(*deployment.ExternalIPs)
		if diags.HasError() {
			return diags
		}
	}
	data.OidcIssuerUrl = types.StringPointerValue(deployment.OidcIssuerUrl)
	data.ResourceQuotaCpu = types.StringPointerValue(deployment.ResourceQuotaCpu)
	data.ResourceQuotaMemory = types.StringPointerValue(deployment.ResourceQuotaMemory)
	data.DefaultTaskPodCpu = types.StringPointerValue(deployment.DefaultTaskPodCpu)
	data.DefaultTaskPodMemory = types.StringPointerValue(deployment.DefaultTaskPodMemory)
	//data.ScalingStatus, diags = ScalingStatusTypesObject(ctx, deployment.ScalingStatus)
	//if diags.HasError() {
	//	return diags
	//}
	//data.ScalingSpec, diags = ScalingSpecTypesObject(ctx, deployment.ScalingSpec)
	//if diags.HasError() {
	//	return diags
	//}

	return nil
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

func WorkerQueueTypesObject(
	ctx context.Context,
	workerQueue platform.WorkerQueue,
) (types.Object, diag.Diagnostics) {
	obj := WorkerQueue{
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

	return types.ObjectValueFrom(ctx, schemas.WorkerQueueAttributeTypes(), obj)
}
