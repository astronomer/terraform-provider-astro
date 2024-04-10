package utils_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/astronomer/astronomer-terraform-provider/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestUnit_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("DataSourceApiClientConfigureError", func(t *testing.T) {
		req := datasource.ConfigureRequest{
			ProviderData: nil,
		}
		resp := datasource.ConfigureResponse{}
		utils.DataSourceApiClientConfigureError(ctx, req, &resp)

		assert.True(t, resp.Diagnostics.HasError())
		assert.Contains(t, resp.Diagnostics[0].Detail(), "Expected apiClientsModel, got:")
	})

	t.Run("ResourceApiClientConfigureError", func(t *testing.T) {
		req := resource.ConfigureRequest{
			ProviderData: nil,
		}
		resp := resource.ConfigureResponse{}
		utils.ResourceApiClientConfigureError(ctx, req, &resp)

		assert.True(t, resp.Diagnostics.HasError())
		assert.Contains(t, resp.Diagnostics[0].Detail(), "Expected apiClientsModel, got:")
	})
}
