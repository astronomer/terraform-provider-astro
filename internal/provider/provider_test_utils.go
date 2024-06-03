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
	if hostedToken := os.Getenv("HOSTED_ORGANIZATION_API_TOKEN"); len(hostedToken) == 0 {
		missingEnvVars = append(missingEnvVars, "HOSTED_ORGANIZATION_API_TOKEN")
	}
	if hostedOrgId := os.Getenv("HOSTED_ORGANIZATION_ID"); len(hostedOrgId) == 0 {
		missingEnvVars = append(missingEnvVars, "HOSTED_ORGANIZATION_ID")
	}
	if hybridToken := os.Getenv("HYBRID_ORGANIZATION_API_TOKEN"); len(hybridToken) == 0 {
		missingEnvVars = append(missingEnvVars, "HYBRID_ORGANIZATION_API_TOKEN")
	}
	if hybridOrgId := os.Getenv("HYBRID_ORGANIZATION_ID"); len(hybridOrgId) == 0 {
		missingEnvVars = append(missingEnvVars, "HYBRID_ORGANIZATION_ID")
	}
	if host := os.Getenv("ASTRO_API_HOST"); len(host) == 0 {
		missingEnvVars = append(missingEnvVars, "ASTRO_API_HOST")
	}
	if hybridClusterId := os.Getenv("HYBRID_CLUSTER_ID"); len(hybridClusterId) == 0 {
		missingEnvVars = append(missingEnvVars, "HYBRID_CLUSTER_ID")
	}
	if hybridNodePoolId := os.Getenv("HYBRID_NODE_POOL_ID"); len(hybridNodePoolId) == 0 {
		missingEnvVars = append(missingEnvVars, "HYBRID_NODE_POOL_ID")
	}
	if hostedTeamId := os.Getenv("HOSTED_TEAM_ID"); len(hostedTeamId) == 0 {
		missingEnvVars = append(missingEnvVars, "HOSTED_TEAM_ID")
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
