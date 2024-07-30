package common

import (
	"context"
	"fmt"
	"strings"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/hashicorp/terraform-plugin-log/tflog"

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

// ValidateRoleMatchesEntityType checks if the role is valid for the entityType
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

type ValidateWorkspaceDeploymentRolesInput struct {
	PlatformClient  *platform.ClientWithResponses
	OrganizationId  string
	DeploymentRoles []iam.DeploymentRole
	WorkspaceRoles  []iam.WorkspaceRole
}

// ValidateWorkspaceDeploymentRoles checks if deployment roles have corresponding workspace roles
func ValidateWorkspaceDeploymentRoles(ctx context.Context, diags diag.Diagnostics, input ValidateWorkspaceDeploymentRolesInput) diag.Diagnostics {
	// return nil if there are no deployment roles
	if len(input.DeploymentRoles) == 0 {
		return nil
	}

	// get list of deployment ids
	deploymentIds := lo.Map(input.DeploymentRoles, func(role iam.DeploymentRole, _ int) string {
		return role.DeploymentId
	})

	// get list of deployments
	listDeployments, err := input.PlatformClient.ListDeploymentsWithResponse(ctx, input.OrganizationId, &platform.ListDeploymentsParams{
		DeploymentIds: &deploymentIds,
	})
	if err != nil {
		tflog.Error(ctx, "failed to mutate Team roles", map[string]interface{}{"error": err})
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to mutate Team roles and list deployments, got error: %s", err),
		)
		return diags
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, listDeployments.HTTPResponse, listDeployments.Body)
	if diagnostic != nil {
		diags.Append(diagnostic)
		return diags
	}

	// get list of workspace ids from deployments
	deploymentWorkspaceIds := lo.Map(listDeployments.JSON200.Deployments, func(deployment platform.Deployment, _ int) string {
		return deployment.WorkspaceId
	})

	// get list of workspaceIds
	workspaceIds := lo.Map(input.WorkspaceRoles, func(role iam.WorkspaceRole, _ int) string {
		return role.WorkspaceId
	})

	// check if deploymentWorkspaceIds are in workspaceIds
	workspaceIds = lo.Intersect(lo.Uniq(workspaceIds), lo.Uniq(deploymentWorkspaceIds))
	if len(workspaceIds) != len(deploymentWorkspaceIds) {
		tflog.Error(ctx, "failed to mutate Team roles", map[string]interface{}{"error": err})
		diags.AddError(
			"Unable to mutate Team roles, not every deployment role has a corresponding workspace role",
			"Please ensure that every deployment role has a corresponding workspace role",
		)
		return diags
	}
	return diags
}

// CheckDuplicateWorkspaceId checks if there are duplicate workspace ids in the workspace roles
func CheckDuplicateWorkspaceId(workspaceRoles []iam.WorkspaceRole) []string {
	workspaceIdCount := make(map[string]int)
	for _, role := range workspaceRoles {
		workspaceIdCount[role.WorkspaceId]++
	}

	var duplicates []string
	for id, count := range workspaceIdCount {
		if count > 1 {
			duplicates = append(duplicates, id)
		}
	}

	return duplicates
}

// CheckDuplicateDeploymentId checks if there are duplicate deployment ids in the deployment roles
func CheckDuplicateDeploymentId(deploymentRoles []iam.DeploymentRole) []string {
	workspaceIdCount := make(map[string]int)
	for _, role := range deploymentRoles {
		workspaceIdCount[role.DeploymentId]++
	}

	var duplicates []string
	for id, count := range workspaceIdCount {
		if count > 1 {
			duplicates = append(duplicates, id)
		}
	}

	return duplicates
}
