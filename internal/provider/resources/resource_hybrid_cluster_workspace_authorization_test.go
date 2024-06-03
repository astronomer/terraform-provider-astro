package resources_test

import (
	"context"
	"fmt"

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
	hybridWorkspaceId := os.Getenv("HYBRID_WORKSPACE_ID")

	clusterId := os.Getenv("HYBRID_CLUSTER_ID")
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
			// Test with two workspaces, an existing one and one created through terraform
			{
				Config: astronomerprovider.ProviderConfig(t, true) +
					workspace(workspaceName, workspaceName, utils.TestResourceDescription, false) +
					hybridClusterWorkspaceAuthorization(hybridClusterWorkspaceAuthorizationInput{
						Name:         clusterWorkspaceAuth,
						ClusterId:    clusterId,
						WorkspaceIds: []string{fmt.Sprintf("%v", hybridWorkspaceId), fmt.Sprintf("%v.id", workspaceResourceVar)},
					}),
				Check: resource.ComposeTestCheckFunc(
					// Check hybrid cluster workspace authorization
					resource.TestCheckResourceAttr(resourceVar, "cluster_id", clusterId),
					resource.TestCheckResourceAttr(resourceVar, "workspace_ids.#", "2"),

					testAccCheckHybridClusterWorkspaceAuthorizationExistence(t, clusterWorkspaceAuth, true),
				),
			},
			// Remove terraform created workspace from cluster workspace authorization
			{
				Config: astronomerprovider.ProviderConfig(t, true) +
					hybridClusterWorkspaceAuthorization(hybridClusterWorkspaceAuthorizationInput{
						Name:         clusterWorkspaceAuth,
						ClusterId:    clusterId,
						WorkspaceIds: []string{fmt.Sprintf("%v", hybridWorkspaceId)},
					}),
				Check: resource.ComposeTestCheckFunc(
					// Check hybrid cluster workspace authorization
					resource.TestCheckResourceAttr(resourceVar, "cluster_id", clusterId),
					resource.TestCheckResourceAttr(resourceVar, "workspace_ids.#", "1"),

					testAccCheckHybridClusterWorkspaceAuthorizationExistence(t, clusterWorkspaceAuth, true),
				),
			},
			// Test with no workspaceIds
			{
				Config: astronomerprovider.ProviderConfig(t, true) +
					hybridClusterWorkspaceAuthorization(hybridClusterWorkspaceAuthorizationInput{
						Name:         clusterWorkspaceAuth,
						ClusterId:    clusterId,
						WorkspaceIds: nil,
					}),
				Check: resource.ComposeTestCheckFunc(
					// Check hybrid cluster workspace authorization
					resource.TestCheckResourceAttr(resourceVar, "cluster_id", clusterId),
					resource.TestCheckResourceAttr(resourceVar, "workspace_ids.#", "0"),

					testAccCheckHybridClusterWorkspaceAuthorizationExistence(t, clusterWorkspaceAuth, true),
				),
			},
			// Import existing hybrid cluster workspace authorization
			{
				ResourceName:      resourceVar,
				ImportState:       true,
				ImportStateVerify: true,
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
	var workspaceIds []string

	for _, id := range input.WorkspaceIds {
		if cuid.IsCuid(id) == nil {
			workspaceIds = append(workspaceIds, fmt.Sprintf(`"%v"`, id))
		} else {
			workspaceIds = append(workspaceIds, id)
		}
	}

	return fmt.Sprintf(`
		resource "astro_hybrid_cluster_workspace_authorization" "%s" {
			cluster_id = "%s"
			workspace_ids = %v
		}`, input.Name, input.ClusterId, workspaceIds)
}

func testAccCheckHybridClusterWorkspaceAuthorizationExistence(t *testing.T, name string, shouldExist bool) func(state *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		client, err := utils.GetTestHostedPlatformClient()
		assert.NoError(t, err)

		organizationId := os.Getenv("HYBRID_ORGANIZATION_ID")
		clusterId := os.Getenv("HYBRID_CLUSTER_ID")

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
			if len(*resp.JSON200.WorkspaceIds) != 1 {
				return fmt.Errorf("cluster workspace authorization %s should exist", name)
			}
		} else {
			if len(*resp.JSON200.WorkspaceIds) != 0 {
				return fmt.Errorf("cluster workspace authorization %s should not exist", name)
			}
		}
		return nil
	}
}
