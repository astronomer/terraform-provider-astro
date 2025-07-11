package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func DeploymentOptionsDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"executors": datasourceSchema.SetAttribute{
			MarkdownDescription: "Available executors",
			ElementType:         types.StringType,
			Computed:            true,
		},
		"resource_quotas": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Resource quota options",
			Attributes:          ResourceQuotaOptionsDataSourceSchemaAttributes(),
			Computed:            true,
		},
		"runtime_releases": datasourceSchema.SetAttribute{
			MarkdownDescription: "Available Astro Runtime versions",
			ElementType: types.ObjectType{
				AttrTypes: RuntimeReleaseAttributeTypes(),
			},
			Computed: true,
		},
		"scheduler_machines": datasourceSchema.SetAttribute{
			MarkdownDescription: "Available scheduler sizes",
			ElementType: types.ObjectType{
				AttrTypes: SchedulerMachineAttributeTypes(),
			},
			Computed: true,
		},
		"worker_machines": datasourceSchema.SetAttribute{
			MarkdownDescription: "Available worker machine types",
			ElementType: types.ObjectType{
				AttrTypes: WorkerMachineAttributeTypes(),
			},
			Computed: true,
		},
		"worker_queues": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Available worker queue options",
			Attributes:          WorkerQueueOptionsDataSourceSchemaAttributes(),
			Computed:            true,
		},
		"workload_identity_options": datasourceSchema.SetAttribute{
			MarkdownDescription: "Available workload identity options",
			ElementType: types.ObjectType{
				AttrTypes: WorkloadIdentityOptionsAttributeTypes(),
			},
			Computed: true,
		},
		"deployment_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment ID",
			Optional:            true,
			Validators: []validator.String{
				validators.IsCuid(),
			},
		},
		"deployment_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment type",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.DeploymentTypeHYBRID),
					string(platform.DeploymentTypeDEDICATED),
					string(platform.DeploymentTypeSTANDARD),
				),
			},
		},
		"executor": datasourceSchema.StringAttribute{
			MarkdownDescription: "Executor. Valid values: CELERY, KUBERNETES, ASTRO.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.DeploymentExecutorCELERY),
					string(platform.DeploymentExecutorKUBERNETES),
					string(platform.DeploymentExecutorASTRO),
				),
			},
		},
		"cloud_provider": datasourceSchema.StringAttribute{
			MarkdownDescription: "Cloud provider",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.DeploymentCloudProviderAWS),
					string(platform.DeploymentCloudProviderAZURE),
					string(platform.DeploymentCloudProviderGCP),
				),
			},
		},
	}
}

func ResourceQuotaOptionsDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"resource_quota": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Resource quota options",
			Attributes:          ResourceOptionDataSourceSchemaAttributes(),
			Computed:            true,
		},
		"default_pod_size": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Default pod size options",
			Attributes:          ResourceOptionDataSourceSchemaAttributes(),
			Computed:            true,
		},
	}
}

func ResourceOptionDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"cpu": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "CPU resource range",
			Attributes:          ResourceRangeDataSourceSchemaAttributes(),
			Computed:            true,
		},
		"memory": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Memory resource range",
			Attributes:          ResourceRangeDataSourceSchemaAttributes(),
			Computed:            true,
		},
	}
}

func ResourceRangeDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"floor": datasourceSchema.StringAttribute{
			MarkdownDescription: "Resource range floor",
			Computed:            true,
		},
		"default": datasourceSchema.StringAttribute{
			MarkdownDescription: "Resource range default",
			Computed:            true,
		},
		"ceiling": datasourceSchema.StringAttribute{
			MarkdownDescription: "Resource range ceiling",
			Computed:            true,
		},
	}
}

func RuntimeReleaseDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"version": datasourceSchema.StringAttribute{
			MarkdownDescription: "Runtime release version",
			Computed:            true,
		},
		"airflow_version": datasourceSchema.StringAttribute{
			MarkdownDescription: "Runtime release Airflow version",
			Computed:            true,
		},
		"release_date": datasourceSchema.StringAttribute{
			MarkdownDescription: "Runtime release date",
			Computed:            true,
		},
		"airflow_database_migration": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether Airflow database migration is required",
			Computed:            true,
		},
		"stellar_database_migration": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether Stellar database migration is required",
			Computed:            true,
		},
		"channel": datasourceSchema.StringAttribute{
			MarkdownDescription: "Runtime release channel",
			Computed:            true,
		},
	}
}

func SchedulerMachineDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"name": datasourceSchema.StringAttribute{
			MarkdownDescription: "Scheduler machine name",
			Computed:            true,
		},
		"spec": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Scheduler machine spec",
			Attributes:          MachineSpecDataSourceSchemaAttributes(),
			Computed:            true,
		},
	}
}

func WorkerMachineDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"name": datasourceSchema.StringAttribute{
			MarkdownDescription: "Worker machine name",
			Computed:            true,
		},
		"spec": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Worker machine spec",
			Attributes:          MachineSpecDataSourceSchemaAttributes(),
			Computed:            true,
		},
		"concurrency": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Worker machine concurrency",
			Attributes:          WorkerQueueOptionsDataSourceSchemaAttributes(),
			Computed:            true,
		},
	}
}

func MachineSpecDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"cpu": datasourceSchema.StringAttribute{
			MarkdownDescription: "Machine spec CPU",
			Computed:            true,
		},
		"memory": datasourceSchema.StringAttribute{
			MarkdownDescription: "Machine spec memory",
			Computed:            true,
		},
		"ephemeral_storage": datasourceSchema.StringAttribute{
			MarkdownDescription: "Machine spec ephemeral storage",
			Computed:            true,
		},
		"concurrency": datasourceSchema.StringAttribute{
			MarkdownDescription: "Machine spec concurrency",
			Computed:            true,
		},
	}
}

func WorkerQueueOptionsDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"min_workers": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Worker queue minimum workers",
			Attributes:          ResourceRangeDataSourceSchemaAttributes(),
			Computed:            true,
		},
		"max_workers": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Worker queue maximum workers",
			Attributes:          ResourceRangeDataSourceSchemaAttributes(),
			Computed:            true,
		},
		"worker_concurrency": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Worker queue worker concurrency",
			Attributes:          ResourceRangeDataSourceSchemaAttributes(),
			Computed:            true,
		},
	}
}

func WorkloadIdentityOptionsDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"role": datasourceSchema.StringAttribute{
			MarkdownDescription: "Workload identity role",
			Computed:            true,
		},
		"label": datasourceSchema.StringAttribute{
			MarkdownDescription: "Workload identity label",
			Computed:            true,
		},
	}
}

func ResourceQuotaOptionsAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"resource_quota": types.ObjectType{
			AttrTypes: ResourceOptionAttributeTypes(),
		},
		"default_pod_size": types.ObjectType{
			AttrTypes: ResourceOptionAttributeTypes(),
		},
	}
}

func ResourceOptionAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"cpu": types.ObjectType{
			AttrTypes: ResourceRangeAttributeTypes(),
		},
		"memory": types.ObjectType{
			AttrTypes: ResourceRangeAttributeTypes(),
		},
	}
}

func ResourceRangeAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"floor":   types.StringType,
		"default": types.StringType,
		"ceiling": types.StringType,
	}
}

func RuntimeReleaseAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"version":                    types.StringType,
		"airflow_version":            types.StringType,
		"release_date":               types.StringType,
		"airflow_database_migration": types.BoolType,
		"stellar_database_migration": types.BoolType,
		"channel":                    types.StringType,
	}
}

func SchedulerMachineAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name": types.StringType,
		"spec": types.ObjectType{
			AttrTypes: MachineSpecAttributeTypes(),
		},
	}
}

func WorkerMachineAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name": types.StringType,
		"spec": types.ObjectType{
			AttrTypes: MachineSpecAttributeTypes(),
		},
		"concurrency": types.ObjectType{
			AttrTypes: ResourceRangeAttributeTypes(),
		},
	}

}

func MachineSpecAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"cpu":               types.StringType,
		"memory":            types.StringType,
		"ephemeral_storage": types.StringType,
		"concurrency":       types.StringType,
	}
}

func WorkerQueueOptionsAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"min_workers": types.ObjectType{
			AttrTypes: ResourceRangeAttributeTypes(),
		},
		"max_workers": types.ObjectType{
			AttrTypes: ResourceRangeAttributeTypes(),
		},
		"worker_concurrency": types.ObjectType{
			AttrTypes: ResourceRangeAttributeTypes(),
		},
	}
}

func WorkloadIdentityOptionsAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"role":  types.StringType,
		"label": types.StringType,
	}
}
