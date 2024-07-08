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
		Role:         types.StringValue(role.Role),
	}
	return types.ObjectValueFrom(ctx, schemas.DeploymentRoleAttributeTypes(), obj)
}

type ApiTokenRole struct {
	EntityId   types.String `tfsdk:"entity_id"`
	EntityType types.String `tfsdk:"entity_type"`
	Role       types.String `tfsdk:"role"`
}

func ApiTokenRoleTypesObject(
	ctx context.Context,
	role iam.ApiTokenRole,
) (types.Object, diag.Diagnostics) {
	obj := ApiTokenRole{
		EntityId:   types.StringValue(role.EntityId),
		EntityType: types.StringValue(string(role.EntityType)),
		Role:       types.StringValue(role.Role),
	}
	return types.ObjectValueFrom(ctx, schemas.ApiTokenRoleAttributeTypes(), obj)
}
