package provider_test

import (
	"context"

	astronomerprovider "github.com/astronomer/astronomer-terraform-provider/internal/provider"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/models"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("Provider Test", func() {
	providerName := "astronomer"
	providerVersion := "test"
	var ctx context.Context
	var p provider.Provider
	BeforeEach(func() {
		ctx = context.Background()
		p = astronomerprovider.New(providerVersion)()
	})
	It("has expected resources", func() {
		expectedResources := []string{
			"astronomer_workspace",
		}

		resources := p.Resources(ctx)
		Expect(resources).To(HaveLen(len(lo.Uniq(expectedResources))))
		for _, resourceFn := range resources {
			res := resourceFn()
			req := resource.MetadataRequest{ProviderTypeName: providerName}
			resp := resource.MetadataResponse{}
			res.Metadata(ctx, req, &resp)
			Expect(expectedResources).To(ContainElement(resp.TypeName))
		}
	})

	It("has expected data sources", func() {
		expectedDataSources := []string{
			"astronomer_workspace",
			"astronomer_workspaces",
			"astronomer_deployment",
			"astronomer_deployments",
		}

		dataSources := p.DataSources(ctx)
		Expect(dataSources).To(HaveLen(len(lo.Uniq(expectedDataSources))))
		for _, datasourceFn := range dataSources {
			res := datasourceFn()
			req := datasource.MetadataRequest{ProviderTypeName: providerName}
			resp := datasource.MetadataResponse{}
			res.Metadata(ctx, req, &resp)
			Expect(expectedDataSources).To(ContainElement(resp.TypeName))
		}
	})

	It("schema validation", func() {
		type params struct {
			name        string
			isOptional  bool
			isSensitive bool
		}
		expectedAttributes := []params{
			{"token", true, true},
			{"organization_id", false, false},
			{"host", true, false},
		}
		req := provider.SchemaRequest{}
		resp := provider.SchemaResponse{}
		p.Schema(ctx, req, &resp)
		schemaAttributes := resp.Schema.Attributes
		Expect(schemaAttributes).To(HaveLen(len(expectedAttributes)))
		for _, attr := range expectedAttributes {
			Expect(schemaAttributes).To(HaveKey(attr.name))
			Expect(schemaAttributes[attr.name].IsOptional()).To(Equal(attr.isOptional))
			Expect(schemaAttributes[attr.name].IsRequired()).To(Equal(!attr.isOptional))
			Expect(schemaAttributes[attr.name].IsSensitive()).To(Equal(attr.isSensitive))
			Expect(schemaAttributes[attr.name].IsComputed()).To(BeFalse())
			Expect(schemaAttributes[attr.name].GetMarkdownDescription()).ToNot(BeEmpty())
		}
	})

	Context("configure", func() {
		var req provider.ConfigureRequest
		var resp provider.ConfigureResponse

		BeforeEach(func() {
			resp = provider.ConfigureResponse{}
		})

		It("errors if missing token", func() {
			req = provider.ConfigureRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"token":           tftypes.String,
							"organization_id": tftypes.String,
							"host":            tftypes.String,
						},
					}, map[string]tftypes.Value{
						"organization_id": tftypes.NewValue(tftypes.String, "sampleOrganizationId"),
						"host":            tftypes.NewValue(tftypes.String, "sampleHost"),
						"token":           tftypes.NewValue(tftypes.String, ""),
					}),
					Schema: astronomerprovider.ProviderSchema(),
				},
			}
			p.Configure(ctx, req, &resp)
			Expect(resp.Diagnostics.HasError()).To(BeTrue())
			Expect(
				resp.Diagnostics.Errors()[0].Summary(),
			).To(ContainSubstring("Missing Astro API Token"))
		})

		It("configures correctly", func() {
			req = provider.ConfigureRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"token":           tftypes.String,
							"organization_id": tftypes.String,
							"host":            tftypes.String,
						},
					}, map[string]tftypes.Value{
						"organization_id": tftypes.NewValue(tftypes.String, "sampleOrganizationId"),
						"host":            tftypes.NewValue(tftypes.String, "sampleHost"),
						"token":           tftypes.NewValue(tftypes.String, "sampleToken"),
					}),
					Schema: astronomerprovider.ProviderSchema(),
				},
			}
			p.Configure(ctx, req, &resp)
			Expect(resp.Diagnostics.HasError()).To(BeFalse())
			dataSourceData := resp.DataSourceData.(models.ApiClientsModel)
			Expect(dataSourceData.OrganizationId).To(Equal("sampleOrganizationId"))
			Expect(dataSourceData.PlatformClient).ToNot(BeNil())
			Expect(dataSourceData.IamClient).ToNot(BeNil())
			resourceData := resp.ResourceData.(models.ApiClientsModel)
			Expect(resourceData.OrganizationId).To(Equal("sampleOrganizationId"))
			Expect(resourceData.PlatformClient).ToNot(BeNil())
			Expect(resourceData.IamClient).ToNot(BeNil())
		})
	})
})
