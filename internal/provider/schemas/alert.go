package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func AlertRulesAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"properties":      types.MapType{ElemType: types.StringType},
		"pattern_matches": types.SetType{ElemType: types.ObjectType{AttrTypes: AlertRulesPatternMatchAttributeTypes()}},
	}
}

// AlertRulesPatternMatchAttributeTypes returns the attribute types for each pattern match in AlertRules.
func AlertRulesPatternMatchAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"entity_type":   types.StringType,
		"operator_type": types.StringType,
		"values":        types.SetType{ElemType: types.StringType},
	}
}

func AlertDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Alert identifier",
			Required:            true,
			Validators:          []validator.String{validators.IsCuid()},
		},
		"name": datasourceSchema.StringAttribute{
			MarkdownDescription: "Alert name",
			Computed:            true,
		},
		"type": datasourceSchema.StringAttribute{
			MarkdownDescription: "Type of alert (e.g., 'DAG_SUCCESS', 'DAG_FAILURE')",
			Computed:            true,
		},
		"rules": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Alert rules defining the conditions for triggering the alert",
			Computed:            true,
			Attributes:          AlertRulesDataSourceSchemaAttributes(),
		},
		"entity_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Entity identifier associated with the alert",
			Computed:            true,
		},
		"entity_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "Type of entity associated with the alert (e.g., 'DEPLOYMENT')",
			Computed:            true,
		},
		"entity_name": datasourceSchema.StringAttribute{
			MarkdownDescription: "Name of the entity associated with the alert",
			Computed:            true,
		},
		"notification_channels": datasourceSchema.SetNestedAttribute{
			MarkdownDescription: "The notification channels to send alerts to",
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: NotificationChannelDataSourceSchemaAttributes(),
			},
			Computed: true,
		},
		"workspace_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Workspace identifier associated with the alert",
			Computed:            true,
		},
		"deployment_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Deployment identifier associated with the alert",
			Computed:            true,
		},
		"severity": datasourceSchema.StringAttribute{
			MarkdownDescription: "Severity level of the alert (e.g., 'INFO', 'WARNING', 'CRITICAL')",
			Computed:            true,
		},
		"created_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Alert creation timestamp",
			Computed:            true,
		},
		"updated_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Alert last updated timestamp",
			Computed:            true,
		},
		"created_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Alert creator",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
		"updated_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Alert updater",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
	}
}

func AlertRulesDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"properties": datasourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "The alert's properties used to define the alert",
			Computed:            true,
		},
		"pattern_matches": datasourceSchema.SetNestedAttribute{
			MarkdownDescription: "The alert's pattern matches to match against",
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: map[string]datasourceSchema.Attribute{
					"entity_type": datasourceSchema.StringAttribute{
						MarkdownDescription: "The type of entity to match against",
						Computed:            true,
					},
					"operator_type": datasourceSchema.StringAttribute{
						MarkdownDescription: "The type of operator to use for the pattern match",
						Computed:            true,
					},
					"values": datasourceSchema.SetAttribute{
						MarkdownDescription: "The values to match against",
						ElementType:         types.StringType,
						Computed:            true,
					},
				},
			},
			Computed: true,
		},
	}
}

func AlertResourceSchemaAttributes() map[string]resourceSchema.Attribute {
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
			MarkdownDescription: "The ID of the Deployment to which the alert is scoped",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.AlertEntityTypeDEPLOYMENT),
				),
			},
		},
		"entity_name": resourceSchema.StringAttribute{
			MarkdownDescription: "The name of the entity the alert is associated with",
			Computed:            true,
		},
		"notification_channel_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "Set of notification channel identifiers to notify when the alert is triggered",
			Required:            true,
			ElementType:         types.StringType,
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(validators.IsCuid()),
			},
		},
		"notification_channels": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "The notification channels to send alerts to",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: NotificationChannelResourceSchemaAttributes(),
			},
			Computed: true,
		},
		"workspace_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the Workspace to which the alert is scoped",
			Computed:            true,
		},
		"deployment_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the Deployment to which the alert is scoped",
			Computed:            true,
		},
		"created_at": resourceSchema.StringAttribute{
			MarkdownDescription: "Alert creation timestamp",
			Computed:            true,
		},
		"updated_at": resourceSchema.StringAttribute{
			MarkdownDescription: "Alert last updated timestamp",
			Computed:            true,
		},
		"created_by": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Alert creator",
			Computed:            true,
			Attributes:          ResourceSubjectProfileSchemaAttributes(),
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_by": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Alert updater",
			Computed:            true,
			Attributes:          ResourceSubjectProfileSchemaAttributes(),
		},
	}
}

func AlertRulesResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"properties": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The alert's properties used to define the alert",
			Required:            true,
			Attributes: map[string]resourceSchema.Attribute{
				"deployment_id": resourceSchema.StringAttribute{
					MarkdownDescription: "The ID of the deployment for the alert rule",
					Required:            true,
					Validators:          []validator.String{validators.IsCuid()},
				},
				"dag_duration_seconds": resourceSchema.Int64Attribute{
					MarkdownDescription: "The duration of the DAG in seconds (minimum 60)",
					Optional:            true,
					Validators: []validator.Int64{
						int64validator.AtLeast(60),
						int64validator.AtMost(86400),
					},
				},
				"dag_deadline": resourceSchema.StringAttribute{
					MarkdownDescription: "The deadline for the DAG in HH:MM 24-hour UTC format",
					Optional:            true,
				},
				"days_of_week": resourceSchema.SetAttribute{
					MarkdownDescription: "The days of the week for the timeliness rule",
					ElementType:         types.StringType,
					Optional:            true,
				},
				"look_back_period_seconds": resourceSchema.Int64Attribute{
					MarkdownDescription: "The look-back period in seconds (minimum 60)",
					Optional:            true,
					Validators: []validator.Int64{
						int64validator.AtLeast(60),
						int64validator.AtMost(86400),
					},
				},
				"task_duration_seconds": resourceSchema.Int64Attribute{
					MarkdownDescription: "The duration of the Task in seconds (minimum 60)",
					Optional:            true,
					Validators: []validator.Int64{
						int64validator.AtLeast(60),
						int64validator.AtMost(86400),
					},
				},
			},
		},
		"pattern_matches": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "The alert's pattern matches to match against",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: map[string]resourceSchema.Attribute{
					"entity_type": resourceSchema.StringAttribute{
						MarkdownDescription: "The type of entity to match against",
						Required:            true,
					},
					"operator_type": resourceSchema.StringAttribute{
						MarkdownDescription: "The type of operator to use for the pattern match",
						Required:            true,
					},
					"values": resourceSchema.SetAttribute{
						MarkdownDescription: "The values to match against",
						ElementType:         types.StringType,
						Required:            true,
					},
				},
			},
			Required: true,
		},
	}
}

// AlertRulesResourceAttributeTypes returns the attribute types for the resource 'rules' block.
func AlertRulesResourceAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"properties": types.ObjectType{AttrTypes: map[string]attr.Type{
			"deployment_id":            types.StringType,
			"dag_duration_seconds":     types.Int64Type,
			"dag_deadline":             types.StringType,
			"days_of_week":             types.SetType{ElemType: types.StringType},
			"look_back_period_seconds": types.Int64Type,
			"task_duration_seconds":    types.Int64Type,
		}},
		"pattern_matches": types.SetType{ElemType: types.ObjectType{AttrTypes: AlertRulesPatternMatchAttributeTypes()}},
	}
}
