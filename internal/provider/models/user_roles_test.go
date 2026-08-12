package models_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
)

// organization_role is Required on both resources, so an omitted one must error rather than
// become a null that Terraform reports as an inconsistent result.
func TestUnit_RolesReadFromResponse_MissingOrganizationRole(t *testing.T) {
	ctx := context.Background()

	t.Run("user_roles", func(t *testing.T) {
		var data models.UserRoles
		diags := data.ReadFromResponse(ctx, "user-id", &iam.SubjectRoles{})
		assert.True(t, diags.HasError())
		assert.Contains(t, diags[0].Detail(), "organization_role")
	})

	t.Run("team_roles", func(t *testing.T) {
		var data models.TeamRoles
		diags := data.ReadFromResponse(ctx, "team-id", &iam.SubjectRoles{})
		assert.True(t, diags.HasError())
		assert.Contains(t, diags[0].Detail(), "organization_role")
	})
}

func TestUnit_RolesReadFromResponse_OrganizationRolePresent(t *testing.T) {
	ctx := context.Background()
	roles := &iam.SubjectRoles{OrganizationRole: lo.ToPtr("ORGANIZATION_MEMBER")}

	t.Run("user_roles", func(t *testing.T) {
		var data models.UserRoles
		diags := data.ReadFromResponse(ctx, "user-id", roles)
		assert.False(t, diags.HasError())
		assert.Equal(t, "ORGANIZATION_MEMBER", data.OrganizationRole.ValueString())
	})

	t.Run("team_roles", func(t *testing.T) {
		var data models.TeamRoles
		diags := data.ReadFromResponse(ctx, "team-id", roles)
		assert.False(t, diags.HasError())
		assert.Equal(t, "ORGANIZATION_MEMBER", data.OrganizationRole.ValueString())
	})
}
