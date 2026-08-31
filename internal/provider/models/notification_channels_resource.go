package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NotificationChannelsResource describes the astro_notification_channels (bulk) resource data model.
// Channels are keyed by a stable user-defined string so Terraform can track each channel's identity
// across applies while the resource batches the underlying API calls.
type NotificationChannelsResource struct {
	NotificationChannels types.Map `tfsdk:"notification_channels"`
}

// NotificationChannelsResourceElementModel describes a single notification channel within the
// astro_notification_channels map. It mirrors the input fields of the singular
// NotificationChannelResource with a computed id, omitting the read-only/expanded fields that the
// bulk resource does not track per element.
type NotificationChannelsResourceElementModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	Definition types.Object `tfsdk:"definition"`
	EntityId   types.String `tfsdk:"entity_id"`
	EntityType types.String `tfsdk:"entity_type"`
	IsShared   types.Bool   `tfsdk:"is_shared"`
}

// ReadFromResponse populates an element from a platform.NotificationChannel.
//
// desiredDefinition carries the definition the user configured (or the prior state) so that
// sensitive fields the API never returns (webhook_url, api_key, integration_key,
// deployment_api_token) are preserved in state, matching the singular resource's behaviour.
func (e *NotificationChannelsResourceElementModel) ReadFromResponse(
	ctx context.Context,
	notificationChannel *platform.NotificationChannel,
	desiredDefinition types.Object,
) diag.Diagnostics {
	var diags diag.Diagnostics

	e.Id = types.StringValue(notificationChannel.Id)
	e.Name = types.StringValue(notificationChannel.Name)
	e.Type = types.StringValue(notificationChannel.Type)
	e.EntityId = types.StringValue(notificationChannel.EntityId)
	e.EntityType = types.StringValue(notificationChannel.EntityType)
	e.IsShared = types.BoolValue(notificationChannel.IsShared)

	e.Definition, diags = NotificationChannelDefinitionResourceTypesObject(
		ctx,
		notificationChannel.Definition,
		desiredDefinition,
		notificationChannel.Type,
	)
	if diags.HasError() {
		return diags
	}

	return diags
}

// NotificationChannelsElementObjectType is the Terraform object type for a notification channel
// element in the resource map.
func NotificationChannelsElementObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: schemas.NotificationChannelsElementResourceAttributeTypes()}
}
