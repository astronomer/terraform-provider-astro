package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/labs"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AlertsResource describes the astro_alerts (bulk) resource data model. Alerts are keyed by a
// stable user-defined string.
type AlertsResource struct {
	Alerts types.Map `tfsdk:"alerts"`
}

// AlertsResourceElementModel describes a single alert within the astro_alerts map. It mirrors the
// input fields of AlertResource with a computed id, omitting the read-only/expanded fields that the
// bulk resource does not track per element.
type AlertsResourceElementModel struct {
	Id                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Type                   types.String `tfsdk:"type"`
	Rules                  types.Object `tfsdk:"rules"`
	Severity               types.String `tfsdk:"severity"`
	EntityId               types.String `tfsdk:"entity_id"`
	EntityType             types.String `tfsdk:"entity_type"`
	NotificationChannelIds types.Set    `tfsdk:"notification_channel_ids"`
}

// ToAlertResource adapts an element to the richer AlertResource model so the shared request
// builders (BuildCreateAlertRequest/BuildUpdateAlertRequest) can be reused.
func (e AlertsResourceElementModel) ToAlertResource() AlertResource {
	return AlertResource{
		Id:                     e.Id,
		Name:                   e.Name,
		Type:                   e.Type,
		Rules:                  e.Rules,
		Severity:               e.Severity,
		EntityId:               e.EntityId,
		EntityType:             e.EntityType,
		NotificationChannelIds: e.NotificationChannelIds,
	}
}

// ReadFromResponse populates an element from a labs.Alert.
func (e *AlertsResourceElementModel) ReadFromResponse(ctx context.Context, alert *labs.Alert) diag.Diagnostics {
	var diags diag.Diagnostics
	e.Id = types.StringValue(alert.Id)
	e.Name = types.StringValue(alert.Name)
	e.Type = types.StringValue(string(alert.Type))
	e.Severity = types.StringValue(string(alert.Severity))
	e.EntityId = types.StringValue(alert.EntityId)
	e.EntityType = types.StringValue(string(alert.EntityType))

	e.Rules, diags = AlertRulesResourceTypesObject(ctx, alert.Rules)
	if diags.HasError() {
		return diags
	}

	var notificationChannelIds []attr.Value
	if alert.NotificationChannels != nil {
		for _, nc := range *alert.NotificationChannels {
			notificationChannelIds = append(notificationChannelIds, types.StringValue(nc.Id))
		}
	}
	e.NotificationChannelIds, diags = types.SetValue(types.StringType, notificationChannelIds)
	if diags.HasError() {
		return diags
	}

	return nil
}

// AlertsElementObjectType is the Terraform object type for an alert element in the resource map.
func AlertsElementObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: schemas.AlertsElementResourceAttributeTypes()}
}
