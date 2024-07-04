package provider

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// TestAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"astro": providerserver.NewProtocol6WithError(New("test")()),
}

func TestAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
	var missingEnvVars []string
	envVars := []string{
		"HOSTED_ORGANIZATION_API_TOKEN",
		"HOSTED_ORGANIZATION_ID",
		"HYBRID_ORGANIZATION_API_TOKEN",
		"HYBRID_ORGANIZATION_ID",
		"HYBRID_DRY_RUN_CLUSTER_ID",
		"ASTRO_API_HOST",
		"HYBRID_CLUSTER_ID",
		"HYBRID_NODE_POOL_ID",
		"HOSTED_TEAM_ID",
		"HOSTED_USER_ID",
	}
	for _, envVar := range envVars {
		if val := os.Getenv(envVar); len(val) == 0 {
			missingEnvVars = append(missingEnvVars, envVar)
		}
	}
	if len(missingEnvVars) > 0 {
		t.Fatalf("Pre-check failed: %+v must be set for acceptance tests", strings.Join(missingEnvVars, ", "))
	}
}

func ProviderConfig(t *testing.T, isHosted bool) string {
	var orgId, token string
	if isHosted {
		orgId = os.Getenv("HOSTED_ORGANIZATION_ID")
		token = os.Getenv("HOSTED_ORGANIZATION_API_TOKEN")
	} else {
		orgId = os.Getenv("HYBRID_ORGANIZATION_ID")
		token = os.Getenv("HYBRID_ORGANIZATION_API_TOKEN")
	}

	return fmt.Sprintf(`
provider "astro" {
	organization_id = "%v"
	host = "%v"
	token = "%v"
}`, orgId, os.Getenv("ASTRO_API_HOST"), token)
}
