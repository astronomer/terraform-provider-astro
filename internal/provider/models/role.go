package models

import (
	"context"
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WorkspaceRole struct {
	WorkspaceId types.String `tfsdk:"workspace_id"`
	Role        types.String `tfsdk:"role"`
}

func WorkspaceRoleTypesObject(
	ctx context.Context,
	role iam.WorkspaceRole,
) (types.Object, diag.Diagnostics) {
	obj := WorkspaceRole{
		WorkspaceId: types.StringValue(role.WorkspaceId),
		Role:        types.StringValue(string(role.Role)),
	}
	return types.ObjectValueFrom(ctx, schemas.WorkspaceRoleAttributeTypes(), obj)
}

type DeploymentRole struct {
	DeploymentId types.String `tfsdk:"deployment_id"`
	Role         types.String `tfsdk:"role"`
}

func DeploymentRoleTypesObject(
	ctx context.Context,
	role iam.DeploymentRole,
) (types.Object, diag.Diagnostics) {
	obj := DeploymentRole{
		DeploymentId: types.StringValue(role.DeploymentId),
		Role:         types.StringValue(string(role.Role)),
	}
	return types.ObjectValueFrom(ctx, schemas.DeploymentRoleAttributeTypes(), obj)
}
