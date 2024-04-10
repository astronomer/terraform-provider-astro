package resources_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	astronomerprovider "github.com/astronomer/astronomer-terraform-provider/internal/provider"
	"github.com/astronomer/astronomer-terraform-provider/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_Resource(t *testing.T) {
	workspaceName := utils.GenerateTestResourceName(10)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig() + workspace(fmt.Sprintf("%v-1", workspaceName), "test", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("astronomer_workspace.test", "name", fmt.Sprintf("%v-1", workspaceName)),
					resource.TestCheckResourceAttr("astronomer_workspace.test", "description", "test"),
					resource.TestCheckResourceAttr("astronomer_workspace.test", "cicd_enforced_default", "false"),
				),
			},
			// CHANGE PROPERTIES
			{
				Config: astronomerprovider.ProviderConfig() + workspace(fmt.Sprintf("%v-2", workspaceName), utils.TestResourceDescription, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("astronomer_workspace.test", "name", fmt.Sprintf("%v-2", workspaceName)),
					resource.TestCheckResourceAttr("astronomer_workspace.test", "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr("astronomer_workspace.test", "cicd_enforced_default", "true"),
				),
			},
			// IMPORT
			{
				ResourceName:      "astronomer_workspace.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_WorkspaceRemovedOutsideOfTerraform(t *testing.T) {
	workspaceName := utils.GenerateTestResourceName(10)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckWorkspaceExistence(t, workspaceName, false),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig() + workspaceWithVariableName(),
				ConfigVariables: map[string]config.Variable{
					"name": config.StringVariable(workspaceName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("astronomer_workspace.test", "name", workspaceName),
					resource.TestCheckResourceAttr("astronomer_workspace.test", "description", utils.TestResourceDescription),
					testAccCheckWorkspaceExistence(t, workspaceName, true),
				),
			},
			{
				PreConfig: func() { deleteWorkspaceOutsideOfTerraform(t, workspaceName) },
				Config:    astronomerprovider.ProviderConfig() + workspaceWithVariableName(),
				ConfigVariables: map[string]config.Variable{
					"name": config.StringVariable(workspaceName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("astronomer_workspace.test", "name", workspaceName),
					resource.TestCheckResourceAttr("astronomer_workspace.test", "description", utils.TestResourceDescription),
					testAccCheckWorkspaceExistence(t, workspaceName, true),
				),
			},
		},
	})
}

func workspaceWithVariableName() string {
	return fmt.Sprintf(`
variable "name" {
	type = string
}

resource "astronomer_workspace" "test" {
	name = var.name
	description = "%s"
	cicd_enforced_default = true
}`, utils.TestResourceDescription)
}

func workspace(name, description string, cicdEnforcedDefault bool) string {
	return fmt.Sprintf(`
resource "astronomer_workspace" "test" {
	name = "%s"
	description = "%s"
	cicd_enforced_default = %t
}
`, name, description, cicdEnforcedDefault)
}

func deleteWorkspaceOutsideOfTerraform(t *testing.T, name string) {
	t.Helper()

	client, err := utils.GetTestPlatformClient()
	assert.NoError(t, err)

	ctx := context.Background()
	resp, err := client.ListWorkspacesWithResponse(ctx, os.Getenv("ASTRO_ORGANIZATION_ID"), &platform.ListWorkspacesParams{
		Names: &[]string{name},
	})
	if err != nil {
		assert.NoError(t, err)
	}
	assert.True(t, len(resp.JSON200.Workspaces) >= 1, "workspace should exist but list workspaces did not find it")
	_, err = client.DeleteWorkspaceWithResponse(ctx, os.Getenv("ASTRO_ORGANIZATION_ID"), resp.JSON200.Workspaces[0].Id)
	assert.NoError(t, err)
}

func testAccCheckWorkspaceExistence(t *testing.T, name string, shouldExist bool) func(state *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		client, err := utils.GetTestPlatformClient()
		assert.NoError(t, err)

		ctx := context.Background()
		resp, err := client.ListWorkspacesWithResponse(ctx, os.Getenv("ASTRO_ORGANIZATION_ID"), &platform.ListWorkspacesParams{
			Names: &[]string{name},
		})
		if err != nil {
			return fmt.Errorf("failed to list workspaces: %w", err)
		}
		if shouldExist {
			if len(resp.JSON200.Workspaces) != 1 {
				return fmt.Errorf("workspace %s should exist", name)
			}
		} else {
			if len(resp.JSON200.Workspaces) != 0 {
				return fmt.Errorf("workspace %s should not exist", name)
			}
		}
		return nil
	}
}
