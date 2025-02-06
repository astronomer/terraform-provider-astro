package schemas

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func DeploymentResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment identifier",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment name",
			Required:            true,
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment description",
			Required:            true,
		},
		"created_at": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment creation timestamp",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_at": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment last updated timestamp",
			Computed:            true,
		},
		"created_by": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Deployment creator",
			Computed:            true,
			Attributes:          ResourceSubjectProfileSchemaAttributes(),
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_by": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Deployment updater",
			Computed:            true,
			Attributes:          ResourceSubjectProfileSchemaAttributes(),
		},
		"workspace_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment workspace identifier - if changing this value, the deployment will be recreated in the new workspace",
			Required:            true,
			Validators:          []validator.String{validators.IsCuid()},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
		"original_astro_runtime_version": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment's original Astro Runtime version. The Terraform provider will use this provided Astro runtime version to create the Deployment. The Astro runtime version can be updated with your Astro project Dockerfile, but if this value is changed, the Deployment will be recreated with this new Astro runtime version.",
			Optional:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
		"astro_runtime_version": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment's current Astro Runtime version",
			Computed:            true,
		},
		"airflow_version": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment Airflow version",
			Computed:            true,
		},
		"namespace": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment namespace",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"contact_emails": resourceSchema.SetAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Deployment contact emails",
			Required:            true,
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(stringvalidator.RegexMatches(regexp.MustCompile(validators.EmailString), "must be a valid email address")),
			},
		},
		"executor": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment executor",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(string(platform.DeploymentExecutorCELERY), string(platform.DeploymentExecutorKUBERNETES)),
			},
		},
		"scheduler_cpu": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment scheduler CPU",
			Computed:            true,
		},
		"scheduler_memory": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment scheduler memory",
			Computed:            true,
		},
		"image_tag": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment image tag",
			Computed:            true,
		},
		"image_repository": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment image repository",
			Computed:            true,
		},
		"image_version": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment image version",
			Computed:            true,
		},
		"environment_variables": resourceSchema.SetNestedAttribute{
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: DeploymentEnvironmentVariableResourceAttributes(),
			},
			MarkdownDescription: "Deployment environment variables",
			Required:            true,
		},
		"webserver_ingress_hostname": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment webserver ingress hostname",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"webserver_url": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment webserver URL",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"webserver_airflow_api_url": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment webserver Airflow API URL",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"status": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment status",
			Computed:            true,
		},
		"status_reason": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment status reason",
			Computed:            true,
		},
		"dag_tarball_version": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment DAG tarball version",
			Computed:            true,
		},
		"desired_dag_tarball_version": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment desired DAG tarball version",
			Computed:            true,
		},
		"is_cicd_enforced": resourceSchema.BoolAttribute{
			MarkdownDescription: "Deployment CI/CD enforced",
			Required:            true,
		},
		"is_dag_deploy_enabled": resourceSchema.BoolAttribute{
			MarkdownDescription: "Whether DAG deploy is enabled - Changing this value may disrupt your deployment. Read more at https://docs.astronomer.io/astro/deploy-dags#enable-or-disable-dag-only-deploys-on-a-deployment",
			Required:            true,
		},
		"external_ips": resourceSchema.SetAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Deployment external IPs",
			Computed:            true,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
		},
		"oidc_issuer_url": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment OIDC issuer URL",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"desired_workload_identity": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment's desired workload identity. The Terraform provider will use this provided workload identity to create the Deployment. If it is not provided the worload identity will be assigned automatically.",
			Optional:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
		"workload_identity": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment workload identity. This value can be changed via the Astro API if applicable.",
			Computed:            true,
		},
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment type - if changing this value, the deployment will be recreated with the new type",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(string(platform.DeploymentTypeSTANDARD), string(platform.DeploymentTypeDEDICATED), string(platform.DeploymentTypeHYBRID)),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
		"worker_queues": resourceSchema.SetNestedAttribute{
			Optional: true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: WorkerQueueResourceSchemaAttributes(),
			},
			MarkdownDescription: "Deployment worker queues - required for deployments with 'CELERY' executor. For 'STANDARD' and 'DEDICATED' deployments, use astro_machine. For 'HYBRID' deployments, use node_pool_id.",
			Validators: []validator.Set{
				// Dynamic validation with 'executor' done in the resource.ValidateConfig function
				setvalidator.SizeAtLeast(1),
			},
		},
		"scheduler_au": resourceSchema.Int64Attribute{
			MarkdownDescription: "Deployment scheduler AU - required for 'HYBRID' deployments",
			Optional:            true,
			Validators: []validator.Int64{
				int64validator.Between(5, 24),
			},
		},
		"scheduler_replicas": resourceSchema.Int64Attribute{
			MarkdownDescription: "Deployment scheduler replicas - required for 'HYBRID' deployments",
			Optional:            true,
			Computed:            true,
			Validators: []validator.Int64{
				int64validator.Between(1, 4),
			},
		},
		"task_pod_node_pool_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment task pod node pool identifier - required if executor is 'KUBERNETES' and type is 'HYBRID'",
			Optional:            true,
		},
		"scheduler_size": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment scheduler size - required for 'STANDARD' and 'DEDICATED' deployments",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.SchedulerMachineNameSMALL),
					string(platform.SchedulerMachineNameMEDIUM),
					string(platform.SchedulerMachineNameLARGE),
					string(platform.SchedulerMachineNameEXTRALARGE),
				),
			},
		},
		"is_high_availability": resourceSchema.BoolAttribute{
			MarkdownDescription: "Deployment high availability - required for 'STANDARD' and 'DEDICATED' deployments",
			Optional:            true,
		},
		"is_development_mode": resourceSchema.BoolAttribute{
			MarkdownDescription: "Deployment development mode - required for 'STANDARD' and 'DEDICATED' deployments. If changing from 'False' to 'True', the deployment will be recreated",
			Optional:            true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplaceIf(
					func(_ context.Context, req planmodifier.BoolRequest, resp *boolplanmodifier.RequiresReplaceIfFuncResponse) {
						prevIsDevelopmentMode := req.StateValue.ValueBool()
						newIsDevelopmentMode := req.ConfigValue.ValueBool()
						if prevIsDevelopmentMode == false && newIsDevelopmentMode == true {
							resp.RequiresReplace = true
						}
					},
					"Converting a non-development mode deployment to a development mode deployment is not allowed. Therefore, the deployment will be recreated as a development-mode deployment",
					"Converting a non-development mode deployment to a development mode deployment is not allowed. Therefore, the deployment will be recreated as a development-mode deployment",
				),
			},
		},
		"resource_quota_cpu": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment resource quota CPU - required for 'STANDARD' and 'DEDICATED' deployments",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile(validators.KubernetesResourceString), "must be a valid kubernetes resource string"),
			},
		},
		"resource_quota_memory": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment resource quota memory - required for 'STANDARD' and 'DEDICATED' deployments",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile(validators.KubernetesResourceString), "must be a valid kubernetes resource string"),
			},
		},
		"default_task_pod_cpu": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment default task pod CPU - required for 'STANDARD' and 'DEDICATED' deployments",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile(validators.KubernetesResourceString), "must be a valid kubernetes resource string"),
			},
		},
		"default_task_pod_memory": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment default task pod memory - required for 'STANDARD' and 'DEDICATED' deployments",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile(validators.KubernetesResourceString), "must be a valid kubernetes resource string"),
			},
		},
		"scaling_status": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Deployment scaling status",
			Computed:            true,
			Attributes:          ScalingStatusResourceAttributes(),
		},
		"scaling_spec": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Deployment scaling spec - only for 'STANDARD' and 'DEDICATED' deployments",
			Optional:            true,
			Attributes:          ScalingSpecResourceSchemaAttributes(),
		},
		"region": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment region - required for 'STANDARD' deployments. If changing this value, the deployment will be recreated in the new region",
			Computed:            true,
			Optional:            true,
			PlanModifiers: []planmodifier.String{
				// Would recreate the deployment if this attribute changes
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
		"cloud_provider": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment cloud provider - required for 'STANDARD' deployments. If changing this value, the deployment will be recreated in the new cloud provider",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				// Would recreate the deployment if this attribute changes
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
			Validators: []validator.String{
				stringvalidator.OneOf(string(platform.ClusterCloudProviderAWS), string(platform.ClusterCloudProviderAZURE), string(platform.ClusterCloudProviderGCP)),
			},
		},
		"cluster_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Deployment cluster identifier - required for 'HYBRID' and 'DEDICATED' deployments. If changing this value, the deployment will be recreated in the new cluster",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				validators.IsCuid(),
			},
			PlanModifiers: []planmodifier.String{
				// Would recreate the deployment if this attribute changes
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
	}
}

func DeploymentDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment identifier",
			Required:            true,
			Validators:          []validator.String{validators.IsCuid()},
		},
		"name": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment name",
			Computed:            true,
		},
		"description": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment description",
			Computed:            true,
		},
		"created_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment creation timestamp",
			Computed:            true,
		},
		"updated_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment last updated timestamp",
			Computed:            true,
		},
		"created_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Deployment creator",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
		"updated_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Deployment updater",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
		"workspace_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment workspace identifier",
			Computed:            true,
		},
		"cluster_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment cluster identifier",
			Computed:            true,
		},
		"region": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment region",
			Computed:            true,
		},
		"cloud_provider": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment cloud provider",
			Computed:            true,
		},
		"astro_runtime_version": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment Astro Runtime version",
			Computed:            true,
		},
		"airflow_version": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment Airflow version",
			Computed:            true,
		},
		"namespace": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment namespace",
			Computed:            true,
		},
		"contact_emails": datasourceSchema.SetAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Deployment contact emails",
			Computed:            true,
		},
		"executor": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment executor",
			Computed:            true,
		},
		"scheduler_au": datasourceSchema.Int64Attribute{
			MarkdownDescription: "Deployment scheduler AU",
			Computed:            true,
		},
		"scheduler_cpu": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment scheduler CPU",
			Computed:            true,
		},
		"scheduler_memory": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment scheduler memory",
			Computed:            true,
		},
		"scheduler_replicas": datasourceSchema.Int64Attribute{
			MarkdownDescription: "Deployment scheduler replicas",
			Computed:            true,
		},
		"image_tag": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment image tag",
			Computed:            true,
		},
		"image_repository": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment image repository",
			Computed:            true,
		},
		"image_version": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment image version",
			Computed:            true,
		},
		"environment_variables": datasourceSchema.SetNestedAttribute{
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: DeploymentEnvironmentVariableDataSourceAttributes(),
			},
			MarkdownDescription: "Deployment environment variables",
			Computed:            true,
		},
		"webserver_ingress_hostname": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment webserver ingress hostname",
			Computed:            true,
		},
		"webserver_url": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment webserver URL",
			Computed:            true,
		},
		"webserver_airflow_api_url": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment webserver Airflow API URL",
			Computed:            true,
		},
		"status": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment status",
			Computed:            true,
		},
		"status_reason": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment status reason",
			Computed:            true,
		},
		"dag_tarball_version": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment DAG tarball version",
			Computed:            true,
		},
		"desired_dag_tarball_version": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment desired DAG tarball version",
			Computed:            true,
		},
		"worker_queues": datasourceSchema.SetNestedAttribute{
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: WorkerQueueDataSourceSchemaAttributes(),
			},
			MarkdownDescription: "Deployment worker queues",
			Computed:            true,
		},
		"task_pod_node_pool_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment task pod node pool identifier",
			Computed:            true,
		},
		"is_cicd_enforced": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether the Deployment enforces CI/CD deploys",
			Computed:            true,
		},
		"type": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment type",
			Computed:            true,
		},
		"is_dag_deploy_enabled": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether DAG deploy is enabled",
			Computed:            true,
		},
		"scheduler_size": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment scheduler size",
			Computed:            true,
		},
		"is_high_availability": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether Deployment has high availability",
			Computed:            true,
		},
		"is_development_mode": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether Deployment is in development mode",
			Computed:            true,
		},
		"workload_identity": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment workload identity",
			Computed:            true,
		},
		"external_ips": datasourceSchema.SetAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Deployment external IPs",
			Computed:            true,
		},
		"oidc_issuer_url": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment OIDC issuer URL",
			Computed:            true,
		},
		"resource_quota_cpu": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment resource quota CPU",
			Computed:            true,
		},
		"resource_quota_memory": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment resource quota memory",
			Computed:            true,
		},
		"default_task_pod_cpu": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment default task pod CPU",
			Computed:            true,
		},
		"default_task_pod_memory": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment default task pod memory",
			Computed:            true,
		},
		"scaling_status": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Deployment scaling status",
			Computed:            true,
			Attributes:          ScalingStatusDataSourceAttributes(),
		},
		"scaling_spec": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Deployment scaling spec",
			Computed:            true,
			Attributes:          ScalingSpecDataSourceSchemaAttributes(),
		},
	}
}

func DeploymentEnvironmentVariableAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"key":        types.StringType,
		"value":      types.StringType,
		"updated_at": types.StringType,
		"is_secret":  types.BoolType,
	}
}

func DeploymentEnvironmentVariableDataSourceAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"key": datasourceSchema.StringAttribute{
			MarkdownDescription: "Environment variable key",
			Computed:            true,
		},
		"value": datasourceSchema.StringAttribute{
			MarkdownDescription: "Environment variable value",
			Computed:            true,
		},
		"updated_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Environment variable last updated timestamp",
			Computed:            true,
		},
		"is_secret": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether Environment variable is a secret",
			Computed:            true,
		},
	}
}

func DeploymentEnvironmentVariableResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"key": resourceSchema.StringAttribute{
			MarkdownDescription: "Environment variable key",
			Required:            true,
		},
		"value": resourceSchema.StringAttribute{
			MarkdownDescription: "Environment variable value",
			Optional:            true,
			Sensitive:           true,
		},
		"updated_at": resourceSchema.StringAttribute{
			MarkdownDescription: "Environment variable last updated timestamp",
			Computed:            true,
		},
		"is_secret": resourceSchema.BoolAttribute{
			MarkdownDescription: "Whether Environment variable is a secret",
			Required:            true,
		},
	}
}

func WorkerQueueResourceAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":               types.StringType,
		"is_default":         types.BoolType,
		"max_worker_count":   types.Int64Type,
		"min_worker_count":   types.Int64Type,
		"pod_cpu":            types.StringType,
		"pod_memory":         types.StringType,
		"worker_concurrency": types.Int64Type,
		"node_pool_id":       types.StringType,
		"astro_machine":      types.StringType,
	}
}

func WorkerQueueDataSourceAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                 types.StringType,
		"name":               types.StringType,
		"is_default":         types.BoolType,
		"max_worker_count":   types.Int64Type,
		"min_worker_count":   types.Int64Type,
		"pod_cpu":            types.StringType,
		"pod_memory":         types.StringType,
		"worker_concurrency": types.Int64Type,
		"node_pool_id":       types.StringType,
		"astro_machine":      types.StringType,
	}
}

func WorkerQueueDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Worker queue identifier",
			Computed:            true,
		},
		"name": datasourceSchema.StringAttribute{
			MarkdownDescription: "Worker queue name",
			Computed:            true,
		},
		"astro_machine": datasourceSchema.StringAttribute{
			MarkdownDescription: "Worker queue Astro machine value",
			Computed:            true,
		},
		"is_default": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether Worker queue is default",
			Computed:            true,
		},
		"max_worker_count": datasourceSchema.Int64Attribute{
			MarkdownDescription: "Worker queue max worker count",
			Computed:            true,
		},
		"min_worker_count": datasourceSchema.Int64Attribute{
			MarkdownDescription: "Worker queue min worker count",
			Computed:            true,
		},
		"node_pool_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Worker queue node pool identifier",
			Computed:            true,
		},
		"pod_cpu": datasourceSchema.StringAttribute{
			MarkdownDescription: "Worker queue pod CPU",
			Computed:            true,
		},
		"pod_memory": datasourceSchema.StringAttribute{
			MarkdownDescription: "Worker queue pod memory",
			Computed:            true,
		},
		"worker_concurrency": datasourceSchema.Int64Attribute{
			MarkdownDescription: "Worker queue worker concurrency",
			Computed:            true,
		},
	}
}

func WorkerQueueResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Worker queue name",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 63),
			},
		},
		"is_default": resourceSchema.BoolAttribute{
			MarkdownDescription: "Worker queue default",
			Required:            true,
		},
		"max_worker_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Worker queue max worker count",
			Required:            true,
			Validators: []validator.Int64{
				int64validator.AtLeast(1),
			},
		},
		"min_worker_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Worker queue min worker count",
			Required:            true,
			Validators: []validator.Int64{
				int64validator.AtLeast(0),
			},
		},
		"pod_cpu": resourceSchema.StringAttribute{
			MarkdownDescription: "Worker queue pod CPU",
			Computed:            true,
		},
		"pod_memory": resourceSchema.StringAttribute{
			MarkdownDescription: "Worker queue pod memory",
			Computed:            true,
		},
		"worker_concurrency": resourceSchema.Int64Attribute{
			MarkdownDescription: "Worker queue worker concurrency",
			Required:            true,
			Validators: []validator.Int64{
				int64validator.AtLeast(1),
			},
		},
		"astro_machine": resourceSchema.StringAttribute{
			MarkdownDescription: "Worker queue Astro machine value - required for 'STANDARD' and 'DEDICATED' deployments",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.WorkerQueueRequestAstroMachineA5),
					string(platform.WorkerQueueRequestAstroMachineA10),
					string(platform.WorkerQueueRequestAstroMachineA20),
					string(platform.WorkerQueueRequestAstroMachineA40),
					string(platform.WorkerQueueRequestAstroMachineA60),
					string(platform.WorkerQueueRequestAstroMachineA120),
					string(platform.WorkerQueueRequestAstroMachineA160),
				),
			},
		},
		"node_pool_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Worker queue Node pool identifier - required for 'HYBRID' deployments",
			Optional:            true,
			Validators: []validator.String{
				validators.IsCuid(),
			},
		},
	}
}
