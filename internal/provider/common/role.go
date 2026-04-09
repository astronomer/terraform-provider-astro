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

// RequestDagRoles converts a Terraform set to a list of iam.DagRole to be used in create and update requests
func RequestDagRoles(ctx context.Context, dagRolesObjSet types.Set) ([]iam.DagRole, diag.Diagnostics) {
	if len(dagRolesObjSet.Elements()) == 0 {
		return []iam.DagRole{}, nil
	}

	var roles []models.DagRole
	diags := dagRolesObjSet.ElementsAs(ctx, &roles, false)
	if diags.HasError() {
		return nil, diags
	}
	dagRoles := lo.Map(roles, func(role models.DagRole, _ int) iam.DagRole {
		dagRole := iam.DagRole{
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
		string(iam.UserOrganizationRoleORGANIZATIONBILLINGADMIN),
		string(iam.UserOrganizationRoleORGANIZATIONMEMBER),
		string(iam.UserOrganizationRoleORGANIZATIONOWNER),
		string(iam.UserOrganizationRoleORGANIZATIONOBSERVEADMIN),
		string(iam.UserOrganizationRoleORGANIZATIONOBSERVEMEMBER),
	}
	workspaceRoles := []string{string(iam.WORKSPACEACCESSOR), string(iam.WORKSPACEAUTHOR), string(iam.WORKSPACEMEMBER), string(iam.WORKSPACEOWNER), string(iam.WORKSPACEOPERATOR)}
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
	PlatformClient  *platform.ClientWithResponses
	OrganizationId  string
	DeploymentRoles []iam.DeploymentRole
	WorkspaceRoles  []iam.WorkspaceRole
}

// ValidateWorkspaceDeploymentRoles checks if deployment roles have corresponding workspace roles
func ValidateWorkspaceDeploymentRoles(ctx context.Context, input ValidateWorkspaceDeploymentRolesInput) diag.Diagnostics {
	// return nil if there are no deployment roles
	if len(input.DeploymentRoles) == 0 {
		return nil
	}

	// get list of deploymentRole ids
	deploymentRoleIds := lo.Map(input.DeploymentRoles, func(role iam.DeploymentRole, _ int) string {
		return role.DeploymentId
	})
	deploymentRoleIds = lo.Uniq(deploymentRoleIds)

	// get list of deployments
	listDeployments, err := input.PlatformClient.ListDeploymentsWithResponse(ctx, input.OrganizationId, &platform.ListDeploymentsParams{
		DeploymentIds: &deploymentRoleIds,
	})
	if err != nil {
		tflog.Error(ctx, "failed to mutate roles", map[string]interface{}{"error": err})
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Client Error",
			fmt.Sprintf("Unable to mutate roles and list deployments, got error: %s", err),
		),
		}
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, listDeployments.HTTPResponse, listDeployments.Body)
	if diagnostic != nil {
		return diag.Diagnostics{diagnostic}
	}

	// get list of deployment ids
	deploymentIds := lo.Map(listDeployments.JSON200.Deployments, func(deployment platform.Deployment, _ int) string {
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
	deploymentWorkspaceIds := lo.Map(listDeployments.JSON200.Deployments, func(deployment platform.Deployment, _ int) string {
		return deployment.WorkspaceId
	})
	deploymentWorkspaceIds = lo.Uniq(deploymentWorkspaceIds)

	// get list of workspaceRole ids
	workspaceRoleIds := lo.Map(input.WorkspaceRoles, func(role iam.WorkspaceRole, _ int) string {
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
func GetDuplicateWorkspaceIds(workspaceRoles []iam.WorkspaceRole) []string {
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
func GetDuplicateDeploymentIds(deploymentRoles []iam.DeploymentRole) []string {
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
func GetDuplicateDagRoleKeys(dagRoles []iam.DagRole) []string {
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
func ValidateDagRoles(dagRoles []iam.DagRole) diag.Diagnostics {
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
	workspaceRoles []iam.WorkspaceRole,
	deploymentRoles []iam.DeploymentRole,
) diag.Diagnostics {
	return ValidateRolesWithDagRoles(workspaceRoles, deploymentRoles, nil)
}

func ValidateRolesWithDagRoles(
	workspaceRoles []iam.WorkspaceRole,
	deploymentRoles []iam.DeploymentRole,
	dagRoles []iam.DagRole,
) diag.Diagnostics {
	for _, role := range workspaceRoles {
		if !ValidateRoleMatchesEntityType(string(role.Role), string(iam.RoleScopeTypeWORKSPACE)) {
			return diag.Diagnostics{diag.NewErrorDiagnostic(
				fmt.Sprintf("Role '%s' is not valid for role type '%s'", string(role.Role), string(iam.RoleScopeTypeWORKSPACE)),
				fmt.Sprintf("Please provide a valid role for the type '%s'", string(iam.RoleScopeTypeWORKSPACE)),
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
		if !ValidateRoleMatchesEntityType(role.Role, string(iam.RoleScopeTypeDEPLOYMENT)) {
			return diag.Diagnostics{diag.NewErrorDiagnostic(
				fmt.Sprintf("Role '%s' is not valid for role type '%s'", role.Role, string(iam.RoleScopeTypeDEPLOYMENT)),
				fmt.Sprintf("Please provide a valid role for the type '%s'", string(iam.RoleScopeTypeDEPLOYMENT)),
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

		dagDeploymentIds := lo.Uniq(lo.Map(dagRoles, func(r iam.DagRole, _ int) string {
			return r.DeploymentId
		}))
		deploymentRoleIds := lo.Map(deploymentRoles, func(r iam.DeploymentRole, _ int) string {
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
