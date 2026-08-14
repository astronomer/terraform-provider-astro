package resources_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestAcc_ResourceCustomRole(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	customRole1Name := fmt.Sprintf("%v_deployment", namePrefix)
	customRole2Name := fmt.Sprintf("%v_dag", namePrefix)
	customRole3Name := fmt.Sprintf("%v_updated", namePrefix)
	description1 := "Test custom role description"
	description2 := utils.TestResourceDescription

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			// Check that custom roles have been removed
			testAccCheckCustomRoleExistence(t, customRole1Name, false),
			testAccCheckCustomRoleExistence(t, customRole2Name, false),
			testAccCheckCustomRoleExistence(t, customRole3Name, false),
		),
		Steps: []resource.TestStep{
			// Test failure: invalid scope type
			{
				Config:      astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + customRole("test", customRole1Name, description1, "WORKSPACE", []string{"deployment.get"}),
				ExpectError: regexp.MustCompile("Attribute scope_type value must be one of"),
			},
			// Test failure: invalid scope type (ORGANIZATION)
			{
				Config:      astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + customRole("test", customRole1Name, description1, "ORGANIZATION", []string{"deployment.get"}),
				ExpectError: regexp.MustCompile("Attribute scope_type value must be one of"),
			},
			// Test failure: empty name
			{
				Config:      astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + customRole("test", "", description1, "DEPLOYMENT", []string{"deployment.get"}),
				ExpectError: regexp.MustCompile("Attribute name string length must be at least 1"),
			},
			// Create the custom role with DEPLOYMENT scope
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + customRole("test", customRole1Name, description1, "DEPLOYMENT", []string{"deployment.get"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("astro_custom_role.test", "name", customRole1Name),
					resource.TestCheckResourceAttr("astro_custom_role.test", "description", description1),
					resource.TestCheckResourceAttr("astro_custom_role.test", "scope_type", "DEPLOYMENT"),
					resource.TestCheckResourceAttr("astro_custom_role.test", "permissions.#", "1"),
					resource.TestCheckResourceAttrSet("astro_custom_role.test", "id"),
					resource.TestCheckResourceAttrSet("astro_custom_role.test", "created_at"),
					resource.TestCheckResourceAttrSet("astro_custom_role.test", "updated_at"),
					resource.TestCheckResourceAttrSet("astro_custom_role.test", "created_by.id"),
					resource.TestCheckResourceAttrSet("astro_custom_role.test", "updated_by.id"),
					// Check via API that custom role exists
					testAccCheckCustomRoleExistence(t, customRole1Name, true),
				),
			},
			// Update name, description, and permissions
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + customRole("test", customRole3Name, description2, "DEPLOYMENT", []string{"deployment.get", "deployment.delete", "deployment.update"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("astro_custom_role.test", "name", customRole3Name),
					resource.TestCheckResourceAttr("astro_custom_role.test", "description", description2),
					resource.TestCheckResourceAttr("astro_custom_role.test", "scope_type", "DEPLOYMENT"),
					resource.TestCheckResourceAttr("astro_custom_role.test", "permissions.#", "3"),
					// Check via API that custom role exists
					testAccCheckCustomRoleExistence(t, customRole3Name, true),
				),
			},
			// Change scope_type (should replace the resource)
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + customRole("test", customRole2Name, description1, "DAG", []string{"dag.airflow.dag.get"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("astro_custom_role.test", "name", customRole2Name),
					resource.TestCheckResourceAttr("astro_custom_role.test", "description", description1),
					resource.TestCheckResourceAttr("astro_custom_role.test", "scope_type", "DAG"),
					resource.TestCheckResourceAttr("astro_custom_role.test", "permissions.#", "1"),
					// Check via API that new custom role exists
					testAccCheckCustomRoleExistence(t, customRole2Name, true),
					// Check via API that old custom role was removed
					testAccCheckCustomRoleExistence(t, customRole3Name, false),
				),
			},
			// Import existing custom role and check it is correctly imported
			{
				ResourceName:      "astro_custom_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ResourceCustomRoleWithRestrictedWorkspaces(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	customRoleName := fmt.Sprintf("%v_restricted", namePrefix)
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckCustomRoleExistence(t, customRoleName, false),
		),
		Steps: []resource.TestStep{
			// Create custom role with restricted workspaces
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + customRoleWithRestrictedWorkspaces("test", customRoleName, "Restricted role", "DEPLOYMENT", []string{"deployment.get"}, []string{workspaceId}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("astro_custom_role.test", "name", customRoleName),
					resource.TestCheckResourceAttr("astro_custom_role.test", "scope_type", "DEPLOYMENT"),
					resource.TestCheckResourceAttr("astro_custom_role.test", "permissions.#", "1"),
					resource.TestCheckResourceAttr("astro_custom_role.test", "restricted_workspace_ids.#", "1"),
					testAccCheckCustomRoleExistence(t, customRoleName, true),
				),
			},
			// Remove restricted workspaces
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + customRoleWithRestrictedWorkspaces("test", customRoleName, "Unrestricted role", "DEPLOYMENT", []string{"deployment.get"}, []string{}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("astro_custom_role.test", "name", customRoleName),
					resource.TestCheckResourceAttr("astro_custom_role.test", "description", "Unrestricted role"),
					resource.TestCheckResourceAttr("astro_custom_role.test", "restricted_workspace_ids.#", "0"),
					testAccCheckCustomRoleExistence(t, customRoleName, true),
				),
			},
		},
	})
}

func TestAcc_CustomRoleRemovedOutsideOfTerraform(t *testing.T) {
	customRoleName := utils.GenerateTestResourceName(10)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckCustomRoleExistence(t, customRoleName, false),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + customRoleWithVariableName(),
				ConfigVariables: map[string]config.Variable{
					"name": config.StringVariable(customRoleName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("astro_custom_role.test", "name", customRoleName),
					resource.TestCheckResourceAttr("astro_custom_role.test", "description", utils.TestResourceDescription),
					// Check via API that custom role exists
					testAccCheckCustomRoleExistence(t, customRoleName, true),
				),
			},
			{
				PreConfig: func() { deleteCustomRoleOutsideOfTerraform(t, customRoleName) },
				Config:    astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + customRoleWithVariableName(),
				ConfigVariables: map[string]config.Variable{
					"name": config.StringVariable(customRoleName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("astro_custom_role.test", "name", customRoleName),
					resource.TestCheckResourceAttr("astro_custom_role.test", "description", utils.TestResourceDescription),
					// Check via API that custom role exists
					testAccCheckCustomRoleExistence(t, customRoleName, true),
				),
			},
		},
	})
}

// TestAcc_ResourceCustomRoleDirectAccessTokenConflict verifies that changing the permissions
// of a custom role that is referenced by a Direct Access Token fails during `terraform plan`
// (via ModifyPlan) rather than surfacing only as a confusing apply-time API error. Direct
// Access Tokens are created outside of Terraform (the astro_api_token resource cannot create
// them), so the conflicting token is created directly via the IAM client.
func TestAcc_ResourceCustomRoleDirectAccessTokenConflict(t *testing.T) {
	customRoleName := utils.GenerateTestResourceName(10)
	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
	organizationId := os.Getenv("HOSTED_ORGANIZATION_ID")

	var directAccessTokenId string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckCustomRoleExistence(t, customRoleName, false),
		Steps: []resource.TestStep{
			// Create the custom role, then attach a Direct Access Token to it outside of Terraform.
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + customRole("test", customRoleName, utils.TestResourceDescription, "DEPLOYMENT", []string{"deployment.get"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomRoleExistence(t, customRoleName, true),
					testAccCreateDirectAccessTokenForRole(t, organizationId, deploymentId, customRoleName, &directAccessTokenId),
				),
			},
			// Adding a permission to the role should now fail at plan time, not apply time.
			{
				Config:      astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + customRole("test", customRoleName, utils.TestResourceDescription, "DEPLOYMENT", []string{"deployment.get", "deployment.update"}),
				ExpectError: regexp.MustCompile("Custom role has Direct Access Tokens attached"),
			},
			// No-op step (state is unchanged from step 1 since step 2 failed at plan) that
			// deletes the out-of-band token before Terraform's end-of-test destroy runs,
			// since the API also refuses to delete a role that a token still references.
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + customRole("test", customRoleName, utils.TestResourceDescription, "DEPLOYMENT", []string{"deployment.get"}),
				Check:  testAccDeleteDirectAccessTokenIfExists(t, &directAccessTokenId),
			},
		},
	})
}

// testAccCreateDirectAccessTokenForRole creates a Direct Access Token scoped to deploymentId
// and assigned to the named custom role, bypassing Terraform (mirroring how a customer would
// create one via `astro deployment token create --direct-access` or the UI).
func testAccCreateDirectAccessTokenForRole(t *testing.T, organizationId, deploymentId, customRoleName string, tokenId *string) func(state *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		client, err := utils.GetTestHostedIamClient()
		assert.NoError(t, err)

		ctx := context.Background()
		// CreateApiTokenRequest.Role takes the role's name, not its ID (token role
		// assignments are keyed by role name, matching what ApiTokenRole.Role returns).
		createResp, err := client.CreateApiTokenWithResponse(ctx, organizationId, iam.CreateApiTokenRequest{
			Name:     fmt.Sprintf("%s-dat", customRoleName),
			Type:     iam.CreateApiTokenRequestTypeDEPLOYMENT,
			EntityId: &deploymentId,
			Role:     customRoleName,
			Kind:     lo.ToPtr(iam.CreateApiTokenRequestKind(iam.ApiTokenKindDIRECTACCESS)),
		})
		if err != nil {
			return fmt.Errorf("failed to create Direct Access Token: %w", err)
		}
		if createResp.JSON200 == nil {
			status, diagnostic := clients.NormalizeAPIError(ctx, createResp.HTTPResponse, createResp.Body)
			return fmt.Errorf("failed to create Direct Access Token, status: %v, err: %v", status, diagnostic.Detail())
		}
		*tokenId = createResp.JSON200.Id
		return nil
	}
}

func testAccDeleteDirectAccessTokenIfExists(t *testing.T, tokenId *string) func(state *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		if tokenId == nil || *tokenId == "" {
			return nil
		}
		client, err := utils.GetTestHostedIamClient()
		assert.NoError(t, err)

		ctx := context.Background()
		_, err = client.DeleteApiTokenWithResponse(ctx, os.Getenv("HOSTED_ORGANIZATION_ID"), *tokenId)
		return err
	}
}

func customRoleWithVariableName() string {
	return fmt.Sprintf(`
variable "name" {
	type = string
}

resource "astro_custom_role" "test" {
	name = var.name
	description = "%s"
	scope_type = "DEPLOYMENT"
	permissions = ["deployment.get"]
}`, utils.TestResourceDescription)
}

func customRole(tfVarName, name, description, scopeType string, permissions []string) string {
	permissionsStr := ""
	for _, perm := range permissions {
		permissionsStr += fmt.Sprintf(`"%s",`, perm)
	}
	// Remove trailing comma
	if len(permissionsStr) > 0 {
		permissionsStr = permissionsStr[:len(permissionsStr)-1]
	}

	return fmt.Sprintf(`
resource "astro_custom_role" "%s" {
	name = "%s"
	description = "%s"
	scope_type = "%s"
	permissions = [%s]
}
`, tfVarName, name, description, scopeType, permissionsStr)
}

func customRoleWithRestrictedWorkspaces(tfVarName, name, description, scopeType string, permissions []string, restrictedWorkspaceIds []string) string {
	permissionsStr := ""
	for _, perm := range permissions {
		permissionsStr += fmt.Sprintf(`"%s",`, perm)
	}
	if len(permissionsStr) > 0 {
		permissionsStr = permissionsStr[:len(permissionsStr)-1]
	}

	workspaceIdsStr := ""
	for _, wsId := range restrictedWorkspaceIds {
		workspaceIdsStr += fmt.Sprintf(`"%s",`, wsId)
	}
	if len(workspaceIdsStr) > 0 {
		workspaceIdsStr = workspaceIdsStr[:len(workspaceIdsStr)-1]
	}

	return fmt.Sprintf(`
resource "astro_custom_role" "%s" {
	name = "%s"
	description = "%s"
	scope_type = "%s"
	permissions = [%s]
	restricted_workspace_ids = [%s]
}
`, tfVarName, name, description, scopeType, permissionsStr, workspaceIdsStr)
}

func deleteCustomRoleOutsideOfTerraform(t *testing.T, name string) {
	t.Helper()

	client, err := utils.GetTestHostedIamClient()
	assert.NoError(t, err)

	ctx := context.Background()
	// List roles to find the custom role by name
	resp, err := client.ListRolesWithResponse(ctx, os.Getenv("HOSTED_ORGANIZATION_ID"), &iam.ListRolesParams{})
	if err != nil {
		assert.NoError(t, err)
	}

	// Find custom role by name
	var customRoleId string
	if resp.JSON200 != nil && len(resp.JSON200.Roles) > 0 {
		for _, role := range resp.JSON200.Roles {
			if role.Name == name {
				customRoleId = role.Id
				break
			}
		}
	}

	assert.NotEmpty(t, customRoleId, "custom role should exist but list roles did not find it")
	_, err = client.DeleteCustomRoleWithResponse(ctx, os.Getenv("HOSTED_ORGANIZATION_ID"), customRoleId)
	assert.NoError(t, err)
}

func testAccCheckCustomRoleExistence(t *testing.T, name string, shouldExist bool) func(state *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		client, err := utils.GetTestHostedIamClient()
		assert.NoError(t, err)

		ctx := context.Background()
		resp, err := client.ListRolesWithResponse(ctx, os.Getenv("HOSTED_ORGANIZATION_ID"), &iam.ListRolesParams{})
		if err != nil {
			return fmt.Errorf("failed to list roles: %w", err)
		}
		if resp.JSON200 == nil {
			status, diag := clients.NormalizeAPIError(ctx, resp.HTTPResponse, resp.Body)
			return fmt.Errorf("response JSON200 is nil status: %v, err: %v", status, diag.Detail())
		}

		// Find custom role by name
		var found bool
		if len(resp.JSON200.Roles) > 0 {
			for _, role := range resp.JSON200.Roles {
				if role.Name == name {
					found = true
					break
				}
			}
		}

		if shouldExist {
			if !found {
				return fmt.Errorf("custom role %s should exist", name)
			}
		} else {
			if found {
				return fmt.Errorf("custom role %s should not exist", name)
			}
		}
		return nil
	}
}
