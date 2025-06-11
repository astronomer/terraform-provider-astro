package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NotificationChannelDataSource describes the data source data model.
type NotificationChannelDataSource struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	ShortToken         types.String `tfsdk:"short_token"`
	Type               types.String `tfsdk:"type"`
	StartAt            types.String `tfsdk:"start_at"`
	EndAt              types.String `tfsdk:"end_at"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
	CreatedBy          types.Object `tfsdk:"created_by"`
	UpdatedBy          types.Object `tfsdk:"updated_by"`
	ExpiryPeriodInDays types.Int64  `tfsdk:"expiry_period_in_days"`
	LastUsedAt         types.String `tfsdk:"last_used_at"`
	Roles              types.Set    `tfsdk:"roles"`
}

func (data *NotificationChannelDataSource) ReadFromResponse(ctx context.Context, notificationChannel *platform.NotificationChannel) diag.Diagnostics {
	var diags diag.Diagnostics
	data.Id = types.StringValue(notificationChannel.Id)
	data.Name = types.StringValue(notificationChannel.Name)
	data.Type = types.StringValue(notificationChannel.Type)
	data.CreatedAt = types.StringValue(notificationChannel.CreatedAt)
	data.UpdatedAt = types.StringValue(notificationChannel.UpdatedAt)
	data.CreatedBy, diags = SubjectProfileTypesObject(ctx, notificationChannel.CreatedBy)
	if diags.HasError() {
		return diags
	}
	data.UpdatedBy, diags = SubjectProfileTypesObject(ctx, notificationChannel.UpdatedBy)
	if diags.HasError() {
		return diags
	}
	return diags
}
