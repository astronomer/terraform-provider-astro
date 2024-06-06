package datasources_test

import (
	"fmt"
	"testing"

	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAcc_DataSourceTeams(t *testing.T) {
	tfVarName := "test_data_teams"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) + teams(tfVarName),
				Check: resource.ComposeTestCheckFunc(
					checkTeams(tfVarName),
				),
			},
		},
	})
}

func teams(tfVarName string) string {
	return fmt.Sprintf(`
data astro_teams "%v" {}`, tfVarName)
}

func checkTeams(tfVarName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		instanceState, numTeams, err := utils.GetDataSourcesLength(s, tfVarName, "teams")
		if err != nil {
			return err
		}
		if numTeams == 0 {
			return fmt.Errorf("expected teams to be greater or equal to 1, got %s", instanceState.Attributes["teams.#"])
		}

		// Check the first team
		teamsIdx := 0

		id := fmt.Sprintf("teams.%d.id", teamsIdx)
		if instanceState.Attributes[id] == "" {
			return fmt.Errorf("expected 'id' to be set")
		}
		name := fmt.Sprintf("teams.%d.name", teamsIdx)
		if instanceState.Attributes[name] == "" {
			return fmt.Errorf("expected 'name' to be set")
		}
		description := fmt.Sprintf("teams.%d.description", teamsIdx)
		if instanceState.Attributes[description] == "" {
			return fmt.Errorf("expected 'description' to be set")
		}
		isIdpManaged := fmt.Sprintf("teams.%d.is_idp_managed", teamsIdx)
		if instanceState.Attributes[isIdpManaged] == "" {
			return fmt.Errorf("expected 'is_idp_managed' to be set")
		}
		organizationRole := fmt.Sprintf("teams.%d.organization_role", teamsIdx)
		if instanceState.Attributes[organizationRole] == "" {
			return fmt.Errorf("expected 'organization_role' to be set")
		}
		deploymentRoles := fmt.Sprintf("teams.%d.deployment_roles", teamsIdx)
		if instanceState.Attributes[deploymentRoles] == "" {
			return fmt.Errorf("expected 'deployment_roles' to be set")
		}
		workspaceRoles := fmt.Sprintf("teams.%d.workspace_roles", teamsIdx)
		if instanceState.Attributes[workspaceRoles] == "" {
			return fmt.Errorf("expected 'workspace_roles' to be set")
		}
		rolesCount := fmt.Sprintf("teams.%d.roles_count", teamsIdx)
		if instanceState.Attributes[rolesCount] == "" {
			return fmt.Errorf("expected 'roles_count' to be set")
		}
		createdAt := fmt.Sprintf("teams.%d.created_at", teamsIdx)
		if instanceState.Attributes[createdAt] == "" {
			return fmt.Errorf("expected 'created_at' to be set")
		}
		updatedAt := fmt.Sprintf("teams.%d.updated_at", teamsIdx)
		if instanceState.Attributes[updatedAt] == "" {
			return fmt.Errorf("expected 'updated_at' to be set")
		}
		createdBy := fmt.Sprintf("teams.%d.created_by", teamsIdx)
		if instanceState.Attributes[createdBy] == "" {
			return fmt.Errorf("expected 'created_by' to be set")
		}
		updatedBy := fmt.Sprintf("teams.%d.updated_by", teamsIdx)
		if instanceState.Attributes[updatedBy] == "" {
			return fmt.Errorf("expected 'updated_by' to be set")
		}

		return nil
	}
}
