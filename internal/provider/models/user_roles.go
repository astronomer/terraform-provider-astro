package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UserRoles describes the user_roles resource
type UserRoles struct {
	UserId           types.String `tfsdk:"user_id"`
	OrganizationRole types.String `tfsdk:"organization_role"`
	WorkspaceRoles   types.Set    `tfsdk:"workspace_roles"`
	DeploymentRoles  types.Set    `tfsdk:"deployment_roles"`
}

func (data *UserRoles) ReadFromResponse(
	ctx context.Context,
	userId string,
	userRoles *iam.SubjectRoles,
) diag.Diagnostics {
	var diags diag.Diagnostics
	data.UserId = types.StringValue(userId)
	data.OrganizationRole = types.StringValue(string(*userRoles.OrganizationRole))
	data.WorkspaceRoles, diags = utils.ObjectSet(ctx, userRoles.WorkspaceRoles, schemas.WorkspaceRoleAttributeTypes(), WorkspaceRoleTypesObject)
	if diags.HasError() {
		return diags
	}
	data.DeploymentRoles, diags = utils.ObjectSet(ctx, userRoles.DeploymentRoles, schemas.DeploymentRoleAttributeTypes(), DeploymentRoleTypesObject)
	if diags.HasError() {
		return diags
	}
	return nil
}
