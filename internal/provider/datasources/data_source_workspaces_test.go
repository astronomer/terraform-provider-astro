package datasources_test

import (
	"fmt"
	"testing"

	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/lucsky/cuid"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_DataSourceWorkspaces(t *testing.T) {
	workspaceName := utils.GenerateTestResourceName(10)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true, false) + workspaces(workspaceName, ""),
				Check: resource.ComposeTestCheckFunc(
					// These checks are for the workspace data source (singular)
					resource.TestCheckResourceAttrSet("data.astro_workspace.test_data_workspace", "id"),
					resource.TestCheckResourceAttr("data.astro_workspace.test_data_workspace", "name", fmt.Sprintf("%v-1", workspaceName)),
					resource.TestCheckResourceAttrSet("data.astro_workspace.test_data_workspace", "description"),
					resource.TestCheckResourceAttr("data.astro_workspace.test_data_workspace", "cicd_enforced_default", "true"),
					resource.TestCheckResourceAttrSet("data.astro_workspace.test_data_workspace", "created_by.id"),
					resource.TestCheckResourceAttrSet("data.astro_workspace.test_data_workspace", "created_at"),
					resource.TestCheckResourceAttrSet("data.astro_workspace.test_data_workspace", "updated_by.id"),
					resource.TestCheckResourceAttrSet("data.astro_workspace.test_data_workspace", "updated_at"),

					// These checks are for the workspaces data source (plural)
					checkWorkspaces(workspaceName+"-1"),
					checkWorkspaces(workspaceName+"-2"),
				),
			},
			// The following tests are for filtering the workspaces data source
			{
				Config: astronomerprovider.ProviderConfig(t, true, false) + workspaces(workspaceName, `workspace_ids = [astro_workspace.test_workspace1.id]`),
				Check: resource.ComposeTestCheckFunc(
					checkWorkspaces(workspaceName + "-1"),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, true, false) + workspaces(workspaceName, fmt.Sprintf(`names = ["%v-1"]`, workspaceName)),
				Check: resource.ComposeTestCheckFunc(
					checkWorkspaces(workspaceName + "-1"),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, true, false) + workspaces(workspaceName, fmt.Sprintf(`names = ["%v"]`, cuid.New())),
				Check:  checkWorkspacesAreEmpty(),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, true, false) + workspaces(workspaceName, fmt.Sprintf(`workspace_ids = ["%v"]`, cuid.New())),
				Check:  checkWorkspacesAreEmpty(),
			},
		},
	})
}

func workspaces(name, filter string) string {
	return fmt.Sprintf(`
resource "astro_workspace" "test_workspace1" {
	name = "%v-1"
	description = "%v"
	cicd_enforced_default = true
}

resource "astro_workspace" "test_workspace2" {
	name = "%v-2"
	description = "%v"
	cicd_enforced_default = true
}

data astro_workspace "test_data_workspace" {
	depends_on = [astro_workspace.test_workspace1]
	id = astro_workspace.test_workspace1.id
}

data astro_workspaces "test_data_workspaces" {
	depends_on = [astro_workspace.test_workspace1, astro_workspace.test_workspace2]
	%v
}`, name, utils.TestResourceDescription, name, utils.TestResourceDescription, filter)
}

func checkWorkspacesAreEmpty() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		instanceState, numWorkspaces, err := utils.GetDataSourcesLength(s, "test_data_workspaces", "workspaces")
		if err != nil {
			return err
		}
		if numWorkspaces != 0 {
			return fmt.Errorf("expected workspaces to be 0, got %s", instanceState.Attributes["workspaces.#"])
		}
		return nil
	}
}

func checkWorkspaces(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		instanceState, numWorkspaces, err := utils.GetDataSourcesLength(s, "test_data_workspaces", "workspaces")
		if err != nil {
			return err
		}
		if numWorkspaces == 0 {
			return fmt.Errorf("expected workspaces to be greater or equal to 1, got %s", instanceState.Attributes["workspaces.#"])
		}
		workspaceIdx := -1
		for i := 0; i < numWorkspaces; i++ {
			idxName := fmt.Sprintf("workspaces.%d.name", i)
			if instanceState.Attributes[idxName] == name {
				workspaceIdx = i
				break
			}
		}
		if workspaceIdx == -1 {
			return fmt.Errorf("workspace %s not found", name)
		}
		description := fmt.Sprintf("workspaces.%d.description", workspaceIdx)
		if instanceState.Attributes[description] == "" {
			return fmt.Errorf("expected 'description' to be set")
		}
		cicdEnforcedDefault := fmt.Sprintf("workspaces.%d.cicd_enforced_default", workspaceIdx)
		if instanceState.Attributes[cicdEnforcedDefault] != "true" {
			return fmt.Errorf("expected 'cicd_enforced_default' to be true")
		}
		createdAt := fmt.Sprintf("workspaces.%d.created_at", workspaceIdx)
		if instanceState.Attributes[createdAt] == "" {
			return fmt.Errorf("expected 'created_at' to be set")
		}
		updatedAt := fmt.Sprintf("workspaces.%d.updated_at", workspaceIdx)
		if instanceState.Attributes[updatedAt] == "" {
			return fmt.Errorf("expected 'updated_at' to be set")
		}
		createdById := fmt.Sprintf("workspaces.%d.created_by.id", workspaceIdx)
		if instanceState.Attributes[createdById] == "" {
			return fmt.Errorf("expected 'created_by.id' to be set")
		}
		updatedById := fmt.Sprintf("workspaces.%d.updated_by.id", workspaceIdx)
		if instanceState.Attributes[updatedById] == "" {
			return fmt.Errorf("expected 'updated_by.id' to be set")
		}
		return nil
	}
}
