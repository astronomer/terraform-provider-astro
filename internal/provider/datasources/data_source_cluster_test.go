package datasources_test

import (
	"fmt"
	"os"
	"testing"

	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_DataSourceCluster(t *testing.T) {
	hybridClusterId := os.Getenv("HYBRID_CLUSTER_ID")
	resourceName := "test_data_cluster_hybrid"
	resourceVar := fmt.Sprintf("data.astro_cluster.%v", resourceName)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Check the data source for cluster for a hybrid organization
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HYBRID) + cluster(resourceName, hybridClusterId),
				Check: resource.ComposeTestCheckFunc(
					// These checks are for the cluster data source (singular)
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttrSet(resourceVar, "name"),
					resource.TestCheckResourceAttrSet(resourceVar, "cloud_provider"),
					resource.TestCheckResourceAttrSet(resourceVar, "db_instance_type"),
					resource.TestCheckResourceAttrSet(resourceVar, "region"),
					resource.TestCheckResourceAttrSet(resourceVar, "vpc_subnet_range"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttr(resourceVar, "type", "HYBRID"),
					resource.TestCheckResourceAttrSet(resourceVar, "provider_account"),
					resource.TestCheckResourceAttrSet(resourceVar, "node_pools.0.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "node_pools.0.name"),
					resource.TestCheckResourceAttrSet(resourceVar, "metadata.external_ips.0"),
				),
			},
		},
	})
}

func cluster(resourceName, clusterId string) string {
	return fmt.Sprintf(`
data astro_cluster "%v" {
	  id = "%v"
}`, resourceName, clusterId)
}
