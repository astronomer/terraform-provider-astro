package schemas

import (
	"context"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ClusterResourceSchemaAttributes(ctx context.Context) map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Cluster identifier",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Cluster name",
			Required:            true,
		},
		"cloud_provider": resourceSchema.StringAttribute{
			MarkdownDescription: "Cluster cloud provide - if changed, the cluster will be recreated.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.ClusterCloudProviderAWS),
					string(platform.ClusterCloudProviderGCP),
					string(platform.ClusterCloudProviderAZURE),
				),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
		"db_instance_type": resourceSchema.StringAttribute{
			MarkdownDescription: "Cluster database instance type",
			Required:            true,
		},
		"region": resourceSchema.StringAttribute{
			MarkdownDescription: "Cluster region - if changed, the cluster will be recreated.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
		"pod_subnet_range": resourceSchema.StringAttribute{
			MarkdownDescription: "Cluster pod subnet range - required for 'GCP' clusters. If changed, the cluster will be recreated.",
			Optional:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
		"service_peering_range": resourceSchema.StringAttribute{
			MarkdownDescription: "Cluster service peering range - required for 'GCP' clusters. If changed, the cluster will be recreated.",
			Optional:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
		"service_subnet_range": resourceSchema.StringAttribute{
			MarkdownDescription: "Cluster service subnet range - required for 'GCP' clusters. If changed, the cluster will be recreated.",
			Optional:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
		"vpc_subnet_range": resourceSchema.StringAttribute{
			MarkdownDescription: "Cluster VPC subnet range. If changed, the cluster will be recreated.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
		"metadata": resourceSchema.SingleNestedAttribute{
			Attributes:          ClusterMetadataResourceAttributes(),
			Computed:            true,
			MarkdownDescription: "Cluster metadata",
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
		},
		"status": resourceSchema.StringAttribute{
			MarkdownDescription: "Cluster status",
			Computed:            true,
		},
		"created_at": resourceSchema.StringAttribute{
			MarkdownDescription: "Cluster creation timestamp",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_at": resourceSchema.StringAttribute{
			MarkdownDescription: "Cluster last updated timestamp",
			Computed:            true,
		},
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: "Cluster type",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.ClusterTypeDEDICATED),
				),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
		"tenant_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Cluster tenant ID",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"provider_account": resourceSchema.StringAttribute{
			MarkdownDescription: "Cluster provider account",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"node_pools": resourceSchema.ListNestedAttribute{
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: NodePoolResourceSchemaAttributes(),
			},
			MarkdownDescription: "Cluster node pools",
			Computed:            true,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
		"workspace_ids": resourceSchema.ListAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Cluster workspace IDs",
			Required:            true,
			Validators: []validator.List{
				listvalidator.ValueStringsAre(validators.IsCuid()),
				listvalidator.UniqueValues(),
			},
		},
		"is_limited": resourceSchema.BoolAttribute{
			MarkdownDescription: "Whether the cluster is limited",
			Computed:            true,
		},
		"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
			Create: true,
			Update: true,
			Delete: true,
		}),
	}
}

func ClusterDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Cluster identifier",
			Required:            true,
			Validators:          []validator.String{validators.IsCuid()},
		},
		"name": datasourceSchema.StringAttribute{
			MarkdownDescription: "Cluster name",
			Computed:            true,
		},
		"cloud_provider": datasourceSchema.StringAttribute{
			MarkdownDescription: "Cluster cloud provider",
			Computed:            true,
		},
		"db_instance_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "Cluster database instance type",
			Computed:            true,
		},
		"region": datasourceSchema.StringAttribute{
			MarkdownDescription: "Cluster region",
			Computed:            true,
		},
		"pod_subnet_range": datasourceSchema.StringAttribute{
			MarkdownDescription: "Cluster pod subnet range",
			Computed:            true,
		},
		"service_peering_range": datasourceSchema.StringAttribute{
			MarkdownDescription: "Cluster service peering range",
			Computed:            true,
		},
		"service_subnet_range": datasourceSchema.StringAttribute{
			MarkdownDescription: "Cluster service subnet range",
			Computed:            true,
		},
		"vpc_subnet_range": datasourceSchema.StringAttribute{
			MarkdownDescription: "Cluster VPC subnet range",
			Computed:            true,
		},
		"metadata": datasourceSchema.SingleNestedAttribute{
			Attributes:          ClusterMetadataDataSourceAttributes(),
			Computed:            true,
			MarkdownDescription: "Cluster metadata",
		},
		"status": datasourceSchema.StringAttribute{
			MarkdownDescription: "Cluster status",
			Computed:            true,
		},
		"created_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Cluster creation timestamp",
			Computed:            true,
		},
		"updated_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Cluster last updated timestamp",
			Computed:            true,
		},
		"type": datasourceSchema.StringAttribute{
			MarkdownDescription: "Cluster type",
			Computed:            true,
		},
		"tenant_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Cluster tenant ID",
			Computed:            true,
		},
		"provider_account": datasourceSchema.StringAttribute{
			MarkdownDescription: "Cluster provider account",
			Computed:            true,
		},
		"node_pools": datasourceSchema.ListNestedAttribute{
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: NodePoolDataSourceSchemaAttributes(),
			},

			MarkdownDescription: "Cluster node pools",
			Computed:            true,
		},
		"workspace_ids": datasourceSchema.ListAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Cluster workspace IDs",
			Computed:            true,
		},
		"tags": datasourceSchema.ListNestedAttribute{
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: ClusterTagDataSourceAttributes(),
			},
			MarkdownDescription: "Cluster tags",
			Computed:            true,
		},
		"is_limited": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether the cluster is limited",
			Computed:            true,
		},
	}
}

func ClusterMetadataAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"external_ips":    types.ListType{ElemType: types.StringType},
		"oidc_issuer_url": types.StringType,
	}
}

func ClusterMetadataDataSourceAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"external_ips": datasourceSchema.ListAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Cluster external IPs",
			Computed:            true,
		},
		"oidc_issuer_url": datasourceSchema.StringAttribute{
			MarkdownDescription: "Cluster OIDC issuer URL",
			Computed:            true,
		},
	}
}

func ClusterMetadataResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"external_ips": resourceSchema.ListAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Cluster external IPs",
			Computed:            true,
		},
		"oidc_issuer_url": resourceSchema.StringAttribute{
			MarkdownDescription: "Cluster OIDC issuer URL",
			Computed:            true,
		},
	}
}

func ClusterTagAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"key":   types.StringType,
		"value": types.StringType,
	}
}

func ClusterTagDataSourceAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"key": datasourceSchema.StringAttribute{
			MarkdownDescription: "Cluster tag key",
			Computed:            true,
		},
		"value": datasourceSchema.StringAttribute{
			MarkdownDescription: "Cluster tag value",
			Computed:            true,
		},
	}
}

func ClusterTagResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"key": resourceSchema.StringAttribute{
			MarkdownDescription: "Cluster tag key",
			Required:            true,
		},
		"value": resourceSchema.StringAttribute{
			MarkdownDescription: "Cluster tag value",
			Required:            true,
		},
	}
}

func NodePoolAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                 types.StringType,
		"name":               types.StringType,
		"cluster_id":         types.StringType,
		"cloud_provider":     types.StringType,
		"max_node_count":     types.Int64Type,
		"node_instance_type": types.StringType,
		"is_default":         types.BoolType,
		"supported_astro_machines": types.ListType{
			ElemType: types.StringType,
		},
		"created_at": types.StringType,
		"updated_at": types.StringType,
	}
}

func NodePoolResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Node pool identifier",
			Computed:            true,
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Node pool name",
			Computed:            true,
		},
		"cluster_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Node pool cluster identifier",
			Computed:            true,
		},
		"cloud_provider": resourceSchema.StringAttribute{
			MarkdownDescription: "Node pool cloud provider",
			Computed:            true,
		},
		"max_node_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Node pool maximum node count",
			Computed:            true,
		},
		"node_instance_type": resourceSchema.StringAttribute{
			MarkdownDescription: "Node pool node instance type",
			Computed:            true,
		},
		"is_default": resourceSchema.BoolAttribute{
			MarkdownDescription: "Whether the node pool is the default node pool of the cluster",
			Computed:            true,
		},
		"supported_astro_machines": resourceSchema.ListAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Node pool supported Astro machines",
			Computed:            true,
		},
		"created_at": resourceSchema.StringAttribute{
			MarkdownDescription: "Node pool creation timestamp",
			Computed:            true,
		},
		"updated_at": resourceSchema.StringAttribute{
			MarkdownDescription: "Node pool last updated timestamp",
			Computed:            true,
		},
	}
}

func NodePoolDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Node pool identifier",
			Computed:            true,
		},
		"name": datasourceSchema.StringAttribute{
			MarkdownDescription: "Node pool name",
			Computed:            true,
		},
		"cluster_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Node pool cluster identifier",
			Computed:            true,
		},
		"cloud_provider": datasourceSchema.StringAttribute{
			MarkdownDescription: "Node pool cloud provider",
			Computed:            true,
		},
		"max_node_count": datasourceSchema.Int64Attribute{
			MarkdownDescription: "Node pool maximum node count",
			Computed:            true,
		},
		"node_instance_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "Node pool node instance type",
			Computed:            true,
		},
		"is_default": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether the node pool is the default node pool of the cluster",
			Computed:            true,
		},
		"supported_astro_machines": datasourceSchema.ListAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Node pool supported Astro machines",
			Computed:            true,
		},
		"created_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Node pool creation timestamp",
			Computed:            true,
		},
		"updated_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Node pool last updated timestamp",
			Computed:            true,
		},
	}
}
