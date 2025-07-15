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

func NotificationChannelsElementAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":              types.StringType,
		"name":            types.StringType,
		"definition":      types.ObjectType{AttrTypes: NotificationChannelDefinitionAttributeTypes()},
		"type":            types.StringType,
		"is_shared":       types.BoolType,
		"entity_id":       types.StringType,
		"entity_type":     types.StringType,
		"entity_name":     types.StringType,
		"organization_id": types.StringType,
		"workspace_id":    types.StringType,
		"deployment_id":   types.StringType,
		"created_at":      types.StringType,
		"updated_at":      types.StringType,
		"created_by": types.ObjectType{
			AttrTypes: SubjectProfileAttributeTypes(),
		},
		"updated_by": types.ObjectType{
			AttrTypes: SubjectProfileAttributeTypes(),
		},
	}
}

func NotificationChannelsDataSourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"notification_channels": schema.SetNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: NotificationChannelDataSourceSchemaAttributes(),
			},
			Computed: true,
		},
		"notification_channel_ids": schema.SetAttribute{
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
		"channel_types": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				setvalidator.ValueStringsAre(stringvalidator.OneOf(
					string(platform.AlertNotificationChannelTypeEMAIL),
					string(platform.AlertNotificationChannelTypeSLACK),
					string(platform.AlertNotificationChannelTypeOPSGENIE),
					string(platform.AlertNotificationChannelTypePAGERDUTY),
					string(platform.AlertNotificationChannelTypeDAGTRIGGER),
				)),
			},
		},
		"entity_type": schema.StringAttribute{
			Optional: true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.AlertNotificationChannelEntityTypeORGANIZATION),
					string(platform.AlertNotificationChannelEntityTypeWORKSPACE),
					string(platform.AlertNotificationChannelEntityTypeDEPLOYMENT),
				),
			},
		},
	}
}
