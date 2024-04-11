package provider_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	astronomerprovider "github.com/astronomer/astronomer-terraform-provider/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lucsky/cuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_Provider(t *testing.T) {
	t.Run("errors if missing token", func(t *testing.T) {
		ctx := context.Background()
		p := astronomerprovider.New("test")()
		resp := provider.ConfigureResponse{}
		req := provider.ConfigureRequest{
			Config: tfsdk.Config{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"token":           tftypes.String,
						"organization_id": tftypes.String,
						"host":            tftypes.String,
					},
				}, map[string]tftypes.Value{
					"organization_id": tftypes.NewValue(tftypes.String, cuid.New()),
					"host":            tftypes.NewValue(tftypes.String, "https://api.astronomer.io"),
					"token":           tftypes.NewValue(tftypes.String, ""),
				}),
				Schema: astronomerprovider.ProviderSchema(),
			},
		}
		p.Configure(ctx, req, &resp)
		assert.True(t, resp.Diagnostics.HasError())
		assert.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Missing Astro API Token")
	})
}

func TestAcc_Provider_config(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      invalidHost(),
				ExpectError: regexp.MustCompile(`.*Attribute host must be a valid Astronomer API host.*`),
			},
			{
				Config:      missingOrganizationIdConfig(),
				ExpectError: regexp.MustCompile(`.*The argument "organization_id" is required*`),
			},
			{
				Config:      organizationIdIsNotCuidConfig(),
				ExpectError: regexp.MustCompile(`.*Attribute organization_id value must be a cuid.*`),
			},
			{
				Config: explicitHostConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.astronomer_organization.test", "id"),
				),
			},
		},
	})
}

func explicitHostConfig() string {
	return fmt.Sprintf(`
provider "astronomer" {
organization_id = "%v"
host = "%v"
}`, os.Getenv("ASTRO_ORGANIZATION_ID"), os.Getenv("ASTRO_API_HOST")) + dataSourceConfig()
}

func missingOrganizationIdConfig() string {
	return `
provider "astronomer" {
}` + dataSourceConfig()
}

func organizationIdIsNotCuidConfig() string {
	return `
provider "astronomer" {
organization_id = "not-a-cuid"
}` + dataSourceConfig()
}

func invalidHost() string {
	return fmt.Sprintf(`
provider "astronomer" {
organization_id = "%v"
host = "https://api.astronomer.com"
}`, os.Getenv("ASTRO_ORGANIZATION_ID")) + dataSourceConfig()
}

// dataSourceConfig is needed to actually run the "Configure" method in the provider
func dataSourceConfig() string {
	return `
data "astronomer_organization" "test" {}`
}
