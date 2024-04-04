package utils_test

import (
	"context"

	"github.com/astronomer/astronomer-terraform-provider/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Errors Test", func() {
	var ctx context.Context
	BeforeEach(func() {
		ctx = context.Background()
	})

	It("should add error to diags if data source api client configure fails", func() {
		req := datasource.ConfigureRequest{
			ProviderData: nil,
		}
		resp := datasource.ConfigureResponse{}
		utils.DataSourceApiClientConfigureError(ctx, req, &resp)
		Expect(resp.Diagnostics.HasError()).To(BeTrue())
		Expect(resp.Diagnostics[0].Detail()).To(ContainSubstring("Expected apiClientsModel, got:"))
	})

	It("should add error to diags if data source api client configure fails", func() {
		req := resource.ConfigureRequest{
			ProviderData: nil,
		}
		resp := resource.ConfigureResponse{}
		utils.ResourceApiClientConfigureError(ctx, req, &resp)
		Expect(resp.Diagnostics.HasError()).To(BeTrue())
		Expect(resp.Diagnostics[0].Detail()).To(ContainSubstring("Expected apiClientsModel, got:"))
	})
})
