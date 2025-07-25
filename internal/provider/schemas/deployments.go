package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func DeploymentsElementAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":          types.StringType,
		"name":        types.StringType,
		"description": types.StringType,
		"created_at":  types.StringType,
		"updated_at":  types.StringType,
		"created_by": types.ObjectType{
			AttrTypes: SubjectProfileAttributeTypes(),
		},
		"updated_by": types.ObjectType{
			AttrTypes: SubjectProfileAttributeTypes(),
		},
		"workspace_id":          types.StringType,
		"cluster_id":            types.StringType,
		"region":                types.StringType,
		"cloud_provider":        types.StringType,
		"astro_runtime_version": types.StringType,
		"airflow_version":       types.StringType,
		"namespace":             types.StringType,
		"contact_emails": types.SetType{
			ElemType: types.StringType,
		},
		"executor":           types.StringType,
		"scheduler_au":       types.Int64Type,
		"scheduler_cpu":      types.StringType,
		"scheduler_memory":   types.StringType,
		"scheduler_replicas": types.Int64Type,
		"image_tag":          types.StringType,
		"image_repository":   types.StringType,
		"image_version":      types.StringType,
		"environment_variables": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: DeploymentEnvironmentVariableAttributeTypes(),
			},
		},
		"webserver_ingress_hostname":  types.StringType,
		"webserver_url":               types.StringType,
		"webserver_airflow_api_url":   types.StringType,
		"status":                      types.StringType,
		"status_reason":               types.StringType,
		"dag_tarball_version":         types.StringType,
		"desired_dag_tarball_version": types.StringType,
		"worker_queues": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: WorkerQueueDataSourceAttributeTypes(),
			},
		},
		"task_pod_node_pool_id": types.StringType,
		"is_cicd_enforced":      types.BoolType,
		"type":                  types.StringType,
		"is_dag_deploy_enabled": types.BoolType,
		"scheduler_size":        types.StringType,
		"is_high_availability":  types.BoolType,
		"is_development_mode":   types.BoolType,
		"workload_identity":     types.StringType,
		"external_ips": types.SetType{
			ElemType: types.StringType,
		},
		"oidc_issuer_url":         types.StringType,
		"resource_quota_cpu":      types.StringType,
		"resource_quota_memory":   types.StringType,
		"default_task_pod_cpu":    types.StringType,
		"default_task_pod_memory": types.StringType,
		"scaling_status": types.ObjectType{
			AttrTypes: ScalingStatusAttributeTypes(),
		},
		"scaling_spec": types.ObjectType{
			AttrTypes: ScalingSpecAttributeTypes(),
		},
		"remote_execution": types.ObjectType{
			AttrTypes: RemoteExecutionAttributeTypes(),
		},
	}
}

func DeploymentsDataSourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"deployments": schema.SetNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: DeploymentDataSourceSchemaAttributes(),
			},
			Computed: true,
		},
		"deployment_ids": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(validators.IsCuid()),
			},
		},
		"workspace_ids": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(validators.IsCuid()),
			},
		},
		"names": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
	}
}
