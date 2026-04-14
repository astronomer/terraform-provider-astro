package resources_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAcc_ResourceAgentToken(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	deploymentId := os.Getenv("REMOTE_EXECUTION_DEPLOYMENT_ID")

	tokenName := fmt.Sprintf("%v_agent", namePrefix)
	resourceVar := fmt.Sprintf("astro_agent_token.%v", tokenName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckAgentTokenExistence(t, deploymentId, tokenName, false),
		),
		Steps: []resource.TestStep{
			// Create with all fields
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + agentToken(agentTokenInput{
					Name:               tokenName,
					Description:        utils.TestResourceDescription,
					DeploymentId:       deploymentId,
					ExpiryPeriodInDays: 30,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", tokenName),
					resource.TestCheckResourceAttr(resourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(resourceVar, "deployment_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "expiry_period_in_days", "30"),
					resource.TestCheckResourceAttrSet(resourceVar, "token"),
					testAccCheckAgentTokenExistence(t, deploymentId, tokenName, true),
				),
			},
			// Import existing agent token and check it is correctly imported
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateIdFunc:       testAccAgentTokenImportStateIdFunc(resourceVar),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func TestAcc_ResourceAgentTokenNoExpiry(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	deploymentId := os.Getenv("REMOTE_EXECUTION_DEPLOYMENT_ID")

	tokenName := fmt.Sprintf("%v_agent_no_expiry", namePrefix)
	resourceVar := fmt.Sprintf("astro_agent_token.%v", tokenName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckAgentTokenExistence(t, deploymentId, tokenName, false),
		),
		Steps: []resource.TestStep{
			// Create without expiry
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + agentToken(agentTokenInput{
					Name:         tokenName,
					DeploymentId: deploymentId,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", tokenName),
					resource.TestCheckNoResourceAttr(resourceVar, "expiry_period_in_days"),
					resource.TestCheckNoResourceAttr(resourceVar, "description"),
					resource.TestCheckResourceAttrSet(resourceVar, "token"),
					testAccCheckAgentTokenExistence(t, deploymentId, tokenName, true),
				),
			},
		},
	})
}

type agentTokenInput struct {
	Name               string
	Description        string
	DeploymentId       string
	ExpiryPeriodInDays int
}

func agentToken(input agentTokenInput) string {
	var description string
	if input.Description != "" {
		description = fmt.Sprintf(`description = "%v"`, input.Description)
	}

	var expiry string
	if input.ExpiryPeriodInDays > 0 {
		expiry = fmt.Sprintf("expiry_period_in_days = %v", input.ExpiryPeriodInDays)
	}

	return fmt.Sprintf(`
resource astro_agent_token "%v" {
	name          = "%v"
	deployment_id = "%v"
	%v
	%v
}`, input.Name, input.Name, input.DeploymentId, description, expiry)
}

func testAccAgentTokenImportStateIdFunc(resourceVar string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceVar]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceVar)
		}
		deploymentId := rs.Primary.Attributes["deployment_id"]
		id := rs.Primary.Attributes["id"]
		return fmt.Sprintf("%s/%s", deploymentId, id), nil
	}
}

func testAccCheckAgentTokenExistence(t *testing.T, deploymentId string, name string, shouldExist bool) func(s *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		client, err := utils.GetTestIamClient(true)
		assert.NoError(t, err)

		organizationId := os.Getenv("HOSTED_ORGANIZATION_ID")
		ctx := context.Background()

		resp, err := client.ListAgentTokensWithResponse(ctx, organizationId, deploymentId, nil)
		if err != nil {
			return fmt.Errorf("failed to list agent tokens: %v", err)
		}
		if resp == nil {
			return fmt.Errorf("nil response from list agent tokens")
		}
		if resp.JSON200 == nil {
			status, diag := clients.NormalizeAPIError(ctx, resp.HTTPResponse, resp.Body)
			return fmt.Errorf("response JSON200 is nil, status: %v, err: %v", status, diag.Detail())
		}

		for _, token := range resp.JSON200.Tokens {
			if token.Name == name {
				if shouldExist {
					return nil
				}
				return fmt.Errorf("agent token %q should not exist but does", name)
			}
		}

		if shouldExist {
			return fmt.Errorf("agent token %q should exist but was not found", name)
		}
		return nil
	}
}
