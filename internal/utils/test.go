package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

var platformClient *platform.ClientWithResponses

const TestResourceDescription = "Created by Terraform Acceptance Test - will self-cleanup"

func GenerateTestResourceName(numRandomChars int) string {
	return fmt.Sprintf("TFAcceptanceTest_%v", strings.ToUpper(acctest.RandStringFromCharSet(numRandomChars, acctest.CharSetAlpha)))
}

func GetTestPlatformClient() (*platform.ClientWithResponses, error) {
	if platformClient != nil {
		return platformClient, nil
	}
	return platform.NewPlatformClient(os.Getenv("ASTRO_API_HOST"), os.Getenv("ASTRO_API_TOKEN"), "acceptancetests")
}

// GetDataSourcesLength retrieves the number of elements returned from a data source in the Terraform state.
// For example, if the config is `data.astronomer_workspaces.my_workspaces`, the `dataSourceName` would be `workspaces` and
// `tfVarName` would `my_workspaces`.
// The returned value is the instance state, the number of elements in `workspaces` of that data source, and an error if there is one.
func GetDataSourcesLength(state *terraform.State, tfVarName, dataSourceName string) (*terraform.InstanceState, int, error) {
	resourceID := fmt.Sprintf("data.astronomer_%s.%s", dataSourceName, tfVarName)

	// Retrieve the resource state by its identifier.
	resourceState := state.Modules[0].Resources[resourceID]
	if resourceState == nil {
		return nil, 0, fmt.Errorf("resource not found in state for data source '%s'", resourceID)
	}

	// Retrieve the primary instance of the resource.
	instanceState := resourceState.Primary
	if instanceState == nil {
		return nil, 0, fmt.Errorf("resource '%s' has no primary instance", resourceID)
	}

	// Retrieve the size of the data sources from the state.
	numDataSources := fmt.Sprintf("%s.#", dataSourceName)

	// Convert the attribute to an integer.
	numAttribute, err := strconv.Atoi(instanceState.Attributes[numDataSources])
	if err != nil {
		return nil, 0, fmt.Errorf("expected a number for field '%s', got '%s'", dataSourceName, instanceState.Attributes[numDataSources])
	}

	return instanceState, numAttribute, nil
}
