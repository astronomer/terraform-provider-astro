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

func TestAcc_DataSourceApiTokens(t *testing.T) {
	tfVarName := "test_data_api_tokens"
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
					checkApiTokens(tfVarName, false, false, false),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiTokensFilterWorkspaceId(tfVarName, tfWorkspaceId),
				Check: resource.ComposeTestCheckFunc(
					checkApiTokens(tfVarName, true, false, false),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiTokensFilterDeploymentId(tfVarName, tfDeploymentId),
				Check: resource.ComposeTestCheckFunc(
					checkApiTokens(tfVarName, false, true, false),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiTokensFilterOrgOnly(tfVarName, tfOrgOnly),
				Check: resource.ComposeTestCheckFunc(
					checkApiTokens(tfVarName, false, false, true),
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

func apiTokensFilterOrgOnly(tfVarName string, orgOnly bool) string {
	return fmt.Sprintf(`
data astro_api_tokens "%v" {
	include_only_organization_tokens = %v
}`, tfVarName, orgOnly)
}

func checkApiTokens(tfVarName string, filterWorkspaceId bool, filterDeploymentId bool, filterOrgOnly bool) resource.TestCheckFunc {
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
		description := fmt.Sprintf("api_tokens.%d.description", apiTokensIdx)
		if instanceState.Attributes[description] == "" {
			return fmt.Errorf("expected 'description' to be set")
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
		endAt := fmt.Sprintf("api_tokens.%d.end_at", apiTokensIdx)
		if instanceState.Attributes[endAt] == "" {
			return fmt.Errorf("expected 'end_at' to be set")
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
		expiryPeriodInDays := fmt.Sprintf("api_tokens.%d.expiry_period_in_days", apiTokensIdx)
		if instanceState.Attributes[expiryPeriodInDays] == "" {
			return fmt.Errorf("expected 'expiry_period_in_days' to be set")
		}
		lastUsedAt := fmt.Sprintf("api_tokens.%d.last_used_at", apiTokensIdx)
		if instanceState.Attributes[lastUsedAt] == "" {
			return fmt.Errorf("expected 'last_used_at' to be set")
		}
		entityId := fmt.Sprintf("api_tokens.%d.roles.0.entity_id", apiTokensIdx)
		if instanceState.Attributes[entityId] == "" {
			return fmt.Errorf("expected 'entity_id' to be set")
		}
		entityType := fmt.Sprintf("api_tokens.%d.roles.0.entity_type", apiTokensIdx)
		if instanceState.Attributes[entityType] == "" {
			return fmt.Errorf("expected 'entity_type' to be set")
		}
		role := fmt.Sprintf("api_tokens.%d.roles.0.role", apiTokensIdx)
		if instanceState.Attributes[role] == "" {
			return fmt.Errorf("expected 'roles' to be set")
		}
		token := fmt.Sprintf("api_tokens.%d.token", apiTokensIdx)
		if instanceState.Attributes[token] == "" {
			return fmt.Errorf("expected 'token' to be set")
		}

		return nil
	}
}