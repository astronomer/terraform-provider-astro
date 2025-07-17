package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func AlertsElementAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":            types.StringType,
		"name":          types.StringType,
		"type":          types.StringType,
		"rules":         types.ObjectType{AttrTypes: AlertRulesAttributeTypes()},
		"entity_id":     types.StringType,
		"entity_type":   types.StringType,
		"entity_name":   types.StringType,
		"workspace_id":  types.StringType,
		"deployment_id": types.StringType,
		"severity":      types.StringType,
		"created_at":    types.StringType,
		"updated_at":    types.StringType,
		"created_by": types.ObjectType{
			AttrTypes: SubjectProfileAttributeTypes(),
		},
		"updated_by": types.ObjectType{
			AttrTypes: SubjectProfileAttributeTypes(),
		},
	}
}

func AlertsDataSourceSchemaAttributes() map[string]schema.Attribute {
	attrs := AlertDataSourceSchemaAttributes()
	delete(attrs, "notification_channels")
	return map[string]schema.Attribute{
		"alerts": schema.SetNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: attrs,
			},
			Computed: true,
		},
		"alert_ids": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				setvalidator.ValueStringsAre(stringvalidator.All(validators.IsCuid())),
			},
		},
		"deployment_ids": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				setvalidator.ValueStringsAre(stringvalidator.All(validators.IsCuid())),
			},
		},
		"workspace_ids": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				setvalidator.ValueStringsAre(stringvalidator.All(validators.IsCuid())),
			},
		},
		"alert_types": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				setvalidator.ValueStringsAre(stringvalidator.OneOf(
					string(platform.CreateDagDurationAlertRequestTypeDAGDURATION),
					string(platform.CreateDagFailureAlertRequestTypeDAGFAILURE),
					string(platform.CreateDagSuccessAlertRequestTypeDAGSUCCESS),
					string(platform.CreateDagTimelinessAlertRequestTypeDAGTIMELINESS),
					string(platform.CreateTaskFailureAlertRequestTypeTASKFAILURE),
					string(platform.CreateTaskDurationAlertRequestTypeTASKDURATION),
				)),
			},
		},
		"entity_type": schema.StringAttribute{
			Optional: true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.AlertEntityTypeDEPLOYMENT),
				),
			},
		},
	}
}
