package resources

import (
	"context"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform_v1"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_buildUpdateLinks(t *testing.T) {
	ctx := context.Background()
	linkObjType := types.ObjectType{AttrTypes: schemas.EnvironmentObjectLinkAttributeTypes()}
	metricsExport := func(links types.Set) *models.EnvironmentObject {
		return &models.EnvironmentObject{
			ObjectType: types.StringValue(string(platform_v1.CreateEnvironmentObjectRequestObjectTypeMETRICSEXPORT)),
			Links:      links,
		}
	}

	t.Run("omitted links are left out so the API keeps the existing links", func(t *testing.T) {
		for _, links := range []types.Set{types.SetUnknown(linkObjType), types.SetNull(linkObjType)} {
			var req platform_v1.UpdateEnvironmentObjectRequest
			diags := buildUpdateLinks(ctx, metricsExport(links), &req)
			require.False(t, diags.HasError(), diags)
			assert.Nil(t, req.Links)
		}
	})

	t.Run("an empty links set is sent so the API removes every link", func(t *testing.T) {
		var req platform_v1.UpdateEnvironmentObjectRequest
		diags := buildUpdateLinks(ctx, metricsExport(types.SetValueMust(linkObjType, []attr.Value{})), &req)
		require.False(t, diags.HasError(), diags)
		require.NotNil(t, req.Links)
		assert.Empty(t, *req.Links)
	})

	t.Run("configured links are sent with their label overrides", func(t *testing.T) {
		link, diags := models.EnvironmentObjectLinkTypesObject(ctx, &platform_v1.EnvironmentObjectLink{
			Scope:         platform_v1.EnvironmentObjectLinkScopeDEPLOYMENT,
			ScopeEntityId: "deployment-1",
			MetricsExportOverrides: &platform_v1.EnvironmentObjectMetricsExportOverrides{
				Labels: &map[string]string{"env": "prod"},
			},
		}, nil)
		require.False(t, diags.HasError(), diags)

		var req platform_v1.UpdateEnvironmentObjectRequest
		diags = buildUpdateLinks(ctx, metricsExport(types.SetValueMust(linkObjType, []attr.Value{link})), &req)
		require.False(t, diags.HasError(), diags)
		require.NotNil(t, req.Links)
		require.Len(t, *req.Links, 1)
		sent := (*req.Links)[0]
		assert.Equal(t, "deployment-1", sent.ScopeEntityId)
		require.NotNil(t, sent.Overrides)
		require.NotNil(t, sent.Overrides.MetricsExport)
		assert.Equal(t, &map[string]string{"env": "prod"}, sent.Overrides.MetricsExport.Labels)
	})
}
