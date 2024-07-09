package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ApiTokenDataSource describes the data source data model.
type ApiTokenDataSource struct {
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

func (data *ApiTokenDataSource) ReadFromResponse(ctx context.Context, apiToken *iam.ApiToken) diag.Diagnostics {
	var diags diag.Diagnostics
	data.Id = types.StringValue(apiToken.Id)
	data.Name = types.StringValue(apiToken.Name)
	data.Description = types.StringValue(apiToken.Description)
	data.ShortToken = types.StringValue(apiToken.ShortToken)
	data.Type = types.StringValue(string(apiToken.Type))
	data.StartAt = types.StringValue(apiToken.StartAt.String())
	if apiToken.EndAt != nil {
		data.EndAt = types.StringValue(apiToken.EndAt.String())
	} else {
		data.EndAt = types.StringValue("")
	}
	data.CreatedAt = types.StringValue(apiToken.CreatedAt.String())
	data.UpdatedAt = types.StringValue(apiToken.UpdatedAt.String())
	data.CreatedBy, diags = SubjectProfileTypesObject(ctx, apiToken.CreatedBy)
	if diags.HasError() {
		return diags
	}
	data.UpdatedBy, diags = SubjectProfileTypesObject(ctx, apiToken.UpdatedBy)
	if diags.HasError() {
		return diags
	}
	if apiToken.ExpiryPeriodInDays != nil {
		data.ExpiryPeriodInDays = types.Int64Value(int64(*apiToken.ExpiryPeriodInDays))
	} else {
		data.ExpiryPeriodInDays = types.Int64Value(0)
	}
	if apiToken.LastUsedAt != nil {
		data.LastUsedAt = types.StringValue(apiToken.LastUsedAt.String())
	} else {
		data.LastUsedAt = types.StringValue("")
	}
	data.Roles, diags = utils.ObjectSet(ctx, apiToken.Roles, schemas.ApiTokenRoleAttributeTypes(), ApiTokenRoleTypesObject)
	if diags.HasError() {
		return diags
	}
	return diags
}
