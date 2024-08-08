package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ClustersElementAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":               types.StringType,
		"name":             types.StringType,
		"cloud_provider":   types.StringType,
		"db_instance_type": types.StringType,
		"health_status": types.ObjectType{
			AttrTypes: ClusterHealthStatusAttributeTypes(),
		},
		"region":                types.StringType,
		"pod_subnet_range":      types.StringType,
		"service_peering_range": types.StringType,
		"service_subnet_range":  types.StringType,
		"vpc_subnet_range":      types.StringType,
		"metadata": types.ObjectType{
			AttrTypes: ClusterMetadataAttributeTypes(),
		},
		"status":           types.StringType,
		"created_at":       types.StringType,
		"updated_at":       types.StringType,
		"type":             types.StringType,
		"tenant_id":        types.StringType,
		"provider_account": types.StringType,
		"node_pools": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: NodePoolAttributeTypes(),
			},
		},
		"workspace_ids": types.SetType{
			ElemType: types.StringType,
		},
		"tags": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: ClusterTagAttributeTypes(),
			},
		},
		"is_limited": types.BoolType,
	}
}

func ClustersDataSourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"clusters": schema.SetNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: ClusterDataSourceSchemaAttributes(),
			},
			Computed: true,
		},
		"cloud_provider": schema.StringAttribute{
			Optional: true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.ClusterCloudProviderAWS),
					string(platform.ClusterCloudProviderGCP),
					string(platform.ClusterCloudProviderAZURE),
				),
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
