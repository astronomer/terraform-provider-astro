package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Workspaces describes the data source data model.
type Workspaces struct {
	Workspaces   types.List `tfsdk:"workspaces"`
	WorkspaceIds types.List `tfsdk:"workspace_ids"` // query parameter
	Names        types.List `tfsdk:"names"`         // query parameter
}

func (data *Workspaces) ReadFromResponse(
	ctx context.Context,
	workspaces []platform.Workspace,
) diag.Diagnostics {
	values := make([]attr.Value, len(workspaces))
	for i, workspace := range workspaces {
		var singleWorkspaceData Workspace
		diags := singleWorkspaceData.ReadFromResponse(ctx, &workspace)
		if diags.HasError() {
			return diags
		}

		objectValue, diags := types.ObjectValueFrom(ctx, schemas.WorkspacesElementAttributeTypes(), singleWorkspaceData)
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
