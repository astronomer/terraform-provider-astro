package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AlertsResourceSchemaAttributes returns the top-level attributes for the astro_alerts (bulk)
// resource. Alerts are keyed by a user-defined string so Terraform can track each alert's identity
// across applies while the resource batches the underlying API calls (chunking when the number of
// alerts exceeds the API's per-request limit).
func AlertsResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"alerts": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "A map of alerts to manage as a single resource, keyed by a stable user-defined string.",
			Required:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: AlertsElementResourceSchemaAttributes(),
			},
		},
	}
}

// AlertsElementResourceSchemaAttributes returns the attributes for a single alert within the
// astro_alerts map. It mirrors the input fields of the singular astro_alert resource, with a
// computed id per element.
func AlertsElementResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Alert identifier",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Alert name",
			Required:            true,
		},
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: "The alert's type",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.CreateDagDurationAlertRequestTypeDAGDURATION),
					string(platform.CreateDagFailureAlertRequestTypeDAGFAILURE),
					string(platform.CreateDagSuccessAlertRequestTypeDAGSUCCESS),
					string(platform.CreateDagTimelinessAlertRequestTypeDAGTIMELINESS),
					string(platform.CreateTaskFailureAlertRequestTypeTASKFAILURE),
					string(platform.CreateTaskDurationAlertRequestTypeTASKDURATION),
				),
			},
		},
		"rules": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Alert rules defining the conditions for triggering the alert",
			Required:            true,
			Attributes:          AlertRulesResourceSchemaAttributes(),
		},
		"severity": resourceSchema.StringAttribute{
			MarkdownDescription: "The alert's severity",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.AlertSeverityINFO),
					string(platform.AlertSeverityWARNING),
					string(platform.AlertSeverityCRITICAL),
				),
			},
		},
		"entity_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The entity ID the alert is associated with",
			Required:            true,
			Validators: []validator.String{
				validators.IsCuid(),
			},
		},
		"entity_type": resourceSchema.StringAttribute{
			MarkdownDescription: "The type of entity the alert is scoped to",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.AlertEntityTypeDEPLOYMENT),
				),
			},
		},
		"notification_channel_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of notification channel identifiers to notify when the alert is triggered",
			Required:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(validators.IsCuid()),
			},
		},
	}
}

// AlertsElementResourceAttributeTypes returns the attribute types for a single alert object in the
// astro_alerts resource map, used when building the types.Map value.
func AlertsElementResourceAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                       types.StringType,
		"name":                     types.StringType,
		"type":                     types.StringType,
		"rules":                    types.ObjectType{AttrTypes: AlertRulesResourceAttributeTypes()},
		"severity":                 types.StringType,
		"entity_id":                types.StringType,
		"entity_type":              types.StringType,
		"notification_channel_ids": types.SetType{ElemType: types.StringType},
	}
}
