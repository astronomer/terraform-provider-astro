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
	if token := os.Getenv("ASTRO_API_TOKEN"); len(token) == 0 {
		t.Fatal("ASTRO_API_TOKEN must be set for acceptance tests")
	}
	if orgId := os.Getenv("ASTRO_ORGANIZATION_ID"); len(orgId) == 0 {
		t.Fatal("ASTRO_ORGANIZATION_ID must be set for acceptance tests")
	}
	if host := os.Getenv("ASTRO_API_HOST"); len(host) == 0 {
		t.Fatal("ASTRO_API_HOST must be set for acceptance tests")
	}
}

func ProviderConfig() string {
	return fmt.Sprintf(`
provider "astronomer" {
	organization_id = "%v"
	host = "%v"
}`, os.Getenv("ASTRO_ORGANIZATION_ID"), os.Getenv("ASTRO_API_HOST"))
}
