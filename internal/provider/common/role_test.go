package common_test

import (
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/common"
)

func TestValidateRoleMatchesEntityType(t *testing.T) {
	tests := []struct {
		name      string
		role      string
		scopeType string
		want      bool
	}{
		{
			name:      "empty role",
			role:      "",
			scopeType: "organization",
			want:      false,
		},
		{
			name:      "empty scope type",
			role:      string(iam.UserOrganizationRoleORGANIZATIONOWNER),
			scopeType: "",
			want:      false,
		},
		{
			name:      "organization owner for organization scope",
			role:      string(iam.UserOrganizationRoleORGANIZATIONOWNER),
			scopeType: "organization",
			want:      true,
		},
		{
			name:      "organization member for organization scope",
			role:      string(iam.UserOrganizationRoleORGANIZATIONMEMBER),
			scopeType: "organization",
			want:      true,
		},
		{
			name:      "organization billing admin for organization scope",
			role:      string(iam.UserOrganizationRoleORGANIZATIONBILLINGADMIN),
			scopeType: "organization",
			want:      true,
		},
		{
			name:      "organization observe admin for organization scope",
			role:      string(iam.UserOrganizationRoleORGANIZATIONOBSERVEADMIN),
			scopeType: "organization",
			want:      true,
		},
		{
			name:      "organization observe member for organization scope",
			role:      string(iam.UserOrganizationRoleORGANIZATIONOBSERVEMEMBER),
			scopeType: "organization",
			want:      true,
		},
		{
			name:      "workspace role invalid for organization scope",
			role:      string(iam.WORKSPACEOWNER),
			scopeType: "organization",
			want:      false,
		},
		{
			name:      "deployment role invalid for organization scope",
			role:      "DEPLOYMENT_ADMIN",
			scopeType: "organization",
			want:      false,
		},
		{
			name:      "observe member invalid for workspace scope",
			role:      string(iam.UserOrganizationRoleORGANIZATIONOBSERVEMEMBER),
			scopeType: "workspace",
			want:      false,
		},
		{
			name:      "observe admin invalid for workspace scope",
			role:      string(iam.UserOrganizationRoleORGANIZATIONOBSERVEADMIN),
			scopeType: "workspace",
			want:      false,
		},
		{
			name:      "organization owner invalid for workspace scope",
			role:      string(iam.UserOrganizationRoleORGANIZATIONOWNER),
			scopeType: "workspace",
			want:      false,
		},
		{
			name:      "workspace owner valid for workspace scope",
			role:      string(iam.WORKSPACEOWNER),
			scopeType: "workspace",
			want:      true,
		},
		{
			name:      "observe member invalid for deployment scope",
			role:      string(iam.UserOrganizationRoleORGANIZATIONOBSERVEMEMBER),
			scopeType: "deployment",
			want:      false,
		},
		{
			name:      "deployment admin valid for deployment scope",
			role:      "DEPLOYMENT_ADMIN",
			scopeType: "deployment",
			want:      true,
		},
		{
			name:      "scope type is case insensitive",
			role:      string(iam.UserOrganizationRoleORGANIZATIONOBSERVEMEMBER),
			scopeType: "ORGANIZATION",
			want:      true,
		},
		{
			name:      "unlisted role not treated as workspace or deployment role for organization scope",
			role:      "ORGANIZATION_UNKNOWN",
			scopeType: "organization",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := common.ValidateRoleMatchesEntityType(tt.role, tt.scopeType)
			if got != tt.want {
				t.Fatalf("ValidateRoleMatchesEntityType(%q, %q) = %v, want %v",
					tt.role, tt.scopeType, got, tt.want)
			}
		})
	}
}
