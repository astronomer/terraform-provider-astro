package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NotificationChannelDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"id": datasourceSchema.StringAttribute{
			MarkdownDescription: "The notification channel's ID",
			Required:            true,
			Validators:          []validator.String{validators.IsCuid()},
		},
		"name": datasourceSchema.StringAttribute{
			MarkdownDescription: "The notification channel's name",
			Computed:            true,
		},
		"definition": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The notification channel's definition",
			Computed:            true,
			Attributes:          NotificationChannelDefinitionDataSourceSchemaAttributes(),
		},
		"type": datasourceSchema.StringAttribute{
			MarkdownDescription: "The notification channel's type",
			Computed:            true,
		},
		"workspace_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "The workspace ID the notification channel is scoped to",
			Computed:            true,
		},
		"deployment_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "The deployment ID the notification channel is scoped to",
			Computed:            true,
		},
		"entity_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "The entity ID the notification channel is scoped to",
			Computed:            true,
		},
		"entity_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "The type of entity the notification channel is scoped to (e.g., 'DEPLOYMENT')",
			Computed:            true,
		},
		"entity_name": datasourceSchema.StringAttribute{
			MarkdownDescription: "The name of the entity the notification channel is scoped to",
			Computed:            true,
		},
		"is_shared": datasourceSchema.BoolAttribute{
			MarkdownDescription: "When entity type is scoped to ORGANIZATION or WORKSPACE, this determines if child entities can access this notification channel.",
			Computed:            true,
		},
		"created_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Notification Channel creation timestamp",
			Computed:            true,
		},
		"updated_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Notification Channel last updated timestamp",
			Computed:            true,
		},
		"created_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Notification Channel creator",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
		"updated_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Notification Channel updater",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
	}
}

func NotificationChannelResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "The notification channel's ID",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "The notification channel's name",
			Required:            true,
		},
		"definition": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The notification channel's definition",
			Required:            true,
			Attributes:          NotificationChannelDefinitionResourceSchemaAttributes(),
			Validators: []validator.Object{
				validators.NotificationChannelDefinitionValidator(),
			},
		},
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: "The notification channel's type",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.AlertNotificationChannelTypeEMAIL),
					string(platform.AlertNotificationChannelTypeSLACK),
					string(platform.AlertNotificationChannelTypePAGERDUTY),
					string(platform.AlertNotificationChannelTypeDAGTRIGGER),
					string(platform.AlertNotificationChannelTypeOPSGENIE),
				),
			},
		},
		"workspace_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The workspace ID the notification channel is scoped to",
			Computed:            true,
		},
		"deployment_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The deployment ID the notification channel is scoped to",
			Computed:            true,
		},
		"entity_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The entity ID the notification channel is scoped to",
			Required:            true,
		},
		"entity_type": resourceSchema.StringAttribute{
			MarkdownDescription: "The type of entity the notification channel is scoped to (e.g., 'DEPLOYMENT')",
			Required:            true,
		},
		"entity_name": resourceSchema.StringAttribute{
			MarkdownDescription: "The name of the entity the notification channel is scoped to",
			Computed:            true,
		},
		"is_shared": resourceSchema.BoolAttribute{
			MarkdownDescription: "When entity type is scoped to ORGANIZATION or WORKSPACE, this determines if child entities can access this notification channel.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"created_at": resourceSchema.StringAttribute{
			MarkdownDescription: "Notification Channel creation timestamp",
			Computed:            true,
		},
		"updated_at": resourceSchema.StringAttribute{
			MarkdownDescription: "Notification Channel last updated timestamp",
			Computed:            true,
		},
		"created_by": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Notification Channel creator",
			Computed:            true,
			Attributes:          ResourceSubjectProfileSchemaAttributes(),
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_by": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Notification Channel updater",
			Computed:            true,
			Attributes:          ResourceSubjectProfileSchemaAttributes(),
		},
	}
}

func NotificationChannelDefinitionDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"dag_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "The DAG ID for the notification channel",
			Computed:            true,
		},
		"deployment_api_token": datasourceSchema.StringAttribute{
			MarkdownDescription: "The deployment API token for the notification channel",
			Computed:            true,
		},
		"deployment_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "The deployment ID for the notification channel",
			Computed:            true,
		},
		"recipients": datasourceSchema.SetAttribute{
			MarkdownDescription: "The recipients for the notification channel",
			ElementType:         types.StringType,
			Computed:            true,
		},
		"api_key": datasourceSchema.StringAttribute{
			MarkdownDescription: "The API key for the notification channel",
			Computed:            true,
		},
		"integration_key": datasourceSchema.StringAttribute{
			MarkdownDescription: "The integration key for the notification channel",
			Computed:            true,
		},
		"webhook_url": datasourceSchema.StringAttribute{
			MarkdownDescription: "The webhook URL for the notification channel",
			Computed:            true,
		},
	}
}

func NotificationChannelDefinitionResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"dag_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The DAG ID for the notification channel",
			Optional:            true,
		},
		"deployment_api_token": resourceSchema.StringAttribute{
			MarkdownDescription: "The deployment API token for the notification channel",
			Optional:            true,
			Sensitive:           true,
		},
		"deployment_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The deployment ID for the notification channel",
			Optional:            true,
		},
		"recipients": resourceSchema.SetAttribute{
			MarkdownDescription: "The recipients for the notification channel",
			ElementType:         types.StringType,
			Optional:            true,
		},
		"api_key": resourceSchema.StringAttribute{
			MarkdownDescription: "The API key for the notification channel",
			Optional:            true,
			Sensitive:           true,
		},
		"integration_key": resourceSchema.StringAttribute{
			MarkdownDescription: "The integration key for the notification channel",
			Optional:            true,
			Sensitive:           true,
		},
		"webhook_url": resourceSchema.StringAttribute{
			MarkdownDescription: "The webhook URL for the notification channel",
			Optional:            true,
			Sensitive:           true,
		},
	}
}

func NotificationChannelDefinitionAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"dag_id":               types.StringType,
		"deployment_api_token": types.StringType,
		"deployment_id":        types.StringType,
		"recipients": types.SetType{
			ElemType: types.StringType,
		},
		"api_key":         types.StringType,
		"integration_key": types.StringType,
		"webhook_url":     types.StringType,
	}
}
