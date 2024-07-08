package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// User describes the data source data model.
type User struct {
	Id               types.String `tfsdk:"id"`
	Username         types.String `tfsdk:"username"`
	FullName         types.String `tfsdk:"full_name"`
	Status           types.String `tfsdk:"status"`
	AvatarUrl        types.String `tfsdk:"avatar_url"`
	OrganizationRole types.String `tfsdk:"organization_role"`
	DeploymentRoles  types.Set    `tfsdk:"deployment_roles"`
	WorkspaceRoles   types.Set    `tfsdk:"workspace_roles"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func (data *User) ReadFromResponse(ctx context.Context, user *iam.User) diag.Diagnostics {
	var diags diag.Diagnostics
	data.Id = types.StringValue(user.Id)
	data.Username = types.StringValue(user.Username)
	data.FullName = types.StringValue(user.FullName)
	data.Status = types.StringValue(string(user.Status))
	data.AvatarUrl = types.StringValue(user.AvatarUrl)
	if user.OrganizationRole != nil {
		data.OrganizationRole = types.StringValue(string(*user.OrganizationRole))
	} else {
		data.OrganizationRole = types.StringValue("")
	}
	data.DeploymentRoles, diags = utils.ObjectSet(ctx, user.DeploymentRoles, schemas.DeploymentRoleAttributeTypes(), DeploymentRoleTypesObject)
	if diags.HasError() {
		return diags
	}
	data.WorkspaceRoles, diags = utils.ObjectSet(ctx, user.WorkspaceRoles, schemas.WorkspaceRoleAttributeTypes(), WorkspaceRoleTypesObject)
	if diags.HasError() {
		return diags
	}
	data.CreatedAt = types.StringValue(user.CreatedAt.String())
	data.UpdatedAt = types.StringValue(user.UpdatedAt.String())

	return nil
}
