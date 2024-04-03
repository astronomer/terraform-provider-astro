package models

import (
	"context"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// WorkspaceDataSourceModel describes the data source data model.
type WorkspaceDataSourceModel struct {
	Id                  types.String         `tfsdk:"id"`
	Name                types.String         `tfsdk:"name"`
	Description         types.String         `tfsdk:"description"`
	OrganizationName    types.String         `tfsdk:"organization_name"`
	CicdEnforcedDefault types.Bool           `tfsdk:"cicd_enforced_default"`
	CreatedAt           types.String         `tfsdk:"created_at"`
	UpdatedAt           types.String         `tfsdk:"updated_at"`
	CreatedBy           *SubjectProfileModel `tfsdk:"created_by"`
	UpdatedBy           *SubjectProfileModel `tfsdk:"updated_by"`
}

// WorkspaceResourceModel describes the resource data model.
type WorkspaceResourceModel struct {
	Id                  types.String         `tfsdk:"id"`
	Name                types.String         `tfsdk:"name"`
	Description         types.String         `tfsdk:"description"`
	OrganizationName    types.String         `tfsdk:"organization_name"`
	CicdEnforcedDefault types.Bool           `tfsdk:"cicd_enforced_default"`
	CreatedAt           types.String         `tfsdk:"created_at"`
	UpdatedAt           types.String         `tfsdk:"updated_at"`
	CreatedBy           *SubjectProfileModel `tfsdk:"created_by"`
	UpdatedBy           *SubjectProfileModel `tfsdk:"updated_by"`
}

func FillWorkspaceResourceState(
	ctx context.Context,
	workspace *platform.Workspace,
	data *WorkspaceResourceModel,
) diag.Diagnostics {
	data.Id = types.StringValue(workspace.Id)
	data.Name = types.StringValue(workspace.Name)
	if workspace.Description != nil {
		data.Description = types.StringValue(*workspace.Description)
	}
	if workspace.OrganizationName != nil {
		data.OrganizationName = types.StringValue(*workspace.OrganizationName)
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

func FillWorkspaceDataSourceState(
	ctx context.Context,
	workspace *platform.Workspace,
	data *WorkspaceDataSourceModel,
) diag.Diagnostics {
	data.Id = types.StringValue(workspace.Id)
	data.Name = types.StringValue(workspace.Name)
	if workspace.Description != nil {
		data.Description = types.StringValue(*workspace.Description)
	}
	if workspace.OrganizationName != nil {
		data.OrganizationName = types.StringValue(*workspace.OrganizationName)
	}
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
