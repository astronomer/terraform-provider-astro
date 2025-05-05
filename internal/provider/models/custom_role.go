package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CustomRoleDataSource describes the data source data model.
type CustomRoleDataSource struct {
	Id                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Permissions            types.Set    `tfsdk:"permissions"`
	RestrictedWorkspaceIds types.Set    `tfsdk:"restricted_workspace_ids"`
	ScopeType              types.String `tfsdk:"scope_type"`
	CreatedAt              types.String `tfsdk:"created_at"`
	UpdatedAt              types.String `tfsdk:"updated_at"`
	CreatedBy              types.Object `tfsdk:"created_by"`
	UpdatedBy              types.Object `tfsdk:"updated_by"`
}

// CustomRoleResource defines the resource implementation.
type CustomRoleResource struct {
	Id                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Permissions            types.Set    `tfsdk:"permissions"`
	RestrictedWorkspaceIds types.Set    `tfsdk:"restricted_workspace_ids"`
	ScopeType              types.String `tfsdk:"scope_type"`
	CreatedAt              types.String `tfsdk:"created_at"`
	UpdatedAt              types.String `tfsdk:"updated_at"`
	CreatedBy              types.Object `tfsdk:"created_by"`
	UpdatedBy              types.Object `tfsdk:"updated_by"`
}

func (data *CustomRoleDataSource) ReadFromResponse(ctx context.Context, customRole *iam.CustomRole) diag.Diagnostics {
	var diags diag.Diagnostics
	data.Id = types.StringValue(customRole.Id)
	data.Name = types.StringValue(customRole.Name)
	data.CreatedAt = types.StringValue(customRole.CreatedAt.String())
	data.UpdatedAt = types.StringValue(customRole.UpdatedAt.String())
	data.CreatedBy, diags = SubjectProfileTypesObject(ctx, customRole.CreatedBy)
	if diags.HasError() {
		return diags
	}
	data.UpdatedBy, diags = SubjectProfileTypesObject(ctx, customRole.UpdatedBy)
	if diags.HasError() {
		return diags
	}

	return diags
}

func (data *CustomRoleResource) ReadFromResponse(ctx context.Context, customRole *iam.CustomRole) diag.Diagnostics {
	var diags diag.Diagnostics
	data.Id = types.StringValue(customRole.Id)
	data.Name = types.StringValue(customRole.Name)
	data.CreatedAt = types.StringValue(customRole.CreatedAt.String())
	data.UpdatedAt = types.StringValue(customRole.UpdatedAt.String())
	data.CreatedBy, diags = SubjectProfileTypesObject(ctx, customRole.CreatedBy)
	if diags.HasError() {
		return diags
	}
	data.UpdatedBy, diags = SubjectProfileTypesObject(ctx, customRole.UpdatedBy)
	if diags.HasError() {
		return diags
	}

	return diags
}
