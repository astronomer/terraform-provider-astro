package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NotificationChannels describes the data source data model.
type NotificationChannels struct {
	NotificationChannels   types.Set    `tfsdk:"notification_channels"`
	NotificationChannelIds types.Set    `tfsdk:"notification_channel_ids"` // query parameter
	ChannelTypes           types.Set    `tfsdk:"channel_types"`            // query parameter
	EntityType             types.String `tfsdk:"entity_type"`              // query parameter
	DeploymentIds          types.Set    `tfsdk:"deployment_ids"`           // query parameter
	WorkspaceIds           types.Set    `tfsdk:"workspace_ids"`            // query parameter
}

func (data *NotificationChannels) ReadFromResponse(
	ctx context.Context,
	notificationChannels []platform.NotificationChannel,
) diag.Diagnostics {
	values := make([]attr.Value, len(notificationChannels))
	for i, notificationChannel := range notificationChannels {
		var singleNotificationChannelData NotificationChannelDataSource
		diags := singleNotificationChannelData.ReadFromResponse(ctx, &notificationChannel)
		if diags.HasError() {
			return diags
		}

		objectValue, diags := types.ObjectValueFrom(ctx, schemas.NotificationChannelsElementAttributeTypes(), singleNotificationChannelData)
		if diags.HasError() {
			return diags
		}
		values[i] = objectValue
	}
	var diags diag.Diagnostics
	data.NotificationChannels, diags = types.SetValue(types.ObjectType{AttrTypes: schemas.NotificationChannelsElementAttributeTypes()}, values)
	if diags.HasError() {
		return diags
	}

	return nil
}
