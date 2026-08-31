package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NotificationChannelsResourceSchemaAttributes returns the top-level attributes for the
// astro_notification_channels (bulk) resource. Channels are keyed by a user-defined string so
// Terraform can track each channel's identity across applies while the resource batches the
// underlying API calls (chunking when the number of channels exceeds the API's per-request limit).
func NotificationChannelsResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"notification_channels": resourceSchema.MapNestedAttribute{
			MarkdownDescription: "A map of notification channels to manage as a single resource, keyed by a stable user-defined string.",
			Required:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: NotificationChannelsElementResourceSchemaAttributes(),
			},
		},
	}
}

// NotificationChannelsElementResourceSchemaAttributes returns the attributes for a single
// notification channel within the astro_notification_channels map. It mirrors the input fields of
// the singular astro_notification_channel resource, with a computed id per element.
func NotificationChannelsElementResourceSchemaAttributes() map[string]resourceSchema.Attribute {
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
		"entity_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The entity ID the notification channel is scoped to",
			Required:            true,
		},
		"entity_type": resourceSchema.StringAttribute{
			MarkdownDescription: "The type of entity the notification channel is scoped to (e.g., 'DEPLOYMENT')",
			Required:            true,
		},
		"is_shared": resourceSchema.BoolAttribute{
			MarkdownDescription: "When entity type is scoped to ORGANIZATION or WORKSPACE, this determines if child entities can access this notification channel.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
	}
}

// NotificationChannelsElementResourceAttributeTypes returns the attribute types for a single
// notification channel object in the astro_notification_channels resource map, used when building
// the types.Map value.
func NotificationChannelsElementResourceAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":          types.StringType,
		"name":        types.StringType,
		"type":        types.StringType,
		"definition":  types.ObjectType{AttrTypes: NotificationChannelDefinitionAttributeTypes()},
		"entity_id":   types.StringType,
		"entity_type": types.StringType,
		"is_shared":   types.BoolType,
	}
}
