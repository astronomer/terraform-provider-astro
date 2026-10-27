package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	platform_v1 "github.com/astronomer/terraform-provider-astro/internal/clients/platform_v1"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
)

func newMetricsExportModel(exporterType string) models.EnvironmentObject {
	return models.EnvironmentObject{
		ObjectKey:     types.StringValue("dd_export"),
		ObjectType:    types.StringValue(string(platform_v1.CreateEnvironmentObjectRequestObjectTypeMETRICSEXPORT)),
		Scope:         types.StringValue(string(platform_v1.CreateEnvironmentObjectRequestScopeWORKSPACE)),
		ScopeEntityId: types.StringValue("clx42sxw501gl01o0gjenthnh"),
		ExporterType:  types.StringValue(exporterType),
		Endpoint:      types.StringNull(),
		AuthType:      types.StringNull(),
		BasicToken:    types.StringNull(),
		Username:      types.StringNull(),
		Password:      types.StringNull(),
		Headers:       types.MapNull(types.StringType),
		Labels:        types.MapNull(types.StringType),
		DataDogSite:   types.StringNull(),
		DataDogApiKey: types.StringNull(),
		Value:         types.StringNull(),
		IsSecret:      types.BoolNull(),
		Type:          types.StringNull(),
		Host:          types.StringNull(),
		Port:          types.Int64Null(),
		Schema:        types.StringNull(),
		Login:         types.StringNull(),
		Extra:         types.StringNull(),
		AuthTypeId:    types.StringNull(),
		Links:         types.SetNull(types.ObjectType{AttrTypes: schemas.EnvironmentObjectLinkAttributeTypes()}),
		ExcludeLinks:  types.SetNull(types.ObjectType{AttrTypes: schemas.EnvironmentObjectExcludeLinkAttributeTypes()}),
	}
}

func newOverrides(set map[string]attr.Value) types.Object {
	values := map[string]attr.Value{}
	for name, typ := range schemas.EnvironmentObjectOverridesAttributeTypes() {
		switch typ {
		case types.Int64Type:
			values[name] = types.Int64Null()
		case types.StringType:
			values[name] = types.StringNull()
		default:
			values[name] = types.MapNull(types.StringType)
		}
	}
	for k, v := range set {
		values[k] = v
	}
	return types.ObjectValueMust(schemas.EnvironmentObjectOverridesAttributeTypes(), values)
}

func newLinks(t *testing.T, overrides types.Object) types.Set {
	t.Helper()
	linkType := types.ObjectType{AttrTypes: schemas.EnvironmentObjectLinkAttributeTypes()}
	link := types.ObjectValueMust(schemas.EnvironmentObjectLinkAttributeTypes(), map[string]attr.Value{
		"scope":           types.StringValue("DEPLOYMENT"),
		"scope_entity_id": types.StringValue("clx44jyu001m201m5dzsbexqr"),
		"overrides":       overrides,
	})
	links, diags := types.SetValue(linkType, []attr.Value{link})
	require.False(t, diags.HasError(), "diags: %v", diags)
	return links
}

func newDatadogModel() models.EnvironmentObject {
	data := newMetricsExportModel("DATADOG")
	data.DataDogSite = types.StringValue("datadoghq.com")
	data.DataDogApiKey = types.StringValue("dd_api_key")
	return data
}

func TestUnit_ValidateMetricsExportFields_Datadog(t *testing.T) {
	t.Run("valid datadog export", func(t *testing.T) {
		data := newDatadogModel()
		data.Labels = types.MapValueMust(types.StringType, map[string]attr.Value{"team": types.StringValue("data")})
		assert.False(t, validateMetricsExportFields(&data).HasError())
	})

	t.Run("site and api key are required", func(t *testing.T) {
		data := newMetricsExportModel("DATADOG")
		diags := validateMetricsExportFields(&data)
		require.Equal(t, 2, diags.ErrorsCount(), "diags: %v", diags)
		assert.Contains(t, diags[0].Detail(), "data_dog_site is required")
		assert.Contains(t, diags[1].Detail(), "data_dog_api_key is required")
	})

	t.Run("unknown site and api key are not flagged as missing", func(t *testing.T) {
		data := newMetricsExportModel("DATADOG")
		data.DataDogSite = types.StringUnknown()
		data.DataDogApiKey = types.StringUnknown()
		assert.False(t, validateMetricsExportFields(&data).HasError())
	})

	t.Run("prometheus-only fields are rejected", func(t *testing.T) {
		data := newDatadogModel()
		data.Endpoint = types.StringValue("https://prometheus.example.com/api/v1/write")
		data.AuthType = types.StringValue("BASIC")
		data.BasicToken = types.StringValue("token")
		data.Username = types.StringValue("user")
		data.Password = types.StringValue("pass")
		data.Headers = types.MapValueMust(types.StringType, map[string]attr.Value{"X-Org": types.StringValue("1")})
		diags := validateMetricsExportFields(&data)
		assert.Equal(t, 6, diags.ErrorsCount(), "diags: %v", diags)
		for _, d := range diags {
			assert.Contains(t, d.Detail(), "not valid when exporter_type=DATADOG")
		}
	})
}

func TestUnit_ValidateMetricsExportFields_Prometheus(t *testing.T) {
	t.Run("endpoint is required", func(t *testing.T) {
		data := newMetricsExportModel("PROMETHEUS")
		diags := validateMetricsExportFields(&data)
		require.Equal(t, 1, diags.ErrorsCount(), "diags: %v", diags)
		assert.Contains(t, diags[0].Detail(), "endpoint is required when exporter_type=PROMETHEUS")
	})

	t.Run("datadog fields are rejected", func(t *testing.T) {
		data := newMetricsExportModel("PROMETHEUS")
		data.Endpoint = types.StringValue("https://prometheus.example.com/api/v1/write")
		data.DataDogSite = types.StringValue("datadoghq.com")
		data.DataDogApiKey = types.StringValue("dd_api_key")
		diags := validateMetricsExportFields(&data)
		assert.Equal(t, 2, diags.ErrorsCount(), "diags: %v", diags)
		for _, d := range diags {
			assert.Contains(t, d.Detail(), "only valid when exporter_type=DATADOG")
		}
	})
}

func TestUnit_MetricsExportOnlyFields_IncludesDatadog(t *testing.T) {
	data := newMetricsExportModel("DATADOG")
	data.ExporterType = types.StringNull()
	data.DataDogSite = types.StringValue("datadoghq.com")
	data.DataDogApiKey = types.StringValue("dd_api_key")
	names := lo.Map(metricsExportOnlyFields(&data), func(f namedField, _ int) string { return f.name })
	assert.ElementsMatch(t, []string{"data_dog_site", "data_dog_api_key"}, names)
}

func TestUnit_BuildRequests_Datadog(t *testing.T) {
	ctx := context.Background()

	t.Run("create sends site and api key", func(t *testing.T) {
		data := newDatadogModel()
		req, diags := buildCreateRequest(ctx, &data)
		require.False(t, diags.HasError(), "diags: %v", diags)
		require.NotNil(t, req.MetricsExport)
		assert.Equal(t, platform_v1.CreateEnvironmentObjectMetricsExportRequestExporterTypeDATADOG, req.MetricsExport.ExporterType)
		assert.Equal(t, lo.ToPtr(platform_v1.CreateEnvironmentObjectMetricsExportRequestDataDogSiteDatadoghqCom), req.MetricsExport.DataDogSite)
		assert.Equal(t, lo.ToPtr("dd_api_key"), req.MetricsExport.DataDogApiKey)
		assert.Nil(t, req.MetricsExport.Endpoint)
		assert.Nil(t, req.MetricsExport.AuthType)
	})

	t.Run("update sends site and api key", func(t *testing.T) {
		data := newDatadogModel()
		data.DataDogSite = types.StringValue("us3.datadoghq.com")
		req, diags := buildUpdateRequest(ctx, &data, nil)
		require.False(t, diags.HasError(), "diags: %v", diags)
		require.NotNil(t, req.MetricsExport)
		assert.Equal(t, lo.ToPtr(platform_v1.UpdateEnvironmentObjectMetricsExportRequestExporterTypeDATADOG), req.MetricsExport.ExporterType)
		assert.Equal(t, lo.ToPtr(platform_v1.UpdateEnvironmentObjectMetricsExportRequestDataDogSiteUs3DatadoghqCom), req.MetricsExport.DataDogSite)
		assert.Equal(t, lo.ToPtr("dd_api_key"), req.MetricsExport.DataDogApiKey)
	})

	t.Run("prometheus create omits datadog fields", func(t *testing.T) {
		data := newMetricsExportModel("PROMETHEUS")
		data.Endpoint = types.StringValue("https://prometheus.example.com/api/v1/write")
		req, diags := buildCreateRequest(ctx, &data)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Nil(t, req.MetricsExport.DataDogSite)
		assert.Nil(t, req.MetricsExport.DataDogApiKey)
	})

	t.Run("link overrides carry site and api key", func(t *testing.T) {
		overrides := newOverrides(map[string]attr.Value{
			"data_dog_site":    types.StringValue("datadoghq.eu"),
			"data_dog_api_key": types.StringValue("override_dd_key"),
		})
		objectType := platform_v1.CreateEnvironmentObjectRequestObjectTypeMETRICSEXPORT

		createOvr, diags := buildCreateOverrides(ctx, overrides, objectType)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Equal(t, lo.ToPtr(platform_v1.CreateEnvironmentObjectMetricsExportOverridesRequestDataDogSiteDatadoghqEu), createOvr.MetricsExport.DataDogSite)
		assert.Equal(t, lo.ToPtr("override_dd_key"), createOvr.MetricsExport.DataDogApiKey)

		updateOvr, diags := buildUpdateOverrides(ctx, overrides, newOverrides(nil), objectType)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Equal(t, lo.ToPtr(platform_v1.UpdateEnvironmentObjectMetricsExportOverridesRequestDataDogSiteDatadoghqEu), updateOvr.MetricsExport.DataDogSite)
		assert.Equal(t, lo.ToPtr("override_dd_key"), updateOvr.MetricsExport.DataDogApiKey)
		assert.Nil(t, updateOvr.UnsetFields)

		lop, diags := extractLinkOverridePreserve(ctx, overrides)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Equal(t, lo.ToPtr("override_dd_key"), lop.DataDogApiKey)
	})

	t.Run("preserve keeps the api key", func(t *testing.T) {
		data := newDatadogModel()
		preserve, diags := buildPreserveFromModel(ctx, &data)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Equal(t, lo.ToPtr("dd_api_key"), preserve.DataDogApiKey)
	})
}

func TestUnit_BuildUpdateOverrides_UnsetsRemovedDataDogApiKey(t *testing.T) {
	ctx := context.Background()
	metricsExport := platform_v1.CreateEnvironmentObjectRequestObjectTypeMETRICSEXPORT
	priorWithKey := newOverrides(map[string]attr.Value{
		"data_dog_site":    types.StringValue("datadoghq.eu"),
		"data_dog_api_key": types.StringValue("old_key"),
	})
	noOverrides := types.ObjectNull(schemas.EnvironmentObjectOverridesAttributeTypes())

	t.Run("key removed from overrides is unset", func(t *testing.T) {
		plan := newOverrides(map[string]attr.Value{"labels": types.MapValueMust(types.StringType, map[string]attr.Value{"env": types.StringValue("stage")})})
		out, diags := buildUpdateOverrides(ctx, plan, priorWithKey, metricsExport)
		require.False(t, diags.HasError(), "diags: %v", diags)
		require.NotNil(t, out.UnsetFields)
		assert.Equal(t, []string{"dataDogApiKey"}, *out.UnsetFields)
		assert.Nil(t, out.MetricsExport.DataDogApiKey)
	})

	t.Run("whole overrides block removed still unsets the key", func(t *testing.T) {
		out, diags := buildUpdateOverrides(ctx, noOverrides, priorWithKey, metricsExport)
		require.False(t, diags.HasError(), "diags: %v", diags)
		require.NotNil(t, out)
		require.NotNil(t, out.MetricsExport)
		assert.Equal(t, []string{"dataDogApiKey"}, *out.UnsetFields)
	})

	t.Run("kept key is not unset", func(t *testing.T) {
		plan := newOverrides(map[string]attr.Value{"data_dog_api_key": types.StringValue("new_key")})
		out, diags := buildUpdateOverrides(ctx, plan, priorWithKey, metricsExport)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Nil(t, out.UnsetFields)
		assert.Equal(t, lo.ToPtr("new_key"), out.MetricsExport.DataDogApiKey)
	})

	t.Run("no prior key leaves requests unchanged", func(t *testing.T) {
		out, diags := buildUpdateOverrides(ctx, noOverrides, noOverrides, metricsExport)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Nil(t, out)

		plan := newOverrides(map[string]attr.Value{"endpoint": types.StringValue("https://prom.example.com")})
		out, diags = buildUpdateOverrides(ctx, plan, newOverrides(nil), metricsExport)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Nil(t, out.UnsetFields)
	})

	t.Run("update request matches prior link by scope and entity", func(t *testing.T) {
		data := newDatadogModel()
		data.Links = newLinks(t, newOverrides(map[string]attr.Value{"data_dog_site": types.StringValue("us5.datadoghq.com")}))
		prior := newDatadogModel()
		prior.Links = newLinks(t, priorWithKey)

		req, diags := buildUpdateRequest(ctx, &data, &prior)
		require.False(t, diags.HasError(), "diags: %v", diags)
		require.NotNil(t, req.Links)
		require.Len(t, *req.Links, 1)
		link := (*req.Links)[0]
		require.NotNil(t, link.Overrides)
		assert.Equal(t, []string{"dataDogApiKey"}, *link.Overrides.UnsetFields)
	})
}
