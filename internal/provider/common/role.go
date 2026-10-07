package common

import (
	"context"
	"fmt"
	"strings"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	platform_v1 "github.com/astronomer/terraform-provider-astro/internal/clients/platform_v1"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

// RequestWorkspaceRoles converts a Terraform set to a list of platform_v1.WorkspaceRole to be used in create and update requests
func RequestWorkspaceRoles(ctx context.Context, workspaceRolesObjSet types.Set) ([]platform_v1.WorkspaceRole, diag.Diagnostics) {
	if len(workspaceRolesObjSet.Elements()) == 0 {
		return []platform_v1.WorkspaceRole{}, nil
	}

	var roles []models.WorkspaceRole
	diags := workspaceRolesObjSet.ElementsAs(ctx, &roles, false)
	if diags.HasError() {
		return nil, diags
	}
	workspaceRoles := lo.Map(roles, func(role models.WorkspaceRole, _ int) platform_v1.WorkspaceRole {
		return platform_v1.WorkspaceRole{
			Role:        platform_v1.WorkspaceRoleRole(role.Role.ValueString()),
			WorkspaceId: role.WorkspaceId.ValueString(),
		}
	})
	return workspaceRoles, nil
}

// RequestDeploymentRoles converts a Terraform set to a list of platform_v1.DeploymentRole to be used in create and update requests
func RequestDeploymentRoles(ctx context.Context, deploymentRolesObjSet types.Set) ([]platform_v1.DeploymentRole, diag.Diagnostics) {
	if len(deploymentRolesObjSet.Elements()) == 0 {
		return []platform_v1.DeploymentRole{}, nil
	}

	var roles []models.DeploymentRole
	diags := deploymentRolesObjSet.ElementsAs(ctx, &roles, false)
	if diags.HasError() {
		return nil, diags
	}
	deploymentRoles := lo.Map(roles, func(role models.DeploymentRole, _ int) platform_v1.DeploymentRole {
		return platform_v1.DeploymentRole{
			Role:         role.Role.ValueString(),
			DeploymentId: role.DeploymentId.ValueString(),
		}
	})
	return deploymentRoles, nil
}

// RequestDagRoles converts a Terraform set to a list of platform_v1.DagRole to be used in create and update requests
func RequestDagRoles(ctx context.Context, dagRolesObjSet types.Set) ([]platform_v1.DagRole, diag.Diagnostics) {
	if len(dagRolesObjSet.Elements()) == 0 {
		return []platform_v1.DagRole{}, nil
	}

	var roles []models.DagRole
	diags := dagRolesObjSet.ElementsAs(ctx, &roles, false)
	if diags.HasError() {
		return nil, diags
	}
	dagRoles := lo.Map(roles, func(role models.DagRole, _ int) platform_v1.DagRole {
		dagRole := platform_v1.DagRole{
			DeploymentId: role.DeploymentId.ValueString(),
			Role:         role.Role.ValueString(),
		}
		if !role.DagId.IsNull() && role.DagId.ValueString() != "" {
			dagRole.DagId = lo.ToPtr(role.DagId.ValueString())
		}
		if !role.Tag.IsNull() && role.Tag.ValueString() != "" {
			dagRole.DagTag = lo.ToPtr(role.Tag.ValueString())
		}
		return dagRole
	})
	return dagRoles, nil
}

// ValidateRoleMatchesEntityType checks if the role is valid for the entityType
func ValidateRoleMatchesEntityType(role string, scopeType string) bool {
	if role == "" || scopeType == "" {
		return false
	}

	organizationRoles := []string{
		string(platform_v1.UserOrganizationRoleORGANIZATIONBILLINGADMIN),
		string(platform_v1.UserOrganizationRoleORGANIZATIONMEMBER),
		string(platform_v1.UserOrganizationRoleORGANIZATIONOWNER),
		string(platform_v1.UserOrganizationRoleORGANIZATIONOBSERVEADMIN),
		string(platform_v1.UserOrganizationRoleORGANIZATIONOBSERVEMEMBER),
	}
	workspaceRoles := []string{string(platform_v1.WORKSPACEACCESSOR), string(platform_v1.WORKSPACEAUTHOR), string(platform_v1.WORKSPACEMEMBER), string(platform_v1.WORKSPACEOWNER), string(platform_v1.WORKSPACEOPERATOR)}
	deploymentRoles := []string{"DEPLOYMENT_ADMIN"}
	var nonEntityRoles []string

	scopeType = strings.ToLower(scopeType)
	switch scopeType {
	case "organization":
		nonEntityRoles = append(workspaceRoles, deploymentRoles...)
	case "workspace":
		nonEntityRoles = append(organizationRoles, deploymentRoles...)
	case "deployment":
		nonEntityRoles = append(organizationRoles, workspaceRoles...)
	}

	return !lo.Contains(nonEntityRoles, role)
}

type ValidateWorkspaceDeploymentRolesInput struct {
	PlatformClient  platform_v1.ClientWithResponsesInterface
	OrganizationId  string
	DeploymentRoles []platform_v1.DeploymentRole
	WorkspaceRoles  []platform_v1.WorkspaceRole
	Limit           int // page size for ListDeployments; defaults to 1000 if zero
}

// ValidateWorkspaceDeploymentRoles checks if deployment roles have corresponding workspace roles
func ValidateWorkspaceDeploymentRoles(ctx context.Context, input ValidateWorkspaceDeploymentRolesInput) diag.Diagnostics {
	// return nil if there are no deployment roles
	if len(input.DeploymentRoles) == 0 {
		return nil
	}

	// get list of deploymentRole ids
	deploymentRoleIds := lo.Map(input.DeploymentRoles, func(role platform_v1.DeploymentRole, _ int) string {
		return role.DeploymentId
	})
	deploymentRoleIds = lo.Uniq(deploymentRoleIds)

	limit := input.Limit
	if limit == 0 {
		limit = 1000
	}

	// get list of deployments (paginated)
	params := &platform_v1.ListDeploymentsParams{
		DeploymentIds: &deploymentRoleIds,
		Limit:         lo.ToPtr(limit),
	}

	var queriedDeployments []platform_v1.Deployment
	offset := 0 // Will be incremented appropriately to ensure all Deployments are returned

	for {
		params.Offset = &offset
		listDeployments, err := input.PlatformClient.ListDeploymentsWithResponse(ctx, input.OrganizationId, params)

		if err != nil {
			tflog.Error(ctx, "failed to mutate roles", map[string]interface{}{"error": err})
			return diag.Diagnostics{diag.NewErrorDiagnostic(
				"Client Error",
				fmt.Sprintf("Unable to mutate roles and list deployments, got error: %s", err),
			)}
		}

		_, diagnostic := clients.NormalizeAPIResponseWithBody(ctx, listDeployments.HTTPResponse, listDeployments.Body, listDeployments.JSON200, "list deployments")
		if diagnostic != nil {
			return diag.Diagnostics{diagnostic}
		}

		// Handle an empty page, breaking early
		if len(listDeployments.JSON200.Deployments) == 0 {
			break
		}

		// Append new Deployments to queriedDeployments, which we'll eventually use to validate Deployment ID's
		queriedDeployments = append(queriedDeployments, listDeployments.JSON200.Deployments...)

		// Break if we've hit the last page; otherwise, increment the offset and continue
		if listDeployments.JSON200.TotalCount <= offset+len(listDeployments.JSON200.Deployments) {
			break
		}
		offset += len(listDeployments.JSON200.Deployments)
	}

	// get list of deployment ids
	deploymentIds := lo.Map(queriedDeployments, func(deployment platform_v1.Deployment, _ int) string {
		return deployment.Id
	})

	// check if deploymentRole ids are in list of deployments
	invalidDeploymentIds, _ := lo.Difference(deploymentRoleIds, deploymentIds)
	if len(invalidDeploymentIds) > 0 {
		tflog.Error(ctx, "failed to mutate roles")
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Unable to mutate roles, not every deployment role has a corresponding valid deployment",
			fmt.Sprintf("Please ensure that every deployment role has a corresponding deployment, got invalid deployment ids: %v", invalidDeploymentIds),
		),
		}
	}

	// get list of workspace ids from deployments
	deploymentWorkspaceIds := lo.Map(queriedDeployments, func(deployment platform_v1.Deployment, _ int) string {
		return deployment.WorkspaceId
	})
	deploymentWorkspaceIds = lo.Uniq(deploymentWorkspaceIds)

	// get list of workspaceRole ids
	workspaceRoleIds := lo.Map(input.WorkspaceRoles, func(role platform_v1.WorkspaceRole, _ int) string {
		return role.WorkspaceId
	})

	// check if deploymentWorkspaceIds are in workspaceRoleIds
	workspaceRoleIds = lo.Intersect(lo.Uniq(workspaceRoleIds), deploymentWorkspaceIds)
	if len(workspaceRoleIds) != len(deploymentWorkspaceIds) {
		tflog.Error(ctx, "failed to mutate roles")
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Unable to mutate roles, not every deployment role has a corresponding workspace role",
			"Please ensure that every deployment role has a corresponding workspace role",
		),
		}
	}
	return nil
}

// GetDuplicateWorkspaceIds checks if there are duplicate workspace ids in the workspace roles
func GetDuplicateWorkspaceIds(workspaceRoles []platform_v1.WorkspaceRole) []string {
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

// GetDuplicateDeploymentIds checks if there are duplicate deployment ids in the deployment roles
func GetDuplicateDeploymentIds(deploymentRoles []platform_v1.DeploymentRole) []string {
	deploymentIdCount := make(map[string]int)
	for _, role := range deploymentRoles {
		deploymentIdCount[role.DeploymentId]++
	}

	var duplicates []string
	for id, count := range deploymentIdCount {
		if count > 1 {
			duplicates = append(duplicates, id)
		}
	}

	return duplicates
}

// GetDuplicateDagRoleKeys checks if there are duplicate dag_id+deployment_id or tag+deployment_id combinations in the dag roles
func GetDuplicateDagRoleKeys(dagRoles []platform_v1.DagRole) []string {
	keyCount := make(map[string]int)
	for _, role := range dagRoles {
		var key string
		if role.DagId != nil {
			key = fmt.Sprintf("dag_id:%s:deployment_id:%s", *role.DagId, role.DeploymentId)
		} else if role.DagTag != nil {
			key = fmt.Sprintf("tag:%s:deployment_id:%s", *role.DagTag, role.DeploymentId)
		}
		if key != "" {
			keyCount[key]++
		}
	}

	var duplicates []string
	for key, count := range keyCount {
		if count > 1 {
			duplicates = append(duplicates, key)
		}
	}

	return duplicates
}

// ValidateDagRoles validates that each dag role has either dag_id or tag (but not both) and a deployment_id
func ValidateDagRoles(dagRoles []platform_v1.DagRole) diag.Diagnostics {
	for _, role := range dagRoles {
		hasDagId := role.DagId != nil && *role.DagId != ""
		hasTag := role.DagTag != nil && *role.DagTag != ""

		if !hasDagId && !hasTag {
			return diag.Diagnostics{diag.NewErrorDiagnostic(
				"Invalid DAG role configuration",
				"Each DAG role must have either 'dag_id' or 'tag' specified",
			)}
		}

		if hasDagId && hasTag {
			return diag.Diagnostics{diag.NewErrorDiagnostic(
				"Invalid DAG role configuration",
				"Each DAG role must have either 'dag_id' or 'tag' specified, but not both",
			)}
		}

		if role.DeploymentId == "" {
			return diag.Diagnostics{diag.NewErrorDiagnostic(
				"Invalid DAG role configuration",
				"Each DAG role must have a 'deployment_id' specified",
			)}
		}
	}

	duplicateKeys := GetDuplicateDagRoleKeys(dagRoles)
	if len(duplicateKeys) > 0 {
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid Configuration: Cannot have multiple DAG roles with the same dag_id/tag and deployment_id combination",
			fmt.Sprintf("Please provide unique dag_id/tag and deployment_id combinations. The following are duplicated: %v", duplicateKeys),
		)}
	}

	return nil
}

func ValidateRoles(
	workspaceRoles []platform_v1.WorkspaceRole,
	deploymentRoles []platform_v1.DeploymentRole,
) diag.Diagnostics {
	return ValidateRolesWithDagRoles(workspaceRoles, deploymentRoles, nil)
}

func ValidateRolesWithDagRoles(
	workspaceRoles []platform_v1.WorkspaceRole,
	deploymentRoles []platform_v1.DeploymentRole,
	dagRoles []platform_v1.DagRole,
) diag.Diagnostics {
	for _, role := range workspaceRoles {
		if !ValidateRoleMatchesEntityType(string(role.Role), string(platform_v1.RoleScopeTypeWORKSPACE)) {
			return diag.Diagnostics{diag.NewErrorDiagnostic(
				fmt.Sprintf("Role '%s' is not valid for role type '%s'", string(role.Role), string(platform_v1.RoleScopeTypeWORKSPACE)),
				fmt.Sprintf("Please provide a valid role for the type '%s'", string(platform_v1.RoleScopeTypeWORKSPACE)),
			)}
		}
	}

	duplicateWorkspaceIds := GetDuplicateWorkspaceIds(workspaceRoles)
	if len(duplicateWorkspaceIds) > 0 {
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid Configuration: Cannot have multiple roles with the same workspace id",
			fmt.Sprintf("Please provide a unique workspace id for each role. The following workspace ids are duplicated: %v", duplicateWorkspaceIds),
		)}
	}

	for _, role := range deploymentRoles {
		if !ValidateRoleMatchesEntityType(role.Role, string(platform_v1.RoleScopeTypeDEPLOYMENT)) {
			return diag.Diagnostics{diag.NewErrorDiagnostic(
				fmt.Sprintf("Role '%s' is not valid for role type '%s'", role.Role, string(platform_v1.RoleScopeTypeDEPLOYMENT)),
				fmt.Sprintf("Please provide a valid role for the type '%s'", string(platform_v1.RoleScopeTypeDEPLOYMENT)),
			)}
		}
	}

	duplicateDeploymentIds := GetDuplicateDeploymentIds(deploymentRoles)
	if len(duplicateDeploymentIds) > 0 {
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid Configuration: Cannot have multiple roles with the same deployment id",
			fmt.Sprintf("Please provide unique deployment id for each role. The following deployment ids are duplicated: %v", duplicateDeploymentIds),
		)}
	}

	// Validate dag roles if provided
	if len(dagRoles) > 0 {
		if diags := ValidateDagRoles(dagRoles); diags.HasError() {
			return diags
		}

		dagDeploymentIds := lo.Uniq(lo.Map(dagRoles, func(r platform_v1.DagRole, _ int) string {
			return r.DeploymentId
		}))
		deploymentRoleIds := lo.Map(deploymentRoles, func(r platform_v1.DeploymentRole, _ int) string {
			return r.DeploymentId
		})
		missingIds, _ := lo.Difference(dagDeploymentIds, deploymentRoleIds)
		if len(missingIds) > 0 {
			return diag.Diagnostics{diag.NewErrorDiagnostic(
				"Invalid Configuration: dag_roles requires corresponding deployment_roles",
				fmt.Sprintf("Each deployment referenced in dag_roles must also have an entry in deployment_roles. Missing deployment_roles for deployment IDs: %v", missingIds),
			)}
		}
	}

	return nil
}
