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
				Config: astronomerprovider.ProviderConfig(t, true) + users(tfVarName),
				Check: resource.ComposeTestCheckFunc(
					checkUsers(tfVarName, false, false),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, true) + usersFilterWorkspaceId(tfVarName, tfWorkspaceId),
				Check: resource.ComposeTestCheckFunc(
					checkUsers(tfVarName, true, false),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, true) + usersFilterDeploymentId(tfVarName, tfDeploymentId),
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

func usersFilter(tfVarName string, filter string, filterId string) string {
	return fmt.Sprintf(`
data astro_users "%v" {
%v = "%v"
}`, tfVarName, filter, filterId)
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
		organizationRole := fmt.Sprintf("users.%d.organization_role", usersIdx)
		if instanceState.Attributes[organizationRole] == "" {
			return fmt.Errorf("expected 'organization_role' to be set")
		}
		if filterWorkspaceId {
			workspaceRoles := fmt.Sprintf("users.%d.workspace_roles.0.role", usersIdx)
			fmt.Printf("****all attributes: %s", instanceState.Attributes)
			fmt.Printf("****workspace roles: %s", instanceState.Attributes[workspaceRoles])
			if instanceState.Attributes[workspaceRoles] == "" {
				return fmt.Errorf("expected 'workspace_roles' to be set")
			}
			if len(instanceState.Attributes[workspaceRoles]) == 0 {
				return fmt.Errorf("expected 'workspace_roles' to be set: %s", instanceState.Attributes[workspaceRoles])
			}
		}
		if filterDeploymentId {
			deploymentRoles := fmt.Sprintf("users.%d.deployment_roles.0.role", usersIdx)
			fmt.Printf("****all attributes: %s", instanceState.Attributes)
			fmt.Printf("****deployment roles: %s", instanceState.Attributes[deploymentRoles])
			if instanceState.Attributes[deploymentRoles] == "" {
				return fmt.Errorf("expected 'deployment_roles' to be set")
			}
			if len(instanceState.Attributes[deploymentRoles]) == 0 {
				fmt.Printf("****deploymentRoles: %s", instanceState.Attributes[deploymentRoles])
				return fmt.Errorf("expected 'deployment_roles' to be set: %s", instanceState.Attributes[deploymentRoles])
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
