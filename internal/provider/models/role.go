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

type DagRole struct {
	DagId        types.String `tfsdk:"dag_id"`
	DeploymentId types.String `tfsdk:"deployment_id"`
	Role         types.String `tfsdk:"role"`
	Tag          types.String `tfsdk:"tag"`
}

func DagRoleTypesObject(
	ctx context.Context,
	role iam.DagRole,
) (types.Object, diag.Diagnostics) {
	obj := DagRole{
		DeploymentId: types.StringValue(role.DeploymentId),
		Role:         types.StringValue(role.Role),
	}
	if role.DagId != nil {
		obj.DagId = types.StringValue(*role.DagId)
	} else {
		obj.DagId = types.StringNull()
	}
	if role.Tag != nil {
		obj.Tag = types.StringValue(*role.Tag)
	} else {
		obj.Tag = types.StringNull()
	}
	return types.ObjectValueFrom(ctx, schemas.DagRoleAttributeTypes(), obj)
}

type ApiTokenRole struct {
	EntityId     types.String `tfsdk:"entity_id"`
	EntityType   types.String `tfsdk:"entity_type"`
	Role         types.String `tfsdk:"role"`
	DeploymentId types.String `tfsdk:"deployment_id"`
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
	if role.DeploymentId != nil {
		obj.DeploymentId = types.StringValue(*role.DeploymentId)
	} else {
		obj.DeploymentId = types.StringNull()
	}
	return types.ObjectValueFrom(ctx, schemas.ApiTokenRoleAttributeTypes(), obj)
}
