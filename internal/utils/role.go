package utils

import (
	"strings"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/samber/lo"
)

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
		return true
	}

	return lo.Contains(roles, role)
}
