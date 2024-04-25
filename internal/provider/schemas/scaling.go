package schemas

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
		"schedules": types.SetType{
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
		"schedules": datasourceSchema.SetNestedAttribute{
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
			Computed:            true,
			MarkdownDescription: "Whether the override is active",
		},
		"is_hibernating": datasourceSchema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "Whether the override is hibernating",
		},
		"override_until": datasourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Time until the override is active",
		},
	}
}

func HibernationScheduleDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"description": datasourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Description of the schedule",
		},
		"hibernate_at_cron": datasourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Cron expression for hibernation",
		},
		"is_enabled": datasourceSchema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "Whether the schedule is enabled",
		},
		"wake_at_cron": datasourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Cron expression for waking",
		},
	}
}

func ScalingSpecResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"hibernation_spec": resourceSchema.SingleNestedAttribute{
			Attributes: HibernationSpecResourceSchemaAttributes(),
			Optional:   true,
		},
	}
}

func HibernationSpecResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"override": resourceSchema.SingleNestedAttribute{
			Attributes: HibernationOverrideResourceSchemaAttributes(),
			Optional:   true,
		},
		"schedules": resourceSchema.SetNestedAttribute{
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: HibernationScheduleResourceSchemaAttributes(),
			},
			Validators: []validator.Set{
				setvalidator.SizeAtMost(10),
			},
			Optional: true,
		},
	}
}

func HibernationOverrideResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"is_active": resourceSchema.BoolAttribute{
			Computed: true,
		},
		"is_hibernating": resourceSchema.BoolAttribute{
			Optional: true,
			Validators: []validator.Bool{
				boolvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("override_until")),
			},
		},
		"override_until": resourceSchema.StringAttribute{
			Optional: true,
		},
	}
}

func HibernationScheduleResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"description": resourceSchema.StringAttribute{
			Optional: true,
			Validators: []validator.String{
				stringvalidator.LengthAtMost(200),
			},
		},
		"hibernate_at_cron": resourceSchema.StringAttribute{
			Required: true,
		},
		"is_enabled": resourceSchema.BoolAttribute{
			Required: true,
		},
		"wake_at_cron": resourceSchema.StringAttribute{
			Required: true,
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
			Computed:            true,
			MarkdownDescription: "Whether the deployment is hibernating",
		},
		"next_event_at": datasourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Time of the next event",
		},
		"next_event_type": datasourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Type of the next event",
		},
		"reason": datasourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Reason for the current state",
		},
	}
}

func ScalingStatusResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"hibernation_status": resourceSchema.SingleNestedAttribute{
			Attributes: HibernationStatusResourceSchemaAttributes(),
			Computed:   true,
		},
	}
}

func HibernationStatusResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"is_hibernating": resourceSchema.BoolAttribute{
			Computed: true,
		},
		"next_event_at": resourceSchema.StringAttribute{
			Computed: true,
		},
		"next_event_type": resourceSchema.StringAttribute{
			Computed: true,
		},
		"reason": resourceSchema.StringAttribute{
			Computed: true,
		},
	}
}
