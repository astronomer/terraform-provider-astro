package datasources_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/terraform"

	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_DataSourceTeam(t *testing.T) {
	teamId := os.Getenv("HOSTED_TEAM_ID")
	teamName := "terraform_acceptance_tests_dnd"
	resourceVar := fmt.Sprintf("data.astro_team.%v", teamName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) + team(teamId, teamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttrSet(resourceVar, "name"),
					testCheckResourceAttrExists(resourcevar, "description", true),
					resource.TestCheckResourceAttrSet(resourceVar, "is_idp_managed"),
					resource.TestCheckResourceAttrSet(resourceVar, "organization_role"),
					resource.TestCheckResourceAttrSet(resourceVar, "workspace_roles"),
					resource.TestCheckResourceAttrSet(resourceVar, "deployment_roles"),
					resource.TestCheckResourceAttrSet(resourceVar, "roles_count"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by"),
				),
			},
		},
	})
}

func team(teamId string, teamName string) string {
	return fmt.Sprintf(`
data astro_team "%v" {
	id = "%v"
}`, teamName, teamId)
}

func testCheckResourceAttrExists(name, key string, canBeEmpty bool) resource.TestCheckFunc {
	return checkIfIndexesIntoTypeSet(key, func(s *terraform.State) error {
		is, err := primaryInstanceState(s, name)
		if err != nil {
			return err
		}

		return testCheckResourceAttrSet(is, name, key, canBeEmpty)
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

func testCheckResourceAttrSet(is *terraform.InstanceState, name string, key string, canBeEmpty bool) error {
	val, ok := is.Attributes[key]

	if canBeEmpty {
		if ok {
			return nil
		}
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
