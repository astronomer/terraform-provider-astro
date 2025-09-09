package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NotificationChannelDefinitionValidator validates that only appropriate fields are set for each channel type
func NotificationChannelDefinitionValidator() validator.Object {
	return notificationChannelDefinitionValidator{}
}

type notificationChannelDefinitionValidator struct{}

func (v notificationChannelDefinitionValidator) Description(ctx context.Context) string {
	return "validates that definition fields match the notification channel type"
}

func (v notificationChannelDefinitionValidator) MarkdownDescription(ctx context.Context) string {
	return "validates that definition fields match the notification channel type"
}

func (v notificationChannelDefinitionValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	// Skip validation if definition is null/unknown
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Get the channel type from the parent resource
	var channelType types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("type"), &channelType)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if channelType.IsNull() || channelType.IsUnknown() {
		return
	}

	// Get definition attributes
	definitionAttrs := req.ConfigValue.Attributes()

	// Define allowed fields for each channel type
	allowedFields := map[string][]string{
		string(platform.AlertNotificationChannelTypeEMAIL): {
			"recipients",
		},
		string(platform.AlertNotificationChannelTypeSLACK): {
			"webhook_url",
		},
		string(platform.AlertNotificationChannelTypePAGERDUTY): {
			"integration_key",
		},
		string(platform.AlertNotificationChannelTypeOPSGENIE): {
			"api_key",
		},
		string(platform.AlertNotificationChannelTypeDAGTRIGGER): {
			"dag_id",
			"deployment_api_token",
			"deployment_id",
		},
	}

	channelTypeStr := channelType.ValueString()
	allowed, exists := allowedFields[channelTypeStr]
	if !exists {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid notification channel type",
			fmt.Sprintf("Unknown notification channel type: %s", channelTypeStr),
		)
		return
	}

	// Check for disallowed fields
	var invalidFields []string
	for fieldName, fieldValue := range definitionAttrs {
		// Skip null/unknown fields
		if fieldValue.IsNull() || fieldValue.IsUnknown() {
			continue
		}

		// Check if field is allowed for this channel type
		fieldAllowed := false
		for _, allowedField := range allowed {
			if fieldName == allowedField {
				fieldAllowed = true
				break
			}
		}

		if !fieldAllowed {
			invalidFields = append(invalidFields, fieldName)
		}
	}

	// Report validation errors
	if len(invalidFields) > 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid definition fields for notification channel type",
			fmt.Sprintf(
				"The following fields are not allowed for %s notification channels: %s.\n\nAllowed fields for %s: %s",
				channelTypeStr,
				strings.Join(invalidFields, ", "),
				channelTypeStr,
				strings.Join(allowed, ", "),
			),
		)
	}

	// Check for required fields
	requiredFields := map[string][]string{
		string(platform.AlertNotificationChannelTypeEMAIL): {
			"recipients",
		},
		string(platform.AlertNotificationChannelTypeSLACK): {
			"webhook_url",
		},
		string(platform.AlertNotificationChannelTypePAGERDUTY): {
			"integration_key",
		},
		string(platform.AlertNotificationChannelTypeOPSGENIE): {
			"api_key",
		},
		string(platform.AlertNotificationChannelTypeDAGTRIGGER): {
			"dag_id",
			"deployment_api_token",
			"deployment_id",
		},
	}

	required := requiredFields[channelTypeStr]
	var missingFields []string
	for _, requiredField := range required {
		if fieldValue, exists := definitionAttrs[requiredField]; !exists || fieldValue.IsNull() {
			missingFields = append(missingFields, requiredField)
		}
	}

	if len(missingFields) > 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing required definition fields",
			fmt.Sprintf(
				"The following required fields are missing for %s notification channels: %s",
				channelTypeStr,
				strings.Join(missingFields, ", "),
			),
		)
	}
}
