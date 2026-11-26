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

const testDynatraceEndpoint = "https://abc12345.live.dynatrace.com/api/v2/otlp/v1/metrics"

func newDynatraceModel() models.EnvironmentObject {
	data := newMetricsExportModel("DYNATRACE")
	data.Endpoint = types.StringValue(testDynatraceEndpoint)
	data.DynatraceApiToken = types.StringValue("dt0c01.token")
	return data
}

func TestUnit_ValidateMetricsExportFields_Dynatrace(t *testing.T) {
	t.Run("valid dynatrace export", func(t *testing.T) {
		data := newDynatraceModel()
		data.Labels = types.MapValueMust(types.StringType, map[string]attr.Value{"team": types.StringValue("data")})
		assert.False(t, validateMetricsExportFields(&data).HasError())
	})

	t.Run("endpoint and api token are required", func(t *testing.T) {
		data := newMetricsExportModel("DYNATRACE")
		diags := validateMetricsExportFields(&data)
		require.Equal(t, 2, diags.ErrorsCount(), "diags: %v", diags)
		assert.Contains(t, diags[0].Detail(), "endpoint is required when exporter_type=DYNATRACE")
		assert.Contains(t, diags[1].Detail(), "dynatrace_api_token is required")
	})

	t.Run("unknown endpoint and api token are not flagged", func(t *testing.T) {
		data := newMetricsExportModel("DYNATRACE")
		data.Endpoint = types.StringUnknown()
		data.DynatraceApiToken = types.StringUnknown()
		assert.False(t, validateMetricsExportFields(&data).HasError())
	})

	t.Run("endpoint must be an https URL", func(t *testing.T) {
		for _, endpoint := range []string{
			"http://abc12345.live.dynatrace.com/api/v2/otlp/v1/metrics",
			"abc12345.live.dynatrace.com/api/v2/otlp/v1/metrics",
			"https://",
		} {
			data := newDynatraceModel()
			data.Endpoint = types.StringValue(endpoint)
			diags := validateMetricsExportFields(&data)
			require.Equal(t, 1, diags.ErrorsCount(), "endpoint %q diags: %v", endpoint, diags)
			assert.Contains(t, diags[0].Detail(), "endpoint must be an https URL")
		}
	})

	t.Run("prometheus auth fields and headers are rejected", func(t *testing.T) {
		data := newDynatraceModel()
		data.AuthType = types.StringValue("BASIC")
		data.BasicToken = types.StringValue("token")
		data.Username = types.StringValue("user")
		data.Password = types.StringValue("pass")
		data.Headers = types.MapValueMust(types.StringType, map[string]attr.Value{"Authorization": types.StringValue("x")})
		diags := validateMetricsExportFields(&data)
		assert.Equal(t, 5, diags.ErrorsCount(), "diags: %v", diags)
		for _, d := range diags {
			assert.Contains(t, d.Detail(), "not valid when exporter_type=DYNATRACE")
		}
	})

	t.Run("datadog fields are rejected", func(t *testing.T) {
		data := newDynatraceModel()
		data.DataDogSite = types.StringValue("datadoghq.com")
		data.DataDogApiKey = types.StringValue("dd_api_key")
		diags := validateMetricsExportFields(&data)
		assert.Equal(t, 2, diags.ErrorsCount(), "diags: %v", diags)
		for _, d := range diags {
			assert.Contains(t, d.Detail(), "only valid when exporter_type=DATADOG")
		}
	})
}

func TestUnit_ValidateMetricsExportFields_DynatraceTokenOnOtherExporters(t *testing.T) {
	t.Run("prometheus", func(t *testing.T) {
		data := newMetricsExportModel("PROMETHEUS")
		data.Endpoint = types.StringValue("https://prometheus.example.com/api/v1/write")
		data.DynatraceApiToken = types.StringValue("dt0c01.token")
		diags := validateMetricsExportFields(&data)
		require.Equal(t, 1, diags.ErrorsCount(), "diags: %v", diags)
		assert.Contains(t, diags[0].Detail(), "dynatrace_api_token is only valid when exporter_type=DYNATRACE")
	})

	t.Run("datadog", func(t *testing.T) {
		data := newDatadogModel()
		data.DynatraceApiToken = types.StringValue("dt0c01.token")
		diags := validateMetricsExportFields(&data)
		require.Equal(t, 1, diags.ErrorsCount(), "diags: %v", diags)
		assert.Contains(t, diags[0].Detail(), "dynatrace_api_token is only valid when exporter_type=DYNATRACE")
	})

	t.Run("other object types", func(t *testing.T) {
		data := newMetricsExportModel("PROMETHEUS")
		data.ExporterType = types.StringNull()
		data.DynatraceApiToken = types.StringValue("dt0c01.token")
		names := lo.Map(metricsExportOnlyFields(&data), func(f namedField, _ int) string { return f.name })
		assert.ElementsMatch(t, []string{"dynatrace_api_token"}, names)
	})
}

func TestUnit_BuildRequests_Dynatrace(t *testing.T) {
	ctx := context.Background()

	t.Run("create sends endpoint and api token", func(t *testing.T) {
		data := newDynatraceModel()
		req, diags := buildCreateRequest(ctx, &data)
		require.False(t, diags.HasError(), "diags: %v", diags)
		require.NotNil(t, req.MetricsExport)
		assert.Equal(t, platform_v1.CreateEnvironmentObjectMetricsExportRequestExporterTypeDYNATRACE, req.MetricsExport.ExporterType)
		assert.Equal(t, lo.ToPtr(testDynatraceEndpoint), req.MetricsExport.Endpoint)
		assert.Equal(t, lo.ToPtr("dt0c01.token"), req.MetricsExport.DynatraceApiToken)
		assert.Nil(t, req.MetricsExport.AuthType)
		assert.Nil(t, req.MetricsExport.DataDogSite)
		assert.Nil(t, req.MetricsExport.DataDogApiKey)
	})

	t.Run("update sends endpoint and api token", func(t *testing.T) {
		data := newDynatraceModel()
		req, diags := buildUpdateRequest(ctx, &data, nil)
		require.False(t, diags.HasError(), "diags: %v", diags)
		require.NotNil(t, req.MetricsExport)
		assert.Equal(t, lo.ToPtr(platform_v1.UpdateEnvironmentObjectMetricsExportRequestExporterTypeDYNATRACE), req.MetricsExport.ExporterType)
		assert.Equal(t, lo.ToPtr(testDynatraceEndpoint), req.MetricsExport.Endpoint)
		assert.Equal(t, lo.ToPtr("dt0c01.token"), req.MetricsExport.DynatraceApiToken)
	})

	t.Run("prometheus create omits the api token", func(t *testing.T) {
		data := newMetricsExportModel("PROMETHEUS")
		data.Endpoint = types.StringValue("https://prometheus.example.com/api/v1/write")
		req, diags := buildCreateRequest(ctx, &data)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Nil(t, req.MetricsExport.DynatraceApiToken)
	})

	t.Run("link overrides carry the api token", func(t *testing.T) {
		overrides := newOverrides(map[string]attr.Value{
			"endpoint":            types.StringValue("https://xyz98765.live.dynatrace.com/api/v2/otlp/v1/metrics"),
			"dynatrace_api_token": types.StringValue("override_dt_token"),
		})
		objectType := platform_v1.CreateEnvironmentObjectRequestObjectTypeMETRICSEXPORT

		createOvr, diags := buildCreateOverrides(ctx, overrides, objectType)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Equal(t, lo.ToPtr("override_dt_token"), createOvr.MetricsExport.DynatraceApiToken)

		updateOvr, diags := buildUpdateOverrides(ctx, overrides, newOverrides(nil), objectType)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Equal(t, lo.ToPtr("override_dt_token"), updateOvr.MetricsExport.DynatraceApiToken)
		assert.Nil(t, updateOvr.UnsetFields)

		lop, diags := extractLinkOverridePreserve(ctx, overrides)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Equal(t, lo.ToPtr("override_dt_token"), lop.DynatraceApiToken)
	})

	t.Run("preserve keeps the api token", func(t *testing.T) {
		data := newDynatraceModel()
		preserve, diags := buildPreserveFromModel(ctx, &data)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Equal(t, lo.ToPtr("dt0c01.token"), preserve.DynatraceApiToken)
	})
}

func TestUnit_BuildUpdateOverrides_UnsetsRemovedDynatraceApiToken(t *testing.T) {
	ctx := context.Background()
	metricsExport := platform_v1.CreateEnvironmentObjectRequestObjectTypeMETRICSEXPORT
	priorWithToken := newOverrides(map[string]attr.Value{
		"dynatrace_api_token": types.StringValue("old_token"),
	})
	noOverrides := types.ObjectNull(schemas.EnvironmentObjectOverridesAttributeTypes())

	t.Run("token removed from overrides is unset", func(t *testing.T) {
		plan := newOverrides(map[string]attr.Value{"labels": types.MapValueMust(types.StringType, map[string]attr.Value{"env": types.StringValue("stage")})})
		out, diags := buildUpdateOverrides(ctx, plan, priorWithToken, metricsExport)
		require.False(t, diags.HasError(), "diags: %v", diags)
		require.NotNil(t, out.UnsetFields)
		assert.Equal(t, []string{"dynatraceApiToken"}, *out.UnsetFields)
		assert.Nil(t, out.MetricsExport.DynatraceApiToken)
	})

	t.Run("whole overrides block removed still unsets the token", func(t *testing.T) {
		out, diags := buildUpdateOverrides(ctx, noOverrides, priorWithToken, metricsExport)
		require.False(t, diags.HasError(), "diags: %v", diags)
		require.NotNil(t, out)
		require.NotNil(t, out.MetricsExport)
		assert.Equal(t, []string{"dynatraceApiToken"}, *out.UnsetFields)
	})

	t.Run("kept token is not unset", func(t *testing.T) {
		plan := newOverrides(map[string]attr.Value{"dynatrace_api_token": types.StringValue("new_token")})
		out, diags := buildUpdateOverrides(ctx, plan, priorWithToken, metricsExport)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Nil(t, out.UnsetFields)
		assert.Equal(t, lo.ToPtr("new_token"), out.MetricsExport.DynatraceApiToken)
	})

	t.Run("both secrets removed are both unset", func(t *testing.T) {
		prior := newOverrides(map[string]attr.Value{
			"data_dog_api_key":    types.StringValue("old_key"),
			"dynatrace_api_token": types.StringValue("old_token"),
		})
		out, diags := buildUpdateOverrides(ctx, noOverrides, prior, metricsExport)
		require.False(t, diags.HasError(), "diags: %v", diags)
		require.NotNil(t, out)
		assert.Equal(t, []string{"dataDogApiKey", "dynatraceApiToken"}, *out.UnsetFields)
	})
}
