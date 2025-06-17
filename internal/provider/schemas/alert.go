package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
			Attributes:          DataSourceAlertRulesSchemaAttributes(),
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
		"notification_channels": datasourceSchema.SetAttribute{
			MarkdownDescription: "The notification channels to send alerts to",
			ElementType:         types.StringType,
			Computed:            true,
		},
		"organization_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Organization identifier associated with the alert",
			Computed:            true,
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

func DataSourceAlertRulesSchemaAttributes() map[string]datasourceSchema.Attribute {
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
