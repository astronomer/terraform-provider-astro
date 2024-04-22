package datasources_test

import (
	"fmt"
	"testing"

	"github.com/astronomer/astronomer-terraform-provider/internal/utils"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	astronomerprovider "github.com/astronomer/astronomer-terraform-provider/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_DataSourceClustersHybrid(t *testing.T) {
	tfVarName := "test_data_clusters_hybrid"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Check the data source for clusters for a hybrid organization
			{
				Config: astronomerprovider.ProviderConfig(t, false) + clusters(tfVarName),
				Check: resource.ComposeTestCheckFunc(
					checkClusters(tfVarName),
				),
			},
		},
	})
}

func clusters(tfVarName string) string {
	return fmt.Sprintf(`
data astronomer_clusters "%v" {}`, tfVarName)
}

func checkClusters(tfVarName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		instanceState, numClusters, err := utils.GetDataSourcesLength(s, tfVarName, "clusters")
		if err != nil {
			return err
		}
		if numClusters == 0 {
			return fmt.Errorf("expected clusters to be greater or equal to 1, got %s", instanceState.Attributes["clusters.#"])
		}

		// Check the first cluster
		clustersIdx := 0

		id := fmt.Sprintf("clusters.%d.id", clustersIdx)
		if instanceState.Attributes[id] == "" {
			return fmt.Errorf("expected 'id' to be set")
		}
		name := fmt.Sprintf("clusters.%d.name", clustersIdx)
		if instanceState.Attributes[name] == "" {
			return fmt.Errorf("expected 'name' to be set")
		}
		cloudProvider := fmt.Sprintf("clusters.%d.cloud_provider", clustersIdx)
		if instanceState.Attributes[cloudProvider] == "" {
			return fmt.Errorf("expected 'cloud_provider' to be set")
		}
		dbInstanceType := fmt.Sprintf("clusters.%d.db_instance_type", clustersIdx)
		if instanceState.Attributes[dbInstanceType] == "" {
			return fmt.Errorf("expected 'db_instance_type' to be set")
		}
		region := fmt.Sprintf("clusters.%d.region", clustersIdx)
		if instanceState.Attributes[region] == "" {
			return fmt.Errorf("expected 'region' to be set")
		}
		vpcSubnetRange := fmt.Sprintf("clusters.%d.vpc_subnet_range", clustersIdx)
		if instanceState.Attributes[vpcSubnetRange] == "" {
			return fmt.Errorf("expected 'vpc_subnet_range' to be set")
		}
		createdAt := fmt.Sprintf("clusters.%d.created_at", clustersIdx)
		if instanceState.Attributes[createdAt] == "" {
			return fmt.Errorf("expected 'created_at' to be set")
		}
		updatedAt := fmt.Sprintf("clusters.%d.updated_at", clustersIdx)
		if instanceState.Attributes[updatedAt] == "" {
			return fmt.Errorf("expected 'updated_at' to be set")
		}
		typ := fmt.Sprintf("clusters.%d.type", clustersIdx)
		if instanceState.Attributes[typ] != string(platform.ClusterTypeHYBRID) {
			return fmt.Errorf("expected 'type' to be set")
		}
		providerAccount := fmt.Sprintf("clusters.%d.provider_account", clustersIdx)
		if instanceState.Attributes[providerAccount] == "" {
			return fmt.Errorf("expected 'provider_account' to be set")
		}
		nodePoolsId := fmt.Sprintf("clusters.%d.node_pools.0.id", clustersIdx)
		if instanceState.Attributes[nodePoolsId] == "" {
			return fmt.Errorf("expected 'node_pools.0.id' to be set")
		}
		nodePoolsName := fmt.Sprintf("clusters.%d.node_pools.0.name", clustersIdx)
		if instanceState.Attributes[nodePoolsName] == "" {
			return fmt.Errorf("expected 'node_pools.0.name' to be set")
		}
		metadataExternalIps := fmt.Sprintf("clusters.%d.metadata.external_ips.0", clustersIdx)
		if instanceState.Attributes[metadataExternalIps] == "" {
			return fmt.Errorf("expected 'metadata.external_ips.0' to be set")
		}

		return nil
	}
}
