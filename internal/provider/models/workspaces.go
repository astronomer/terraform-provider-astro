package models

import (
	"context"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// WorkspacesDataSource describes the data source data model.
type WorkspacesDataSource struct {
	Workspaces   types.List `tfsdk:"workspaces"`
	WorkspaceIds types.List `tfsdk:"workspace_ids"`
	Names        types.List `tfsdk:"names"`
}

func (data *WorkspacesDataSource) ReadFromResponse(
	ctx context.Context,
	workspaces []platform.Workspace,
) diag.Diagnostics {
	if len(workspaces) == 0 {
		types.ListNull(types.ObjectType{AttrTypes: schemas.WorkspacesElementAttributeTypes()})
	}

	values := make([]attr.Value, len(workspaces))
	for i, workspace := range workspaces {
		v := map[string]attr.Value{}
		v["id"] = types.StringValue(workspace.Id)
		v["name"] = types.StringValue(workspace.Name)
		if workspace.Description != nil {
			v["description"] = types.StringValue(*workspace.Description)
		} else {
			v["description"] = types.StringNull()
		}
		if workspace.OrganizationName != nil {
			v["organization_name"] = types.StringValue(*workspace.OrganizationName)
		} else {
			v["organization_name"] = types.StringNull()
		}
		v["cicd_enforced_default"] = types.BoolValue(workspace.CicdEnforcedDefault)
		v["created_at"] = types.StringValue(workspace.CreatedAt.String())
		v["updated_at"] = types.StringValue(workspace.UpdatedAt.String())
		if workspace.CreatedBy != nil {
			createdBy, diags := SubjectProfileTypesObject(ctx, workspace.CreatedBy)
			if diags.HasError() {
				return diags
			}
			v["created_by"] = createdBy
		}
		if workspace.UpdatedBy != nil {
			updatedBy, diags := SubjectProfileTypesObject(ctx, workspace.UpdatedBy)
			if diags.HasError() {
				return diags
			}
			v["updated_by"] = updatedBy
		}

		objectValue, diags := types.ObjectValue(schemas.WorkspacesElementAttributeTypes(), v)
		if diags.HasError() {
			return diags
		}
		values[i] = objectValue
	}
	var diags diag.Diagnostics
	data.Workspaces, diags = types.ListValue(types.ObjectType{AttrTypes: schemas.WorkspacesElementAttributeTypes()}, values)
	if diags.HasError() {
		return diags
	}

	return nil
}
