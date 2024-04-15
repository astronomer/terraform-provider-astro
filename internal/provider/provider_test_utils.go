package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// TestAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"astronomer": providerserver.NewProtocol6WithError(New("test")()),
}

func TestAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
	if hostedToken := os.Getenv("HOSTED_ORGANIZATION_API_TOKEN"); len(hostedToken) == 0 {
		t.Fatal("HOSTED_ORGANIZATION_API_TOKEN must be set for acceptance tests")
	}
	if hostedOrgId := os.Getenv("HOSTED_ORGANIZATION_ID"); len(hostedOrgId) == 0 {
		t.Fatal("HOSTED_ORGANIZATION_ID must be set for acceptance tests")
	}
	if hybridToken := os.Getenv("HYBRID_ORGANIZATION_API_TOKEN"); len(hybridToken) == 0 {
		t.Fatal("HYBRID_ORGANIZATION_API_TOKEN must be set for acceptance tests")
	}
	if hybridOrgId := os.Getenv("HYBRID_ORGANIZATION_ID"); len(hybridOrgId) == 0 {
		t.Fatal("HYBRID_ORGANIZATION_ID must be set for acceptance tests")
	}
	if host := os.Getenv("ASTRO_API_HOST"); len(host) == 0 {
		t.Fatal("ASTRO_API_HOST must be set for acceptance tests")
	}
}

func ProviderConfig(t *testing.T, isHosted bool) string {
	var orgId string
	if isHosted {
		orgId = os.Getenv("HOSTED_ORGANIZATION_ID")
		t.Setenv("ASTRO_API_TOKEN", os.Getenv("HOSTED_ORGANIZATION_API_TOKEN"))
	} else {
		orgId = os.Getenv("HYBRID_ORGANIZATION_ID")
		t.Setenv("ASTRO_API_TOKEN", os.Getenv("HYBRID_ORGANIZATION_API_TOKEN"))
	}

	return fmt.Sprintf(`
provider "astronomer" {
	organization_id = "%v"
	host = "%v"
}`, orgId, os.Getenv("ASTRO_API_HOST"))
}
