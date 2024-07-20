package resources_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
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

// Acceptance tests for Organization, Workspace, and Deployment API tokens
// Within each test we create, update, change resource type and import the resource

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
			testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{name: apiTokenName, organization: true, shouldExist: false}),
		),
		Steps: []resource.TestStep{
			// Test invalid role for token type
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name: apiTokenName,
					Type: string(iam.ORGANIZATION),
					Roles: []apiTokenRole{
						{
							Role:       string(iam.WORKSPACEOWNER),
							EntityId:   workspaceId,
							EntityType: string(iam.WORKSPACE),
						},
					},
				}),
				ExpectError: regexp.MustCompile("Bad Request Error"),
			},
			// Test invalid role for entity type
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name: apiTokenName,
					Type: string(iam.ORGANIZATION),
					Roles: []apiTokenRole{
						{
							Role:       string(iam.WORKSPACEOWNER),
							EntityId:   workspaceId,
							EntityType: string(iam.ORGANIZATION),
						},
					},
				}),
				ExpectError: regexp.MustCompile("Bad Request Error"),
			},
			// Test multiple roles of the same type
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name: apiTokenName,
					Type: string(iam.ORGANIZATION),
					Roles: []apiTokenRole{
						{
							Role:       string(iam.ORGANIZATIONOWNER),
							EntityId:   workspaceId,
							EntityType: string(iam.ORGANIZATION),
						},
						{
							Role:       string(iam.ORGANIZATIONBILLINGADMIN),
							EntityId:   workspaceId,
							EntityType: string(iam.ORGANIZATION),
						},
					},
				}),
				ExpectError: regexp.MustCompile("Bad Request Error"),
			},
			// Create the organization api token
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: utils.TestResourceDescription,
					Type:        string(iam.ORGANIZATION),
					Roles: []apiTokenRole{
						{
							Role:       string(iam.ORGANIZATIONOWNER),
							EntityId:   organizationId,
							EntityType: string(iam.ORGANIZATION),
						},
						{
							Role:       string(iam.WORKSPACEOWNER),
							EntityId:   workspaceId,
							EntityType: string(iam.WORKSPACE),
						},
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: string(iam.DEPLOYMENT),
						},
					},
					ExpiryPeriodInDays: 30,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", apiTokenName),
					resource.TestCheckResourceAttr(resourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(resourceVar, "type", string(iam.ORGANIZATION)),
					resource.TestCheckResourceAttrSet(resourceVar, "short_token"),
					resource.TestCheckResourceAttrSet(resourceVar, "start_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					resource.TestCheckResourceAttr(resourceVar, "expiry_period_in_days", "30"),
					resource.TestCheckResourceAttr(resourceVar, "roles.#", "3"),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_id", organizationId),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_type", string(iam.ORGANIZATION)),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.role", string(iam.ORGANIZATIONOWNER)),
					resource.TestCheckResourceAttr(resourceVar, "roles.1.entity_id", workspaceId),
					resource.TestCheckResourceAttr(resourceVar, "roles.1.entity_type", string(iam.WORKSPACE)),
					resource.TestCheckResourceAttr(resourceVar, "roles.1.role", string(iam.WORKSPACEOWNER)),
					resource.TestCheckResourceAttr(resourceVar, "roles.2.entity_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "roles.2.entity_type", string(iam.DEPLOYMENT)),
					resource.TestCheckResourceAttr(resourceVar, "roles.2.role", "DEPLOYMENT_ADMIN"),
					// Check via API that organization api token exists
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{name: apiTokenName, organization: true, shouldExist: true}),
				),
			},
			// Change properties and check they have been updated in terraform state
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: "new description",
					Type:        string(iam.ORGANIZATION),
					Roles: []apiTokenRole{
						{
							Role:       string(iam.ORGANIZATIONOWNER),
							EntityId:   organizationId,
							EntityType: string(iam.ORGANIZATION),
						},
						{
							Role:       string(iam.WORKSPACEOWNER),
							EntityId:   workspaceId,
							EntityType: string(iam.WORKSPACE),
						},
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: string(iam.DEPLOYMENT),
						},
					},
					ExpiryPeriodInDays: 30,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "description", "new description"),
					// Check via API that organization api token exists
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{name: apiTokenName, organization: true, shouldExist: true}),
				),
			},
			// Change the resource type and remove roles and optional fields
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name: apiTokenName,
					Type: string(iam.WORKSPACE),
					Roles: []apiTokenRole{
						{
							Role:       string(iam.WORKSPACEOWNER),
							EntityId:   workspaceId,
							EntityType: string(iam.WORKSPACE),
						},
					},
					ExpiryPeriodInDays: 30,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "type", string(iam.WORKSPACE)),
					resource.TestCheckResourceAttr(resourceVar, "roles.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_id", workspaceId),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_type", string(iam.WORKSPACE)),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.role", string(iam.WORKSPACEOWNER)),
					// Check via API that api token was destroyed and recreated
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{name: apiTokenName, workspace: true, shouldExist: true}),
				),
			},
			// Change resource type back to ORGANIZATION
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: utils.TestResourceDescription,
					Type:        string(iam.ORGANIZATION),
					Roles: []apiTokenRole{
						{
							Role:       string(iam.ORGANIZATIONOWNER),
							EntityId:   organizationId,
							EntityType: string(iam.ORGANIZATION),
						},
						{
							Role:       string(iam.WORKSPACEOWNER),
							EntityId:   workspaceId,
							EntityType: string(iam.WORKSPACE),
						},
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: string(iam.DEPLOYMENT),
						},
					},
					ExpiryPeriodInDays: 30,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "type", string(iam.ORGANIZATION)),
					resource.TestCheckResourceAttr(resourceVar, "description", utils.TestResourceDescription),
					// Check via API that organization api token exists
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{name: apiTokenName, organization: true, shouldExist: true}),
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

	apiTokenName := fmt.Sprintf("%v_workspace", namePrefix)
	resourceVar := fmt.Sprintf("astro_api_token.%v", apiTokenName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			// Check that the organization api token has been removed
			testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{name: apiTokenName, workspace: true, shouldExist: false}),
		),
		Steps: []resource.TestStep{
			// Test invalid role for token type
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name: apiTokenName,
					Type: string(iam.WORKSPACE),
					Roles: []apiTokenRole{
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: string(iam.DEPLOYMENT),
						},
					},
				}),
				ExpectError: regexp.MustCompile("Bad Request Error"),
			},
			// Test invalid role for entity type
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name: apiTokenName,
					Type: string(iam.WORKSPACE),
					Roles: []apiTokenRole{
						{
							Role:       string(iam.ORGANIZATIONOWNER),
							EntityId:   workspaceId,
							EntityType: string(iam.WORKSPACE),
						},
					},
				}),
				ExpectError: regexp.MustCompile("Bad Request Error"),
			},
			// Test invalid role for API token type
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name: apiTokenName,
					Type: string(iam.WORKSPACE),
					Roles: []apiTokenRole{
						{
							Role:       string(iam.ORGANIZATIONOWNER),
							EntityId:   workspaceId,
							EntityType: string(iam.ORGANIZATION),
						},
					},
				}),
				ExpectError: regexp.MustCompile("Bad Request Error"),
			},
			// Test multiple roles of the same type
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name: apiTokenName,
					Type: string(iam.WORKSPACE),
					Roles: []apiTokenRole{
						{
							Role:       string(iam.WORKSPACEOWNER),
							EntityId:   workspaceId,
							EntityType: string(iam.WORKSPACE),
						},
						{
							Role:       string(iam.WORKSPACEOPERATOR),
							EntityId:   workspaceId,
							EntityType: string(iam.WORKSPACE),
						},
					},
					ExpiryPeriodInDays: 30,
				}),
				ExpectError: regexp.MustCompile("Bad Request Error"),
			},
			// Create the workspace api token
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: utils.TestResourceDescription,
					Type:        string(iam.WORKSPACE),
					Roles: []apiTokenRole{
						{
							Role:       string(iam.WORKSPACEOWNER),
							EntityId:   workspaceId,
							EntityType: string(iam.WORKSPACE),
						},
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: string(iam.DEPLOYMENT),
						},
					},
					ExpiryPeriodInDays: 30,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", apiTokenName),
					resource.TestCheckResourceAttr(resourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(resourceVar, "type", string(iam.WORKSPACE)),
					resource.TestCheckResourceAttrSet(resourceVar, "short_token"),
					resource.TestCheckResourceAttrSet(resourceVar, "start_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					resource.TestCheckResourceAttr(resourceVar, "expiry_period_in_days", "30"),
					resource.TestCheckResourceAttr(resourceVar, "roles.#", "2"),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_id", workspaceId),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_type", string(iam.WORKSPACE)),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.role", string(iam.WORKSPACEOWNER)),
					resource.TestCheckResourceAttr(resourceVar, "roles.1.entity_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "roles.1.entity_type", string(iam.DEPLOYMENT)),
					resource.TestCheckResourceAttr(resourceVar, "roles.1.role", "DEPLOYMENT_ADMIN"),
					// Check via API that organization api token exists
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{name: apiTokenName, workspace: true, shouldExist: true}),
				),
			},
			// Change properties and check they have been updated in terraform state
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: "new description",
					Type:        string(iam.WORKSPACE),
					Roles: []apiTokenRole{
						{
							Role:       string(iam.WORKSPACEOWNER),
							EntityId:   workspaceId,
							EntityType: string(iam.WORKSPACE),
						},
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: string(iam.DEPLOYMENT),
						},
					},
					ExpiryPeriodInDays: 30,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "description", "new description"),
					// Check via API that organization api token exists
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{name: apiTokenName, workspace: true, shouldExist: true}),
				),
			},
			// Change the resource type and remove roles and optional fields
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name: apiTokenName,
					Type: string(iam.ORGANIZATION),
					Roles: []apiTokenRole{
						{
							Role:       string(iam.ORGANIZATIONOWNER),
							EntityId:   organizationId,
							EntityType: string(iam.ORGANIZATION),
						},
					},
					ExpiryPeriodInDays: 30,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "type", string(iam.ORGANIZATION)),
					resource.TestCheckResourceAttr(resourceVar, "roles.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_id", organizationId),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_type", string(iam.ORGANIZATION)),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.role", string(iam.ORGANIZATIONOWNER)),
					// Check via API that api token was destroyed and recreated
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{name: apiTokenName, organization: true, shouldExist: true}),
				),
			},
			// Change resource type back to WORKSPACE
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: utils.TestResourceDescription,
					Type:        string(iam.WORKSPACE),
					Roles: []apiTokenRole{
						{
							Role:       string(iam.WORKSPACEOWNER),
							EntityId:   workspaceId,
							EntityType: string(iam.WORKSPACE),
						},
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: string(iam.DEPLOYMENT),
						},
					},
					ExpiryPeriodInDays: 30,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "type", string(iam.WORKSPACE)),
					resource.TestCheckResourceAttr(resourceVar, "description", utils.TestResourceDescription),
					// Check via API that organization api token exists
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{name: apiTokenName, workspace: true, shouldExist: true}),
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
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")

	apiTokenName := fmt.Sprintf("%v_deployment", namePrefix)
	resourceVar := fmt.Sprintf("astro_api_token.%v", apiTokenName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			// Check that the organization api token has been removed
			testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{name: apiTokenName, deployment: true, shouldExist: false}),
		),
		Steps: []resource.TestStep{
			// Test invalid role for token type
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name: apiTokenName,
					Type: string(iam.DEPLOYMENT),
					Roles: []apiTokenRole{
						{
							Role:       string(iam.WORKSPACEOWNER),
							EntityId:   workspaceId,
							EntityType: string(iam.WORKSPACE),
						},
					},
				}),
				ExpectError: regexp.MustCompile("Bad Request Error"),
			},
			// Test invalid role for entity type
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name: apiTokenName,
					Type: string(iam.DEPLOYMENT),
					Roles: []apiTokenRole{
						{
							Role:       string(iam.ORGANIZATIONOWNER),
							EntityId:   deploymentId,
							EntityType: string(iam.DEPLOYMENT),
						},
					},
				}),
				ExpectError: regexp.MustCompile("Bad Request Error"),
			},
			// Test invalid role for API token type
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name: apiTokenName,
					Type: string(iam.DEPLOYMENT),
					Roles: []apiTokenRole{
						{
							Role:       string(iam.ORGANIZATIONOWNER),
							EntityId:   organizationId,
							EntityType: string(iam.ORGANIZATION),
						},
					},
				}),
				ExpectError: regexp.MustCompile("Bad Request Error"),
			},
			// Test multiple roles of the same type
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name: apiTokenName,
					Type: string(iam.DEPLOYMENT),
					Roles: []apiTokenRole{
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: string(iam.DEPLOYMENT),
						},
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: string(iam.DEPLOYMENT),
						},
					},
				}),
				ExpectError: regexp.MustCompile("Bad Request Error"),
			},
			// Create the deployment api token
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: utils.TestResourceDescription,
					Type:        string(iam.DEPLOYMENT),
					Roles: []apiTokenRole{
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: string(iam.DEPLOYMENT),
						},
					},
					ExpiryPeriodInDays: 30,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", apiTokenName),
					resource.TestCheckResourceAttr(resourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(resourceVar, "type", string(iam.DEPLOYMENT)),
					resource.TestCheckResourceAttrSet(resourceVar, "short_token"),
					resource.TestCheckResourceAttrSet(resourceVar, "start_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					resource.TestCheckResourceAttr(resourceVar, "expiry_period_in_days", "30"),
					resource.TestCheckResourceAttr(resourceVar, "roles.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_type", string(iam.DEPLOYMENT)),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.role", "DEPLOYMENT_ADMIN"),
					// Check via API that organization api token exists
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{name: apiTokenName, deployment: true, shouldExist: true}),
				),
			},
			// Change properties and check they have been updated in terraform state
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: "new description",
					Type:        string(iam.DEPLOYMENT),
					Roles: []apiTokenRole{
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: string(iam.DEPLOYMENT),
						},
					},
					ExpiryPeriodInDays: 30,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "description", "new description"),
					// Check via API that organization api token exists
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{name: apiTokenName, deployment: true, shouldExist: true}),
				),
			},
			// Change the resource type
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: utils.TestResourceDescription,
					Type:        string(iam.ORGANIZATION),
					Roles: []apiTokenRole{
						{
							Role:       string(iam.ORGANIZATIONOWNER),
							EntityId:   organizationId,
							EntityType: string(iam.ORGANIZATION),
						},
					},
					ExpiryPeriodInDays: 30,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "type", string(iam.ORGANIZATION)),
					resource.TestCheckResourceAttr(resourceVar, "roles.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_id", organizationId),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.entity_type", string(iam.ORGANIZATION)),
					resource.TestCheckResourceAttr(resourceVar, "roles.0.role", string(iam.ORGANIZATIONOWNER)),
					// Check via API that api token was destroyed and recreated
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{name: apiTokenName, organization: true, shouldExist: true}),
				),
			},
			// Change resource type back to DEPLOYMENT
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenInput{
					Name:        apiTokenName,
					Description: utils.TestResourceDescription,
					Type:        string(iam.DEPLOYMENT),
					Roles: []apiTokenRole{
						{
							Role:       "DEPLOYMENT_ADMIN",
							EntityId:   deploymentId,
							EntityType: string(iam.DEPLOYMENT),
						},
					},
					ExpiryPeriodInDays: 30,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "type", string(iam.DEPLOYMENT)),
					resource.TestCheckResourceAttr(resourceVar, "description", utils.TestResourceDescription),
					// Check via API that organization api token exists
					testAccCheckApiTokenExistence(t, checkApiTokensExistenceInput{name: apiTokenName, deployment: true, shouldExist: true}),
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
	name         string
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

		// Check that the api token exists, multiple api tokens exist in the entity
		for _, token := range resp.JSON200.Tokens {
			if token.Name == input.name {
				if input.shouldExist {
					return nil
				} else {
					return fmt.Errorf("api token should not exist")
				}
			}
		}

		if input.shouldExist {
			return fmt.Errorf("api token should exist")
		}

		return nil
	}

}
