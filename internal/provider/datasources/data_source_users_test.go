package datasources_test

import (
	"fmt"
	"os"
	"testing"

	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAcc_DataSourceUsers(t *testing.T) {
	tfVarName := "test_data_users"
	tfWorkspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	tfDeploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + users(tfVarName),
				Check: resource.ComposeTestCheckFunc(
					checkUsers(tfVarName, false, false),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + usersFilterWorkspaceId(tfVarName, tfWorkspaceId),
				Check: resource.ComposeTestCheckFunc(
					checkUsers(tfVarName, true, false),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + usersFilterDeploymentId(tfVarName, tfDeploymentId),
				Check: resource.ComposeTestCheckFunc(
					checkUsers(tfVarName, false, true),
				),
			},
		},
	})
}

func users(tfVarName string) string {
	return fmt.Sprintf(`
data astro_users "%v" {}`, tfVarName)
}

func usersFilterWorkspaceId(tfVarName string, workspaceId string) string {
	return fmt.Sprintf(`
data astro_users "%v" {
workspace_id = "%v"
}`, tfVarName, workspaceId)
}

func usersFilterDeploymentId(tfVarName string, deploymentId string) string {
	return fmt.Sprintf(`
data astro_users "%v" {
deployment_id = "%v"
}`, tfVarName, deploymentId)
}

func checkUsers(tfVarName string, filterWorkspaceId bool, filterDeploymentId bool) resource.TestCheckFunc {
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
		status := fmt.Sprintf("users.%d.status", usersIdx)
		if instanceState.Attributes[status] == "" {
			return fmt.Errorf("expected 'status' to be set")
		}
		organizationRole := fmt.Sprintf("users.%d.organization_role", usersIdx)
		if instanceState.Attributes[organizationRole] == "" {
			return fmt.Errorf("expected 'organization_role' to be set")
		}
		if filterWorkspaceId {
			workspaceRole := fmt.Sprintf("users.%d.workspace_roles.0.role", usersIdx)
			workspaceId := fmt.Sprintf("users.%d.workspace_roles.0.workspace_id", usersIdx)

			if instanceState.Attributes[workspaceRole] == "" {
				return fmt.Errorf("expected 'workspace_roles' to be set")
			}
			if instanceState.Attributes[workspaceId] == "" {
				return fmt.Errorf("expected 'workspace_id' to be set")
			}
		}
		if filterDeploymentId {
			deploymentRole := fmt.Sprintf("users.%d.deployment_roles.0.role", usersIdx)
			deploymentId := fmt.Sprintf("users.%d.deployment_roles.0.deployment_id", usersIdx)

			if instanceState.Attributes[deploymentRole] == "" {
				return fmt.Errorf("expected 'deployment_roles' to be set")
			}
			if instanceState.Attributes[deploymentId] == "" {
				return fmt.Errorf("expected 'deployment_id' to be set")
			}
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
