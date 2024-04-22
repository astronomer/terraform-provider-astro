package schemas

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//DefaultRegion              ProviderRegionResponse         `json:"defaultRegion,required" validate:"required"`
//DefaultNodeInstance        ProviderInstanceTypeResponse   `json:"defaultNodeInstance,required" validate:"required"`
//DefaultDatabaseInstance    ProviderInstanceTypeResponse   `json:"defaultDatabaseInstance,required" validate:"required"`
//NodeInstances              []ProviderInstanceTypeResponse `json:"nodeInstances,required" validate:"required"`
//DatabaseInstances          []ProviderInstanceTypeResponse `json:"databaseInstances,required" validate:"required"`
//TemplateVersions           []TemplateVersionResponse      `json:"templateVersions,required" validate:"required"`
//Regions                    []ProviderRegionResponse       `json:"regions,required" validate:"required"`

func ClusterOptionsElementAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"provider":                      types.StringType,
		"default_vpc_subnet_range":      types.StringType,
		"default_pod_subnet_range":      types.StringType,
		"default_service_subnet_range":  types.StringType,
		"default_service_peering_range": types.StringType,
		"node_count_min":                types.Int64Type,
		"node_count_max":                types.Int64Type,
		"node_count_default":            types.Int64Type,
		"default_region": types.ObjectType{
			AttrTypes: DefaultRegionAttributeTypes(),
		},
	}
}

func DefaultRegionAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":    types.StringType,
		"limited": types.BoolType,
		//"banned_instances": types.ListType{
		//	ElemType: types.StringType,
		//},
	}
}

func ClusterOptionsDataSourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"cluster_options": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: ClusterOptionDataSourceSchemaAttributes(),
			},
			Computed: true,
		},
		"type": schema.StringAttribute{
			Optional: true,
		},
		"cloud_provider": schema.StringAttribute{
			Optional: true,
		},
	}
}

func ClusterOptionDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"provider": datasourceSchema.StringAttribute{
			MarkdownDescription: "ClusterOption provider",
			Computed:            true,
		},
		"default_vpc_subnet_range": datasourceSchema.StringAttribute{
			MarkdownDescription: "ClusterOption default vps subnet range",
			Computed:            true,
		},
		"default_pod_subnet_range": datasourceSchema.StringAttribute{
			MarkdownDescription: "ClusterOption default pod subnet range",
			Computed:            true,
		},
		"default_service_subnet_range": datasourceSchema.StringAttribute{
			MarkdownDescription: "ClusterOption default service subnet range",
			Computed:            true,
		},
		"default_service_peering_range": datasourceSchema.StringAttribute{
			MarkdownDescription: "ClusterOption default service peering range",
			Computed:            true,
		},
		"node_count_min": datasourceSchema.Int64Attribute{
			MarkdownDescription: "ClusterOption node count min",
			Computed:            true,
		},
		"node_count_max": datasourceSchema.Int64Attribute{
			MarkdownDescription: "ClusterOption node count max",
			Computed:            true,
		},
		"node_count_default": datasourceSchema.Int64Attribute{
			MarkdownDescription: "ClusterOption node count default",
			Computed:            true,
		},
		"default_region": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "ClusterOption default region",
			Computed:            true,
			Attributes:          DatasourceDefaultRegionAttributes(),
		},
	}
}

func DatasourceDefaultRegionAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"name": datasourceSchema.StringAttribute{
			Computed: true,
		},
		"limited": datasourceSchema.StringAttribute{
			Computed: true,
		},
		//"banned_instances": datasourceSchema.ListAttribute{
		//	ElementType:         types.StringType,
		//	MarkdownDescription: "Default region banned instances",
		//	Computed:            true,
		//	Optional:            true,
		//},
	}
}
