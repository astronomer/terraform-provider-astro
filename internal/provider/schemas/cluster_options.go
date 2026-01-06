package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
			AttrTypes: RegionAttributeTypes(),
		},
		"regions": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: RegionAttributeTypes(),
			},
		},
		"default_node_instance": types.ObjectType{
			AttrTypes: ProviderInstanceAttributeTypes(),
		},
		"node_instances": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: ProviderInstanceAttributeTypes(),
			},
		},
		"default_database_instance": types.ObjectType{
			AttrTypes: ProviderInstanceAttributeTypes(),
		},
		"database_instances": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: ProviderInstanceAttributeTypes(),
			},
		},
	}
}

func RegionAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":    types.StringType,
		"limited": types.BoolType,
		"banned_instances": types.SetType{
			ElemType: types.StringType,
		},
	}
}

func ProviderInstanceAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":   types.StringType,
		"memory": types.StringType,
		"cpu":    types.Int64Type,
	}
}

func TemplateVersionAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"version": types.StringType,
		"url":     types.StringType,
	}
}

func ClusterOptionsDataSourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"cluster_options": schema.SetNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: ClusterOptionDataSourceSchemaAttributes(),
			},
			Computed: true,
		},
		"type": schema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.ClusterTypeDEDICATED),
					string(platform.ClusterTypeHYBRID),
				),
			},
		},
		"cloud_provider": schema.StringAttribute{
			MarkdownDescription: "ClusterOptions cloud provider. Allowed values: `AWS`, `GCP`, `AZURE`.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.ClusterCloudProviderAWS),
					string(platform.ClusterCloudProviderGCP),
					string(platform.ClusterCloudProviderAZURE),
				),
			},
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
			Attributes:          DatasourceRegionAttributes(),
		},
		"regions": datasourceSchema.SetNestedAttribute{
			MarkdownDescription: "ClusterOption regions",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: DatasourceRegionAttributes(),
			},
		},
		"default_node_instance": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "ClusterOption default node instance",
			Computed:            true,
			Attributes:          DatasourceProviderInstanceAttributes(),
		},
		"node_instances": datasourceSchema.SetNestedAttribute{
			MarkdownDescription: "ClusterOption node instances",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: DatasourceProviderInstanceAttributes(),
			},
		},
		"default_database_instance": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "ClusterOption default database instance",
			Computed:            true,
			Attributes:          DatasourceProviderInstanceAttributes(),
		},
		"database_instances": datasourceSchema.SetNestedAttribute{
			MarkdownDescription: "ClusterOption database instances",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: DatasourceProviderInstanceAttributes(),
			},
		},
	}
}

func DatasourceRegionAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"name": datasourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Region is limited bool",
		},
		"limited": datasourceSchema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "Region is limited bool",
		},
		"banned_instances": datasourceSchema.SetAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Region banned instances",
			Computed:            true,
		},
	}
}

func DatasourceProviderInstanceAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"name": datasourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Provider instance name",
		},
		"cpu": datasourceSchema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Provider instance cpu",
		},
		"memory": datasourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Provider instance memory",
		},
	}
}
