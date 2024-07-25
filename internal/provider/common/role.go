package common

import (
	"context"
	"strings"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

// RequestWorkspaceRoles converts a Terraform set to a list of iam.WorkspaceRole to be used in create and update requests
func RequestWorkspaceRoles(ctx context.Context, workspaceRolesObjSet types.Set) ([]iam.WorkspaceRole, diag.Diagnostics) {
	if len(workspaceRolesObjSet.Elements()) == 0 {
		return []iam.WorkspaceRole{}, nil
	}

	var roles []models.WorkspaceRole
	diags := workspaceRolesObjSet.ElementsAs(ctx, &roles, false)
	if diags.HasError() {
		return nil, diags
	}
	workspaceRoles := lo.Map(roles, func(role models.WorkspaceRole, _ int) iam.WorkspaceRole {
		return iam.WorkspaceRole{
			Role:        iam.WorkspaceRoleRole(role.Role.ValueString()),
			WorkspaceId: role.WorkspaceId.ValueString(),
		}
	})
	return workspaceRoles, nil
}

// RequestDeploymentRoles converts a Terraform set to a list of iam.DeploymentRole to be used in create and update requests
func RequestDeploymentRoles(ctx context.Context, deploymentRolesObjSet types.Set) ([]iam.DeploymentRole, diag.Diagnostics) {
	if len(deploymentRolesObjSet.Elements()) == 0 {
		return []iam.DeploymentRole{}, nil
	}

	var roles []models.DeploymentRole
	diags := deploymentRolesObjSet.ElementsAs(ctx, &roles, false)
	if diags.HasError() {
		return nil, diags
	}
	deploymentRoles := lo.Map(roles, func(role models.DeploymentRole, _ int) iam.DeploymentRole {
		return iam.DeploymentRole{
			Role:         role.Role.ValueString(),
			DeploymentId: role.DeploymentId.ValueString(),
		}
	})
	return deploymentRoles, nil
}

func ValidateRoleMatchesEntityType(role string, scopeType string) bool {
	organizationRoles := []string{string(iam.ORGANIZATIONBILLINGADMIN), string(iam.ORGANIZATIONMEMBER), string(iam.ORGANIZATIONOWNER)}
	workspaceRoles := []string{string(iam.WORKSPACEACCESSOR), string(iam.WORKSPACEAUTHOR), string(iam.WORKSPACEMEMBER), string(iam.WORKSPACEOWNER), string(iam.WORKSPACEOPERATOR)}
	var roles []string

	scopeType = strings.ToLower(scopeType)
	if scopeType == "organization" {
		roles = organizationRoles
	} else if scopeType == "workspace" {
		roles = workspaceRoles
	} else if scopeType == "deployment" {
		nonDeploymentRoles := append(organizationRoles, workspaceRoles...)
		return !lo.Contains(nonDeploymentRoles, role)
	}

	return lo.Contains(roles, role)
}
