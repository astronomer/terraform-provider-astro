package schemas

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ScalingSpec
func ScalingSpecAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"hibernation_spec": types.ObjectType{
			AttrTypes: HibernationSpecAttributeTypes(),
		},
	}
}

func HibernationSpecAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"override": types.ObjectType{
			AttrTypes: HibernationOverrideAttributeTypes(),
		},
		"schedules": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: HibernationScheduleAttributeTypes(),
			},
		},
	}
}

func ScalingSpecDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"hibernation_spec": datasourceSchema.SingleNestedAttribute{
			Attributes: HibernationSpecDataSourceSchemaAttributes(),
			Computed:   true,
		},
	}
}

func HibernationSpecDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"override": datasourceSchema.SingleNestedAttribute{
			Attributes: HibernationOverrideDataSourceSchemaAttributes(),
			Computed:   true,
		},
		"schedules": datasourceSchema.ListNestedAttribute{
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: HibernationScheduleDataSourceSchemaAttributes(),
			},
			Computed: true,
		},
	}
}

func HibernationOverrideDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"is_active": datasourceSchema.BoolAttribute{
			Computed: true,
		},
		"is_hibernating": datasourceSchema.BoolAttribute{
			Computed: true,
		},
		"override_until": datasourceSchema.StringAttribute{
			Computed: true,
		},
	}
}

func HibernationScheduleDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"description": datasourceSchema.StringAttribute{
			Computed: true,
		},
		"hibernate_at_cron": datasourceSchema.StringAttribute{
			Computed: true,
		},
		"is_enabled": datasourceSchema.BoolAttribute{
			Computed: true,
		},
		"wake_at_cron": datasourceSchema.StringAttribute{
			Computed: true,
		},
	}
}

// ScalingStatus
func ScalingStatusAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"hibernation_status": types.ObjectType{
			AttrTypes: HibernationStatusAttributeTypes(),
		},
	}
}

func HibernationStatusAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"is_hibernating":  types.BoolType,
		"next_event_at":   types.StringType,
		"next_event_type": types.StringType,
		"reason":          types.StringType,
	}
}

func HibernationOverrideAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"is_active":      types.BoolType,
		"is_hibernating": types.BoolType,
		"override_until": types.StringType,
	}
}

func HibernationScheduleAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"description":       types.StringType,
		"hibernate_at_cron": types.StringType,
		"is_enabled":        types.BoolType,
		"wake_at_cron":      types.StringType,
	}
}

func ScalingStatusDataSourceAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"hibernation_status": datasourceSchema.SingleNestedAttribute{
			Attributes: HibernationStatusDataSourceSchemaAttributes(),
			Computed:   true,
		},
	}
}

func HibernationStatusDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"is_hibernating": datasourceSchema.BoolAttribute{
			Computed: true,
		},
		"next_event_at": datasourceSchema.StringAttribute{
			Computed: true,
		},
		"next_event_type": datasourceSchema.StringAttribute{
			Computed: true,
		},
		"reason": datasourceSchema.StringAttribute{
			Computed: true,
		},
	}
}
