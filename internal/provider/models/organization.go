package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Organization describes the data source data model.
type Organization struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	SupportPlan    types.String `tfsdk:"support_plan"`
	Product        types.String `tfsdk:"product"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	CreatedBy      types.Object `tfsdk:"created_by"`
	UpdatedBy      types.Object `tfsdk:"updated_by"`
	TrialExpiresAt types.String `tfsdk:"trial_expires_at"`
	Status         types.String `tfsdk:"status"`
	PaymentMethod  types.String `tfsdk:"payment_method"`
	IsScimEnabled  types.Bool   `tfsdk:"is_scim_enabled"`
	BillingEmail   types.String `tfsdk:"billing_email"`
}

func (data *Organization) ReadFromResponse(
	ctx context.Context,
	organization *platform.Organization,
) diag.Diagnostics {
	data.Id = types.StringValue(organization.Id)
	data.Name = types.StringValue(organization.Name)
	data.SupportPlan = types.StringValue(string(organization.SupportPlan))
	data.Product = types.StringPointerValue((*string)(organization.Product))
	data.CreatedAt = types.StringValue(organization.CreatedAt.String())
	data.UpdatedAt = types.StringValue(organization.UpdatedAt.String())
	var diags diag.Diagnostics
	data.CreatedBy, diags = SubjectProfileTypesObject(ctx, organization.CreatedBy)
	if diags.HasError() {
		return diags
	}
	data.UpdatedBy, diags = SubjectProfileTypesObject(ctx, organization.CreatedBy)
	if diags.HasError() {
		return diags
	}
	if organization.TrialExpiresAt != nil {
		data.TrialExpiresAt = types.StringValue(organization.TrialExpiresAt.String())
	}
	data.Status = types.StringPointerValue((*string)(organization.Status))
	data.PaymentMethod = types.StringPointerValue((*string)(organization.PaymentMethod))
	data.IsScimEnabled = types.BoolValue(organization.IsScimEnabled)
	data.BillingEmail = types.StringPointerValue(organization.BillingEmail)

	return nil
}
