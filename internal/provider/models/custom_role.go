package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CustomRole describes the data source data model.
type CustomRole struct {
	Id                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	Permissions            types.Set    `tfsdk:"permissions"`
	ScopeType              types.String `tfsdk:"scope_type"`
	RestrictedWorkspaceIds types.Set    `tfsdk:"restricted_workspace_ids"`
	CreatedAt              types.String `tfsdk:"created_at"`
	CreatedBy              types.Object `tfsdk:"created_by"`
	UpdatedAt              types.String `tfsdk:"updated_at"`
	UpdatedBy              types.Object `tfsdk:"updated_by"`
}

func (data *CustomRole) ReadFromResponse(ctx context.Context, role *iam.RoleWithPermission) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Id = types.StringValue(role.Id)
	data.Name = types.StringValue(role.Name)

	if role.Description != nil {
		data.Description = types.StringValue(*role.Description)
	} else {
		data.Description = types.StringNull()
	}

	// Convert permissions slice to Set
	permissionsSet, permDiags := types.SetValueFrom(ctx, types.StringType, role.Permissions)
	if permDiags.HasError() {
		diags.Append(permDiags...)
		return diags
	}
	data.Permissions = permissionsSet

	data.ScopeType = types.StringValue(string(role.ScopeType))

	// Convert restricted workspace IDs slice to Set
	restrictedWorkspaceIdsSet, wsDiags := types.SetValueFrom(ctx, types.StringType, role.RestrictedWorkspaceIds)
	if wsDiags.HasError() {
		diags.Append(wsDiags...)
		return diags
	}
	data.RestrictedWorkspaceIds = restrictedWorkspaceIdsSet

	data.CreatedAt = types.StringValue(role.CreatedAt)
	data.UpdatedAt = types.StringValue(role.UpdatedAt)

	// Convert CreatedBy to Object
	createdByObj, createdByDiags := SubjectProfileTypesObject(ctx, role.CreatedBy)
	if createdByDiags.HasError() {
		diags.Append(createdByDiags...)
		return diags
	}
	data.CreatedBy = createdByObj

	// Convert UpdatedBy to Object
	updatedByObj, updatedByDiags := SubjectProfileTypesObject(ctx, role.UpdatedBy)
	if updatedByDiags.HasError() {
		diags.Append(updatedByDiags...)
		return diags
	}
	data.UpdatedBy = updatedByObj

	return diags
}
