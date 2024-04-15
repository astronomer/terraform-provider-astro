package models

import (
	"context"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// WorkspaceDataSource describes the data source data model.
type WorkspaceDataSource struct {
	Id                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	CicdEnforcedDefault types.Bool   `tfsdk:"cicd_enforced_default"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
	CreatedBy           types.Object `tfsdk:"created_by"`
	UpdatedBy           types.Object `tfsdk:"updated_by"`
}

// WorkspaceResource describes the resource data model.
type WorkspaceResource struct {
	Id                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	CicdEnforcedDefault types.Bool   `tfsdk:"cicd_enforced_default"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
	CreatedBy           types.Object `tfsdk:"created_by"`
	UpdatedBy           types.Object `tfsdk:"updated_by"`
}

func (data *WorkspaceResource) ReadFromResponse(
	ctx context.Context,
	workspace *platform.Workspace,
) diag.Diagnostics {
	data.Id = types.StringValue(workspace.Id)
	data.Name = types.StringValue(workspace.Name)
	// If the description is nil, set it to an empty string since the terraform state/config for this resource
	// cannot have a null value for a string.
	if workspace.Description != nil {
		data.Description = types.StringValue(*workspace.Description)
	} else {
		data.Description = types.StringValue("")
	}
	data.CicdEnforcedDefault = types.BoolValue(workspace.CicdEnforcedDefault)
	data.CreatedAt = types.StringValue(workspace.CreatedAt.String())
	data.UpdatedAt = types.StringValue(workspace.UpdatedAt.String())
	var diags diag.Diagnostics
	data.CreatedBy, diags = SubjectProfileTypesObject(ctx, workspace.CreatedBy)
	if diags.HasError() {
		return diags
	}
	data.UpdatedBy, diags = SubjectProfileTypesObject(ctx, workspace.CreatedBy)
	if diags.HasError() {
		return diags
	}

	return nil
}

func (data *WorkspaceDataSource) ReadFromResponse(
	ctx context.Context,
	workspace *platform.Workspace,
) diag.Diagnostics {
	data.Id = types.StringValue(workspace.Id)
	data.Name = types.StringValue(workspace.Name)
	data.Description = types.StringPointerValue(workspace.Description)
	data.CicdEnforcedDefault = types.BoolValue(workspace.CicdEnforcedDefault)
	data.CreatedAt = types.StringValue(workspace.CreatedAt.String())
	data.UpdatedAt = types.StringValue(workspace.UpdatedAt.String())
	var diags diag.Diagnostics
	data.CreatedBy, diags = SubjectProfileTypesObject(ctx, workspace.CreatedBy)
	if diags.HasError() {
		return diags
	}
	data.UpdatedBy, diags = SubjectProfileTypesObject(ctx, workspace.UpdatedBy)
	if diags.HasError() {
		return diags
	}

	return nil
}
