package datasources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"

	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

type checkApiTokensInput struct {
	filterWorkspaceId  bool
	filterDeploymentId bool
	filterOrgOnly      bool
	workspaceId        string
	deploymentId       string
	organizationId     string
}

func TestAcc_DataSourceApiTokens(t *testing.T) {
	tfVarName := "test_data_api_tokens"
	tfOrganizationId := os.Getenv("HOSTED_ORGANIZATION_ID")
	tfWorkspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	tfDeploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
	tfOrgOnly := true

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiTokens(tfVarName),
				Check: resource.ComposeTestCheckFunc(
					checkApiTokens(tfVarName, checkApiTokensInput{
						filterWorkspaceId:  false,
						filterDeploymentId: false,
						filterOrgOnly:      false,
						workspaceId:        "",
						deploymentId:       "",
						organizationId:     tfOrganizationId,
					}),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiTokensFilterWorkspaceId(tfVarName, tfWorkspaceId),
				Check: resource.ComposeTestCheckFunc(
					checkApiTokens(tfVarName, checkApiTokensInput{
						filterWorkspaceId:  true,
						filterDeploymentId: false,
						filterOrgOnly:      false,
						workspaceId:        tfWorkspaceId,
						deploymentId:       "",
						organizationId:     tfOrganizationId,
					}),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiTokensFilterDeploymentId(tfVarName, tfDeploymentId),
				Check: resource.ComposeTestCheckFunc(
					checkApiTokens(tfVarName, checkApiTokensInput{
						filterWorkspaceId:  false,
						filterDeploymentId: true,
						filterOrgOnly:      false,
						workspaceId:        "",
						deploymentId:       tfDeploymentId,
						organizationId:     tfOrganizationId,
					}),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiTokensFilterOrgOnly(tfVarName),
				Check: resource.ComposeTestCheckFunc(
					checkApiTokens(tfVarName, checkApiTokensInput{
						filterWorkspaceId:  false,
						filterDeploymentId: false,
						filterOrgOnly:      true,
						workspaceId:        "",
						deploymentId:       "",
						organizationId:     tfOrganizationId,
					}),
				),
			},
		},
	})
}

func apiTokens(tfVarName string) string {
	return fmt.Sprintf(`
data astro_api_tokens "%v" {}`, tfVarName)
}

func apiTokensFilterWorkspaceId(tfVarName string, workspaceId string) string {
	return fmt.Sprintf(`
data astro_api_tokens "%v" {
	workspace_id = "%v"
}`, tfVarName, workspaceId)
}

func apiTokensFilterDeploymentId(tfVarName string, deploymentId string) string {
	return fmt.Sprintf(`
data astro_api_tokens "%v" {
	deployment_id = "%v"
}`, tfVarName, deploymentId)
}

func apiTokensFilterOrgOnly(tfVarName string) string {
	return fmt.Sprintf(`
data astro_api_tokens "%v" {
	include_only_organization_tokens = true
}`, tfVarName, orgOnly)
}

func checkApiTokens(tfVarName string, input checkApiTokensInput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		instanceState, numApiTokens, err := utils.GetDataSourcesLength(s, tfVarName, "api_tokens")
		if err != nil {
			return err
		}
		if numApiTokens == 0 {
			return fmt.Errorf("expected api_tokens to be greater or equal to 1, got %s", instanceState.Attributes["api_tokens.#"])
		}

		// Check the first api_token
		apiTokensIdx := 0

		id := fmt.Sprintf("api_tokens.%d.id", apiTokensIdx)
		if instanceState.Attributes[id] == "" {
			return fmt.Errorf("expected 'id' to be set")
		}
		name := fmt.Sprintf("api_tokens.%d.name", apiTokensIdx)
		if instanceState.Attributes[name] == "" {
			return fmt.Errorf("expected 'name' to be set")
		}
		shortToken := fmt.Sprintf("api_tokens.%d.short_token", apiTokensIdx)
		if instanceState.Attributes[shortToken] == "" {
			return fmt.Errorf("expected 'short_token' to be set")
		}
		tokenType := fmt.Sprintf("api_tokens.%d.type", apiTokensIdx)
		if instanceState.Attributes[tokenType] == "" {
			return fmt.Errorf("expected 'type' to be set")
		}
		startAt := fmt.Sprintf("api_tokens.%d.start_at", apiTokensIdx)
		if instanceState.Attributes[startAt] == "" {
			return fmt.Errorf("expected 'start_at' to be set")
		}
		createdAt := fmt.Sprintf("api_tokens.%d.created_at", apiTokensIdx)
		if instanceState.Attributes[createdAt] == "" {
			return fmt.Errorf("expected 'created_at' to be set")
		}
		updatedAt := fmt.Sprintf("api_tokens.%d.updated_at", apiTokensIdx)
		if instanceState.Attributes[updatedAt] == "" {
			return fmt.Errorf("expected 'updated_at' to be set")
		}
		createdBy := fmt.Sprintf("api_tokens.%d.created_by.id", apiTokensIdx)
		if instanceState.Attributes[createdBy] == "" {
			return fmt.Errorf("expected 'created_by.id' to be set")
		}
		updatedBy := fmt.Sprintf("api_tokens.%d.updated_by.id", apiTokensIdx)
		if instanceState.Attributes[updatedBy] == "" {
			return fmt.Errorf("expected 'updated_by.id' to be set")
		}
		entityIdKey := fmt.Sprintf("api_tokens.%d.roles.0.entity_id", apiTokensIdx)
		entityId := instanceState.Attributes[entityIdKey]
		entityTypeKey := fmt.Sprintf("api_tokens.%d.roles.0.entity_type", apiTokensIdx)
		entityType := instanceState.Attributes[entityTypeKey]
		role := fmt.Sprintf("api_tokens.%d.roles.0.role", apiTokensIdx)
		if input.filterWorkspaceId {
			if entityType != string(iam.ApiTokenRoleEntityTypeWORKSPACE) {
				return fmt.Errorf("expected 'entity_type' to be set to 'workspace'")
			}
			if entityId != input.workspaceId {
				return fmt.Errorf("expected 'entity_id' to be set to workspace_id")
			}
			if utils.CheckRole(role, "workspace") {
				return fmt.Errorf("expected 'role' to be set as a workspace role")
			}
		}

		if input.filterDeploymentId {
			if entityType != string(iam.ApiTokenRoleEntityTypeDEPLOYMENT) {
				return fmt.Errorf("expected 'entity_type' to be set to 'deployment'")
			}
			if entityId != input.deploymentId {
				return fmt.Errorf("expected 'entity_id' to be set to deployment_id")
			}
		}

		if input.filterOrgOnly {
			if entityType != string(iam.ApiTokenRoleEntityTypeORGANIZATION) {
				return fmt.Errorf("expected 'entity_type' to be set to 'organization'")
			}
			if entityId != input.organizationId {
				return fmt.Errorf("expected 'entity_id' to be set to organization_id")
			}
			if utils.CheckRole(role, "organization") {
				return fmt.Errorf("expected 'role' to be set as an organization role")
			}
		}

		return nil
	}
}
