package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"

	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var hostedPlatformClient, hybridPlatformClient *platform.ClientWithResponses
var hostedIamClient, hybridIamClient *iam.ClientWithResponses

const TestResourceDescription = "Created by Terraform Acceptance Test - should self-cleanup but can delete manually if needed after 2 hours."

func GenerateTestResourceName(numRandomChars int) string {
	return fmt.Sprintf("TFAcceptanceTest_%v", strings.ToUpper(acctest.RandStringFromCharSet(numRandomChars, acctest.CharSetAlpha)))
}

func GetTestIamClient(isHosted bool) (*iam.ClientWithResponses, error) {
	if isHosted {
		return GetTestHostedIamClient()
	} else {
		return GetTestHybridIamClient()
	}
}

func GetTestHybridIamClient() (*iam.ClientWithResponses, error) {
	if hybridIamClient != nil {
		return hybridIamClient, nil
	}
	var err error
	hybridIamClient, err = iam.NewIamClient(os.Getenv("ASTRO_API_HOST"), os.Getenv("HYBRID_ORGANIZATION_API_TOKEN"), "acceptancetests")
	return hybridIamClient, err
}

func GetTestHostedIamClient() (*iam.ClientWithResponses, error) {
	if hostedIamClient != nil {
		return hostedIamClient, nil
	}
	var err error
	hostedIamClient, err = iam.NewIamClient(os.Getenv("ASTRO_API_HOST"), os.Getenv("HOSTED_ORGANIZATION_API_TOKEN"), "acceptancetests")
	return hostedIamClient, err
}

func GetTestPlatformClient(isHosted bool) (*platform.ClientWithResponses, error) {
	if isHosted {
		return GetTestHostedPlatformClient()
	} else {
		return GetTestHybridPlatformClient()
	}
}

func GetTestHybridPlatformClient() (*platform.ClientWithResponses, error) {
	if hybridPlatformClient != nil {
		return hybridPlatformClient, nil
	}
	var err error
	hybridPlatformClient, err = platform.NewPlatformClient(os.Getenv("ASTRO_API_HOST"), os.Getenv("HYBRID_ORGANIZATION_API_TOKEN"), "acceptancetests")
	return hybridPlatformClient, err
}

func GetTestHostedPlatformClient() (*platform.ClientWithResponses, error) {
	if hostedPlatformClient != nil {
		return hostedPlatformClient, nil
	}
	var err error
	hostedPlatformClient, err = platform.NewPlatformClient(os.Getenv("ASTRO_API_HOST"), os.Getenv("HOSTED_ORGANIZATION_API_TOKEN"), "acceptancetests")
	return hostedPlatformClient, err
}

// GetDataSourcesLength retrieves the number of elements returned from a data source in the Terraform state.
// For example, if the config is `data.astro_workspaces.my_workspaces`, the `dataSourceName` would be `workspaces` and
// `tfVarName` would `my_workspaces`.
// The returned value is the instance state, the number of elements in `workspaces` of that data source, and an error if there is one.
func GetDataSourcesLength(state *terraform.State, tfVarName, dataSourceName string) (*terraform.InstanceState, int, error) {
	resourceID := fmt.Sprintf("data.astro_%s.%s", dataSourceName, tfVarName)

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

func TestCheckResourceAttrExists(name, key string, isOptional bool) resource.TestCheckFunc {
	return checkIfIndexesIntoTypeSet(key, func(s *terraform.State) error {
		is, err := primaryInstanceState(s, name)
		if err != nil {
			return err
		}

		return testCheckResourceAttrSet(is, name, key, isOptional)
	})
}

func checkIfIndexesIntoTypeSet(key string, f resource.TestCheckFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		err := f(s)
		if err != nil && indexesIntoTypeSet(key) {
			return fmt.Errorf("Error in test check: %s\nTest check address %q likely indexes into TypeSet\nThis is currently not possible in the SDK", err, key)
		}
		return err
	}
}

// indexesIntoTypeSet is a heuristic to try and identify if a flatmap style
// string address uses a precalculated TypeSet hash, which are integers and
// typically are large and obviously not a list index
func indexesIntoTypeSet(key string) bool {
	for _, part := range strings.Split(key, ".") {
		if i, err := strconv.Atoi(part); err == nil && i > 100 {
			return true
		}
	}
	return false
}

// primaryInstanceState returns the primary instance state for the given
// resource name in the root module.
func primaryInstanceState(s *terraform.State, name string) (*terraform.InstanceState, error) {
	ms := s.RootModule() //nolint:staticcheck // legacy usage
	return modulePrimaryInstanceState(ms, name)
}

// modulePrimaryInstanceState returns the instance state for the given resource
// name in a ModuleState
func modulePrimaryInstanceState(ms *terraform.ModuleState, name string) (*terraform.InstanceState, error) {
	rs, ok := ms.Resources[name]
	if !ok {
		return nil, fmt.Errorf("Not found: %s in %s", name, ms.Path)
	}

	is := rs.Primary
	if is == nil {
		return nil, fmt.Errorf("No primary instance: %s in %s", name, ms.Path)
	}

	return is, nil
}

func testCheckResourceAttrSet(is *terraform.InstanceState, name string, key string, isOptional bool) error {
	val, ok := is.Attributes[key]

	if ok && isOptional {
		return nil
	}

	if ok && val != "" {
		return nil
	}

	if _, ok := is.Attributes[key+".#"]; ok {
		return fmt.Errorf(
			"%s: list or set attribute '%s' must be checked by element count key (%s) or element value keys (e.g. %s). Set element value checks should use TestCheckTypeSet functions instead.",
			name,
			key,
			key+".#",
			key+".0",
		)
	}

	if _, ok := is.Attributes[key+".%"]; ok {
		return fmt.Errorf(
			"%s: map attribute '%s' must be checked by element count key (%s) or element value keys (e.g. %s).",
			name,
			key,
			key+".%",
			key+".examplekey",
		)
	}

	return fmt.Errorf("%s: Attribute '%s' expected to be set", name, key)
}
