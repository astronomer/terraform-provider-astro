package schemas

import (
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
		"workspace_name": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment workspace name",
			Computed:            true,
		},
		"cluster_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment cluster identifier",
			Computed:            true,
		},
		"cluster_name": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment cluster name",
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
		"contact_emails": datasourceSchema.ListAttribute{
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
			MarkdownDescription: "Deployment mage tag",
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
		"environment_variables": datasourceSchema.ListNestedAttribute{
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: DeploymentEnvironmentVariableAttributes(),
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
		"webserver_cpu": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment webserver CPU",
			Computed:            true,
		},
		"webserver_memory": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment webserver memory",
			Computed:            true,
		},
		"webserver_replicas": datasourceSchema.Int64Attribute{
			MarkdownDescription: "Deployment webserver replicas",
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
		"worker_queues": datasourceSchema.ListNestedAttribute{
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: WorkerQueueSchemaAttributes(),
			},
			MarkdownDescription: "Deployment worker queues",
			Computed:            true,
		},
		"task_pod_node_pool_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment task pod node pool identifier",
			Computed:            true,
		},
		"is_cicd_enforced": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Deployment CI/CD enforced",
			Computed:            true,
		},
		"type": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment type",
			Computed:            true,
		},
		"is_dag_deploy_enabled": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Deployment DAG deploy enabled",
			Computed:            true,
		},
		"scheduler_size": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment scheduler size",
			Computed:            true,
		},
		"is_high_availability": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Deployment high availability",
			Computed:            true,
		},
		"is_development_mode": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Deployment development mode",
			Computed:            true,
		},
		"workload_identity": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment workload identity",
			Computed:            true,
		},
		"external_ips": datasourceSchema.ListAttribute{
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

func DeploymentEnvironmentVariableAttributes() map[string]datasourceSchema.Attribute {
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

func WorkerQueueAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                 types.StringType,
		"name":               types.StringType,
		"astro_machine":      types.StringType,
		"is_default":         types.BoolType,
		"max_worker_count":   types.Int64Type,
		"min_worker_count":   types.Int64Type,
		"node_pool_id":       types.StringType,
		"pod_cpu":            types.StringType,
		"pod_memory":         types.StringType,
		"worker_concurrency": types.Int64Type,
	}
}

func WorkerQueueSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"id": datasourceSchema.StringAttribute{
			Computed: true,
		},
		"name": datasourceSchema.StringAttribute{
			Computed: true,
		},
		"astro_machine": datasourceSchema.StringAttribute{
			Computed: true,
		},
		"is_default": datasourceSchema.BoolAttribute{
			Computed: true,
		},
		"max_worker_count": datasourceSchema.Int64Attribute{
			Computed: true,
		},
		"min_worker_count": datasourceSchema.Int64Attribute{
			Computed: true,
		},
		"node_pool_id": datasourceSchema.StringAttribute{
			Computed: true,
		},
		"pod_cpu": datasourceSchema.StringAttribute{
			Computed: true,
		},
		"pod_memory": datasourceSchema.StringAttribute{
			Computed: true,
		},
		"worker_concurrency": datasourceSchema.Int64Attribute{
			Computed: true,
		},
	}
}
