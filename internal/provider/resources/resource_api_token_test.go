package resources_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

// test organization api token
// test workpsace api token
// test deployment api token

// within each test:
// - create the resource
// - check that the resource was created
// - update the resource
// - check that the resource was updated
// - change the resource type
// - check that the resource was destroyed and recreated
// - import existing resource
// - check that the resource was imported
// - check that for each role, the role matches the entity type

func TestAcc_ResourceOrganizationApiToken(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)

	organizationId := os.Getenv("HOSTED_ORGANIZATION_ID")
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")

	apiTokenName := fmt.Sprintf("%v_org", namePrefix)
	resourceVar := fmt.Sprintf("astro_api_token.%v", apiTokenName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			// Check that the organization api token has been removed
			testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{organization: true, shouldExist: false}),
		),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: utils.TestResourceDescription,
					Type:        "ORGANIZATION",
					Roles: []apiTokenRole{
						{
							Role:       "ORGANIZATION_OWNER",
							EntityId:   organizationId,
							EntityType: "ORGANIZATION",
						},
						{
							Role:       "WORKSPACE_OWNER",
							EntityId:   workspaceId,
							EntityType: "WORKSPACE",
						},
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: "DEPLOYMENT",
						},
					},
					ExpiryPeriodInDays: 30,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", apiTokenName),
					resource.TestCheckResourceAttr(resourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(resourceVar, "type", "ORGANIZATION"),
					resource.TestCheckResourceAttrSet(resourceVar, "short_token"),
					resource.TestCheckResourceAttrSet(resourceVar, "start_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					resource.TestCheckResourceAttr(resourceVar, "expiry_period_in_days", "30"),
					resource.TestCheckResourceAttr(resourceVar, "roles.#", "3"),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_id", organizationId),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_type", "ORGANIZATION"),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.role", "ORGANIZATION_OWNER"),
					resource.TestCheckResourceAttr(resourceVar, "roles.1.entity_id", workspaceId),
					resource.TestCheckResourceAttr(resourceVar, "roles.1.entity_type", "WORKSPACE"),
					resource.TestCheckResourceAttr(resourceVar, "roles.1.role", "WORKSPACE_OWNER"),
					resource.TestCheckResourceAttr(resourceVar, "roles.2.entity_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "roles.2.entity_type", "DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceVar, "roles.2.role", "DEPLOYMENT_ADMIN"),
					// Check via API that organization api token exists
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{organization: true, shouldExist: true}),
				),
			},
			// Change properties and check they have been updated in terraform state
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: "new description",
					Type:        "ORGANIZATION",
					Roles: []apiTokenRole{
						{
							Role:       "ORGANIZATION_OWNER",
							EntityId:   organizationId,
							EntityType: "ORGANIZATION",
						},
						{
							Role:       "WORKSPACE_OWNER",
							EntityId:   workspaceId,
							EntityType: "WORKSPACE",
						},
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: "DEPLOYMENT",
						},
					},
					ExpiryPeriodInDays: 60,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "description", "new description"),
					resource.TestCheckResourceAttr(resourceVar, "expiry_period_in_days", "60"),
					// Check via API that organization api token exists
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{organization: true, shouldExist: true}),
				),
			},
			// Change the resource type and remove roles
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: utils.TestResourceDescription,
					Type:        "WORKSPACE",
					Roles: []apiTokenRole{
						{
							Role:       "WORKSPACE_OWNER",
							EntityId:   workspaceId,
							EntityType: "WORKSPACE",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "type", "WORKSPACE"),
					resource.TestCheckResourceAttr(resourceVar, "roles.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_id", workspaceId),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_type", "WORKSPACE"),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.role", "WORKSPACE_OWNER"),
					// Check via API that api token was destroyed and recreated
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{workspace: true, shouldExist: true}),
				),
			},
			// Change resource type back to ORGANIZATION
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: utils.TestResourceDescription,
					Type:        "ORGANIZATION",
					Roles: []apiTokenRole{
						{
							Role:       "ORGANIZATION_OWNER",
							EntityId:   organizationId,
							EntityType: "ORGANIZATION",
						},
						{
							Role:       "WORKSPACE_OWNER",
							EntityId:   workspaceId,
							EntityType: "WORKSPACE",
						},
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: "DEPLOYMENT",
						},
					},
					ExpiryPeriodInDays: 30,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "type", "ORGANIZATION"),
					resource.TestCheckResourceAttr(resourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(resourceVar, "expiry_period_in_days", "30"),
					// Check via API that organization api token exists
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{organization: true, shouldExist: true}),
				),
			},
			// Import existing api token and check it is correctly imported
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func TestAcc_ResourceWorkspaceApiToken(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)

	organizationId := os.Getenv("HOSTED_ORGANIZATION_ID")
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")

	apiTokenName := fmt.Sprintf("%v_org", namePrefix)
	resourceVar := fmt.Sprintf("astro_api_token.%v", apiTokenName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			// Check that the organization api token has been removed
			testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{workspace: true, shouldExist: false}),
		),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: utils.TestResourceDescription,
					Type:        "WORKSPACE",
					Roles: []apiTokenRole{
						{
							Role:       "WORKSPACE_OWNER",
							EntityId:   workspaceId,
							EntityType: "WORKSPACE",
						},
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: "DEPLOYMENT",
						},
					},
					ExpiryPeriodInDays: 30,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", apiTokenName),
					resource.TestCheckResourceAttr(resourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(resourceVar, "type", "WORKSPACE"),
					resource.TestCheckResourceAttrSet(resourceVar, "short_token"),
					resource.TestCheckResourceAttrSet(resourceVar, "start_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					resource.TestCheckResourceAttr(resourceVar, "expiry_period_in_days", "30"),
					resource.TestCheckResourceAttr(resourceVar, "roles.#", "2"),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_id", workspaceId),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_type", "WORKSPACE"),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.role", "WORKSPACE_OWNER"),
					resource.TestCheckResourceAttr(resourceVar, "roles.1.entity_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "roles.1.entity_type", "DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceVar, "roles.1.role", "DEPLOYMENT_ADMIN"),
					// Check via API that organization api token exists
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{workspace: true, shouldExist: true}),
				),
			},
			// Change properties and check they have been updated in terraform state
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: "new description",
					Type:        "WORKSPACE",
					Roles: []apiTokenRole{
						{
							Role:       "WORKSPACE_OWNER",
							EntityId:   workspaceId,
							EntityType: "WORKSPACE",
						},
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: "DEPLOYMENT",
						},
					},
					ExpiryPeriodInDays: 60,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "description", "new description"),
					resource.TestCheckResourceAttr(resourceVar, "expiry_period_in_days", "60"),
					// Check via API that organization api token exists
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{workspace: true, shouldExist: true}),
				),
			},
			// Change the resource type and remove roles
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: utils.TestResourceDescription,
					Type:        "ORGANIZATION",
					Roles: []apiTokenRole{
						{
							Role:       "ORGANIZATION_OWNER",
							EntityId:   organizationId,
							EntityType: "ORGANIZATION",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "type", "ORGANIZATION"),
					resource.TestCheckResourceAttr(resourceVar, "roles.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_id", organizationId),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_type", "ORGANIZATION"),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.role", "ORGANIZATION_OWNER"),
					// Check via API that api token was destroyed and recreated
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{organization: true, shouldExist: true}),
				),
			},
			// Change resource type back to WORKSPACE
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: utils.TestResourceDescription,
					Type:        "WORKSPACE",
					Roles: []apiTokenRole{
						{
							Role:       "WORKSPACE_OWNER",
							EntityId:   workspaceId,
							EntityType: "WORKSPACE",
						},
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: "DEPLOYMENT",
						},
					},
					ExpiryPeriodInDays: 30,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "type", "WORKSPACE"),
					resource.TestCheckResourceAttr(resourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(resourceVar, "expiry_period_in_days", "30"),
					// Check via API that organization api token exists
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{workspace: true, shouldExist: true}),
				),
			},
			// Import existing api token and check it is correctly imported
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func TestAcc_ResourceDeploymentApiToken(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)

	organizationId := os.Getenv("HOSTED_ORGANIZATION_ID")
	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")

	apiTokenName := fmt.Sprintf("%v_org", namePrefix)
	resourceVar := fmt.Sprintf("astro_api_token.%v", apiTokenName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			// Check that the organization api token has been removed
			testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{deployment: true, shouldExist: false}),
		),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: utils.TestResourceDescription,
					Type:        "DEPLOYMENT",
					Roles: []apiTokenRole{
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: "DEPLOYMENT",
						},
					},
					ExpiryPeriodInDays: 30,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", apiTokenName),
					resource.TestCheckResourceAttr(resourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(resourceVar, "type", "DEPLOYMENT"),
					resource.TestCheckResourceAttrSet(resourceVar, "short_token"),
					resource.TestCheckResourceAttrSet(resourceVar, "start_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					resource.TestCheckResourceAttr(resourceVar, "expiry_period_in_days", "30"),
					resource.TestCheckResourceAttr(resourceVar, "roles.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_type", "DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.role", "DEPLOYMENT_ADMIN"),
					// Check via API that organization api token exists
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{deployment: true, shouldExist: true}),
				),
			},
			// Change properties and check they have been updated in terraform state
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: "new description",
					Type:        "DEPLOYMENT",
					Roles: []apiTokenRole{
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: "DEPLOYMENT",
						},
					},
					ExpiryPeriodInDays: 60,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "description", "new description"),
					resource.TestCheckResourceAttr(resourceVar, "expiry_period_in_days", "60"),
					// Check via API that organization api token exists
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{deployment: true, shouldExist: true}),
				),
			},
			// Change the resource type
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: utils.TestResourceDescription,
					Type:        "ORGANIZATION",
					Roles: []apiTokenRole{
						{
							Role:       "ORGANIZATION_OWNER",
							EntityId:   organizationId,
							EntityType: "ORGANIZATION",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "type", "ORGANIZATION"),
					resource.TestCheckResourceAttr(resourceVar, "roles.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_id", organizationId),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_type", "ORGANIZATION"),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.role", "ORGANIZATION_OWNER"),
					// Check via API that api token was destroyed and recreated
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{organization: true, shouldExist: true}),
				),
			},
			// Change resource type back to DEPLOYMENT
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: utils.TestResourceDescription,
					Type:        "DEPLOYMENT",
					Roles: []apiTokenRole{
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: "DEPLOYMENT",
						},
					},
					ExpiryPeriodInDays: 30,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "type", "DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(resourceVar, "expiry_period_in_days", "30"),
					// Check via API that organization api token exists
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{deployment: true, shouldExist: true}),
				),
			},
			// Import existing api token and check it is correctly imported
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

type apiTokenRole struct {
	Role       string
	EntityId   string
	EntityType string
}

type apiTokenInput struct {
	Name               string
	Description        string
	Type               string
	Roles              []apiTokenRole
	ExpiryPeriodInDays int
}

func apiToken(input apiTokenInput) string {
	roles := lo.Map(input.Roles, func(role apiTokenRole, _ int) string {
		return fmt.Sprintf(`
		{
			role = "%v"
			entity_id = "%v"
			entity_type = "%v"
		}`, role.Role, role.EntityId, role.EntityType)
	})

	var rolesString string
	if len(input.Roles) > 0 {
		rolesString = fmt.Sprintf("roles = [%v]", strings.Join(roles, ", "))
	}

	return fmt.Sprintf(`
resource astro_api_token "%v" {
	name = "%v"
	description = "%s"
	type = "%s"
	%v
	expiry_period_in_days = %v
}`, input.Name, input.Name, input.Description, input.Type, rolesString, input.ExpiryPeriodInDays)
}

type checkApiTokensExistenceInput struct {
	organization bool
	workspace    bool
	deployment   bool
	shouldExist  bool
}

func testAccCheckApiTokenExistence(t *testing.T, input checkApiTokensExistenceInput) func(s *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		client, err := utils.GetTestIamClient(true)
		assert.NoError(t, err)

		organizationId := os.Getenv("HOSTED_ORGANIZATION_ID")

		ctx := context.Background()

		apiTokensParams := &iam.ListApiTokensParams{}

		if input.organization {
			apiTokensParams.IncludeOnlyOrganizationTokens = lo.ToPtr(true)
		} else if input.workspace {
			workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
			apiTokensParams.WorkspaceId = lo.ToPtr(workspaceId)
		} else if input.deployment {
			deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
			apiTokensParams.DeploymentId = lo.ToPtr(deploymentId)
		}

		resp, err := client.ListApiTokensWithResponse(ctx, organizationId, apiTokensParams)
		if err != nil {
			return fmt.Errorf("failed to list api tokens: %v", err)
		}
		if resp == nil {
			return fmt.Errorf("nil response from list api tokens")
		}
		if resp.JSON200 == nil {
			status, diag := clients.NormalizeAPIError(ctx, resp.HTTPResponse, resp.Body)
			return fmt.Errorf("response JSON200 is nil status: %v, err: %v", status, diag.Detail())
		}

		if input.shouldExist {
			if len(resp.JSON200.Tokens) != 1 {
				return fmt.Errorf("api token should exist")
			}
		} else {
			if len(resp.JSON200.Tokens) != 0 {
				return fmt.Errorf("api token should not exist")
			}
		}

		return nil
	}

}
