package datasources_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/utils"

	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAcc_DataSourceClusterOptions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      astronomerprovider.ProviderConfig(t, true) + clusterOptions("invalid", "AWS"),
				ExpectError: regexp.MustCompile(`type value must be one of`),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, true) + clusterOptions("HYBRID", "AWS"),
				Check: resource.ComposeTestCheckFunc(
					checkClusterOptions("AWS"),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, true) + clusterOptionsWithoutProviderFilter("HYBRID"),
				Check: resource.ComposeTestCheckFunc(
					checkClusterOptionsWithoutProviderFilter(),
				),
			},
		},
	})
}

func clusterOptions(clusterType, provider string) string {
	return fmt.Sprintf(`
data astro_cluster_options "test_data_cluster_options" {
 type = "%v"
 cloud_provider = "%v"
}`, clusterType, provider)
}

func clusterOptionsWithoutProviderFilter(clusterType string) string {
	return fmt.Sprintf(`
data astro_cluster_options "test_data_cluster_options" {
 type = "%v"
}`, clusterType)
}

func checkClusterOptions(provider string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		instanceState, numClusterOptions, err := utils.GetDataSourcesLength(s, "test_data_cluster_options", "cluster_options")
		if err != nil {
			return err
		}
		if numClusterOptions == 0 {
			return fmt.Errorf("expected clusterOptions to be greater or equal to 1, got %s", instanceState.Attributes["cluster_options.#"])
		}

		clusterOptionIdx := -1
		for i := 0; i < numClusterOptions; i++ {
			idxProvider := fmt.Sprintf("cluster_options.%d.provider", i)
			if instanceState.Attributes[idxProvider] == provider {
				clusterOptionIdx = i
				break
			}
		}
		if clusterOptionIdx == -1 {
			return fmt.Errorf("cluster option with provider %s not found", provider)
		}
		databaseInstance1 := fmt.Sprintf("cluster_options.%d.database_instances.0", clusterOptionIdx)
		resource.TestCheckResourceAttrSet(databaseInstance1, "cpu")
		resource.TestCheckResourceAttrSet(databaseInstance1, "memory")
		resource.TestCheckResourceAttrSet(databaseInstance1, "name")

		defaultDatabaseInstance := fmt.Sprintf("cluster_options.%d.default_database_instance", clusterOptionIdx)
		resource.TestCheckResourceAttrSet(defaultDatabaseInstance, "cpu")
		resource.TestCheckResourceAttrSet(defaultDatabaseInstance, "memory")
		resource.TestCheckResourceAttrSet(defaultDatabaseInstance, "name")

		nodeInstance1 := fmt.Sprintf("cluster_options.%d.node_instances.0", clusterOptionIdx)
		resource.TestCheckResourceAttrSet(nodeInstance1, "cpu")
		resource.TestCheckResourceAttrSet(nodeInstance1, "memory")
		resource.TestCheckResourceAttrSet(nodeInstance1, "name")

		defaultNodeInstance := fmt.Sprintf("cluster_options.%d.default_node_instance", clusterOptionIdx)
		resource.TestCheckResourceAttrSet(defaultNodeInstance, "cpu")
		resource.TestCheckResourceAttrSet(defaultNodeInstance, "memory")
		resource.TestCheckResourceAttrSet(defaultNodeInstance, "name")

		region1 := fmt.Sprintf("cluster_options.%d.regions.0", clusterOptionIdx)
		resource.TestCheckResourceAttrSet(region1, "name")

		defaultRegion := fmt.Sprintf("cluster_options.%d.default_region", clusterOptionIdx)
		resource.TestCheckResourceAttrSet(defaultRegion, "name")

		clusterOption1 := fmt.Sprintf("cluster_options.%d", clusterOptionIdx)
		resource.TestCheckResourceAttrSet(clusterOption1, "node_count_min")
		resource.TestCheckResourceAttrSet(clusterOption1, "node_count_max")
		resource.TestCheckResourceAttrSet(clusterOption1, "node_count_default")
		resource.TestCheckResourceAttrSet(clusterOption1, "default_vpc_subnet_range")
		resource.TestCheckResourceAttrSet(clusterOption1, "default_pod_subnet_range")
		resource.TestCheckResourceAttrSet(clusterOption1, "default_service_subnet_range")
		resource.TestCheckResourceAttrSet(clusterOption1, "default_service_peering_range")

		return nil
	}
}

func checkClusterOptionsWithoutProviderFilter() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		instanceState, numClusterOptions, err := utils.GetDataSourcesLength(s, "test_data_cluster_options", "cluster_options")
		if err != nil {
			return err
		}
		if numClusterOptions <= 1 {
			return fmt.Errorf("expected clusterOptions to be greater or equal to 1, got %s", instanceState.Attributes["cluster_options.#"])
		}
		var providers []string
		for i := 0; i < numClusterOptions; i++ {
			idxProvider := fmt.Sprintf("cluster_options.%d.provider", i)
			providers = append(providers, instanceState.Attributes[idxProvider])
		}
		if len(providers) == 0 {
			return fmt.Errorf("expected providers to be greater than 0")
		}

		for _, provider := range providers {
			checkClusterOptions(provider)
		}

		return nil
	}
}
