package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Alerts describes the data source data model.
type Alerts struct {
	Alerts        types.Set    `tfsdk:"alerts"`
	AlertIds      types.Set    `tfsdk:"alert_ids"`      // query parameter
	AlertTypes    types.Set    `tfsdk:"alert_types"`    // query parameter
	EntityType    types.String `tfsdk:"entity_type"`    // query parameter
	DeploymentIds types.Set    `tfsdk:"deployment_ids"` // query parameter
	WorkspaceIds  types.Set    `tfsdk:"workspace_ids"`  // query parameter
}

func (data *Alerts) ReadFromResponse(
	ctx context.Context,
	alerts []platform.Alert,
) diag.Diagnostics {
	values := make([]attr.Value, len(alerts))
	for i, alert := range alerts {
		var singleAlertData AlertListModel
		diags := singleAlertData.ReadFromAlertListResponse(ctx, &alert)
		if diags.HasError() {
			return diags
		}

		objectValue, diags := types.ObjectValueFrom(ctx, schemas.AlertsElementAttributeTypes(), singleAlertData)
		if diags.HasError() {
			return diags
		}
		values[i] = objectValue
	}
	var diags diag.Diagnostics
	data.Alerts, diags = types.SetValue(types.ObjectType{AttrTypes: schemas.AlertsElementAttributeTypes()}, values)
	if diags.HasError() {
		return diags
	}

	return nil
}
