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

type Role struct {
	Role     string
	EntityId string
}

// ContainsWorkspaceRole checks if a workspace role is in the list of workspace roles
func ContainsWorkspaceRole(workspaceRoles []iam.WorkspaceRole, role Role) bool {
	for _, r := range workspaceRoles {
		if r.WorkspaceId == role.EntityId && string(r.Role) == role.Role {
			return true
		}
	}
	return false
}

// ContainsWorkspaceRoles checks if a list of workspace roles contains a list of roles
func ContainsWorkspaceRoles(userRoles []iam.WorkspaceRole, roles []Role) []Role {
	var missingRoles []Role
	for _, role := range roles {
		if !ContainsWorkspaceRole(userRoles, role) {
			missingRoles = append(missingRoles, role)
		}
	}
	return missingRoles
}

// ContainsDeploymentRole checks if a deployment role is in the list of deployment roles
func ContainsDeploymentRole(roles []iam.DeploymentRole, role Role) bool {
	for _, r := range roles {
		if r.DeploymentId == role.EntityId && r.Role == role.Role {
			return true
		}
	}
	return false
}

// ContainsDeploymentRoles checks if a list of deployment roles contains a list of roles
func ContainsDeploymentRoles(userRoles []iam.DeploymentRole, roles []Role) []Role {
	var missingRoles []Role
	for _, role := range roles {
		if !ContainsDeploymentRole(userRoles, role) {
			missingRoles = append(missingRoles, role)
		}
	}
	return missingRoles
}
