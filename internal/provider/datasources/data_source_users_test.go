package datasources_test

import (
	"fmt"
	"testing"

	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAcc_DataSourceUsers(t *testing.T) {
	tfVarName := "test_data_users"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) + users(tfVarName),
				Check: resource.ComposeTestCheckFunc(
					checkUsers(tfVarName),
				),
			},
		},
	})
}

func users(tfVarName string) string {
	return fmt.Sprintf(`
data astro_users "%v" {}`, tfVarName)
}

func checkUsers(tfVarName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		instanceState, numUsers, err := utils.GetDataSourcesLength(s, tfVarName, "users")
		if err != nil {
			return err
		}
		if numUsers == 0 {
			return fmt.Errorf("expected users to be greater or equal to 1, got %s", instanceState.Attributes["users.#"])
		}

		// Check the first user
		usersIdx := 0

		id := fmt.Sprintf("users.%d.id", usersIdx)
		if instanceState.Attributes[id] == "" {
			return fmt.Errorf("expected 'id' to be set")
		}
		username := fmt.Sprintf("users.%d.username", usersIdx)
		if instanceState.Attributes[username] == "" {
			return fmt.Errorf("expected 'username' to be set")
		}
		fullName := fmt.Sprintf("users.%d.full_name", usersIdx)
		if instanceState.Attributes[fullName] == "" {
			return fmt.Errorf("expected 'full_name' to be set")
		}
		status := fmt.Sprintf("users.%d.status", usersIdx)
		if instanceState.Attributes[status] == "" {
			return fmt.Errorf("expected 'status' to be set")
		}
		avatarUrl := fmt.Sprintf("users.%d.avatar_url", usersIdx)
		if instanceState.Attributes[avatarUrl] == "" {
			return fmt.Errorf("expected 'avatar_url' to be set")
		}
		organizationRole := fmt.Sprintf("teams.%d.organization_role", usersIdx)
		if instanceState.Attributes[organizationRole] == "" {
			return fmt.Errorf("expected 'organization_role' to be set")
		}
		createdAt := fmt.Sprintf("users.%d.created_at", usersIdx)
		if instanceState.Attributes[createdAt] == "" {
			return fmt.Errorf("expected 'created_at' to be set")
		}
		updatedAt := fmt.Sprintf("users.%d.updated_at", usersIdx)
		if instanceState.Attributes[updatedAt] == "" {
			return fmt.Errorf("expected 'updated_at' to be set")
		}

		return nil
	}
}
