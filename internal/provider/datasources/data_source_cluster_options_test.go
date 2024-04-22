package datasources_test

import (
	"fmt"
	"testing"

	"github.com/astronomer/astronomer-terraform-provider/internal/utils"

	astronomerprovider "github.com/astronomer/astronomer-terraform-provider/internal/provider"
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
				Config: astronomerprovider.ProviderConfig(t, true) + clusterOptions("HYBRID", "AWS"),
				Check: resource.ComposeTestCheckFunc(
					checkClusterOptions(),
				),
			},
		},
	})
}

func clusterOptions(clusterType, provider string) string {
	return fmt.Sprintf(`
data astronomer_cluster_options "test_data_cluster_options" {
  type = "%v"
  cloud_provider = "%v"
}`, clusterType, provider)
}

func clusterOptionsWithoutProviderFilter(clusterType, provider string) string {
	return fmt.Sprintf(`
data astronomer_cluster_options "test_data_cluster_options" {
  type = "%v"
}`, clusterType)
}

func checkClusterOptions() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		instanceState, numClusterOptions, err := utils.GetDataSourcesLength(s, "test_data_cluster_options", "cluster_options")
		if err != nil {
			return err
		}
		if numClusterOptions == 0 {
			return fmt.Errorf("expected clusterOptions to be greater or equal to 1, got %s", instanceState.Attributes["cluster_options.#"])
		}
		fmt.Println("AHAAHHA")
		fmt.Println(instanceState)
		return nil
	}
}
