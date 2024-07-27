package resources_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/lucsky/cuid"
	"github.com/stretchr/testify/assert"

	"os"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/utils"
)

func TestAcc_ResourceHybridClusterWorkspaceAuthorization(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)

	workspaceName := fmt.Sprintf("%v_workspace", namePrefix)
	workspaceResourceVar := fmt.Sprintf("astro_workspace.%v", workspaceName)

	clusterId := os.Getenv("HYBRID_DRY_RUN_CLUSTER_ID")
	clusterWorkspaceAuth := fmt.Sprintf("%v_auth", namePrefix)
	resourceVar := fmt.Sprintf("astro_hybrid_cluster_workspace_authorization.%v", clusterWorkspaceAuth)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			// Check that cluster workspace authorizations have been removed
			testAccCheckHybridClusterWorkspaceAuthorizationExistence(t, clusterWorkspaceAuth, false),
		),
		Steps: []resource.TestStep{
			// Test with workspace created through terraform
			{
				Config: astronomerprovider.ProviderConfig(t, false, false) +
					workspace(workspaceName, workspaceName, utils.TestResourceDescription, false) +
					hybridClusterWorkspaceAuthorization(hybridClusterWorkspaceAuthorizationInput{
						Name:         clusterWorkspaceAuth,
						ClusterId:    clusterId,
						WorkspaceIds: []string{fmt.Sprintf("%v.id", workspaceResourceVar)},
					}),
				Check: resource.ComposeTestCheckFunc(
					// Check hybrid cluster workspace authorization
					resource.TestCheckResourceAttr(resourceVar, "cluster_id", clusterId),
					resource.TestCheckResourceAttr(resourceVar, "workspace_ids.#", "1"),

					testAccCheckHybridClusterWorkspaceAuthorizationExistence(t, clusterWorkspaceAuth, true),
				),
			},
			// Import existing hybrid cluster workspace authorization
			{
				ResourceName:                         resourceVar,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        clusterId,
				ImportStateVerifyIdentifierAttribute: "cluster_id",
			},
			// Test with no workspaceIds
			{
				Config: astronomerprovider.ProviderConfig(t, false, false) +
					hybridClusterWorkspaceAuthorization(hybridClusterWorkspaceAuthorizationInput{
						Name:      clusterWorkspaceAuth,
						ClusterId: clusterId,
					}),
				Check: resource.ComposeTestCheckFunc(
					// Check hybrid cluster workspace authorization
					resource.TestCheckResourceAttr(resourceVar, "cluster_id", clusterId),
					resource.TestCheckNoResourceAttr(resourceVar, "workspace_ids"),

					testAccCheckHybridClusterWorkspaceAuthorizationExistence(t, clusterWorkspaceAuth, false),
				),
			},
		},
	})
}

type hybridClusterWorkspaceAuthorizationInput struct {
	Name         string
	ClusterId    string
	WorkspaceIds []string
}

func hybridClusterWorkspaceAuthorization(input hybridClusterWorkspaceAuthorizationInput) string {
	workspaceIds := lo.Map(input.WorkspaceIds, func(id string, _ int) string {
		if cuid.IsCuid(id) == nil {
			return fmt.Sprintf(`"%v"`, id)
		}
		return id
	})
	var workspaceIdsString string
	if len(workspaceIds) > 0 {
		workspaceIdsString = fmt.Sprintf("workspace_ids = [%v]", strings.Join(workspaceIds, ", "))
	}

	return fmt.Sprintf(`
		resource "astro_hybrid_cluster_workspace_authorization" "%s" {
			cluster_id = "%s"
			%v
		}`, input.Name, input.ClusterId, workspaceIdsString)
}

func testAccCheckHybridClusterWorkspaceAuthorizationExistence(t *testing.T, name string, shouldExist bool) func(state *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		client, err := utils.GetTestHybridPlatformClient()
		assert.NoError(t, err)

		organizationId := os.Getenv("HYBRID_ORGANIZATION_ID")
		clusterId := os.Getenv("HYBRID_DRY_RUN_CLUSTER_ID")

		ctx := context.Background()
		resp, err := client.GetClusterWithResponse(ctx, organizationId, clusterId)
		if err != nil {
			return fmt.Errorf("failed to get cluster: %w", err)
		}
		if resp == nil {
			return fmt.Errorf("response is nil")
		}
		if resp.JSON200 == nil {
			status, diag := clients.NormalizeAPIError(ctx, resp.HTTPResponse, resp.Body)
			return fmt.Errorf("response JSON200 is nil status: %v, err: %v", status, diag.Detail())
		}
		if shouldExist {
			if resp.JSON200.WorkspaceIds == nil || len(*resp.JSON200.WorkspaceIds) < 1 {
				return fmt.Errorf("cluster workspace authorization %s should exist", name)
			}
		} else {
			if resp.JSON200.WorkspaceIds != nil && len(*resp.JSON200.WorkspaceIds) != 0 {
				return fmt.Errorf("cluster workspace authorization %s should not exist", name)
			}
		}
		return nil
	}
}
