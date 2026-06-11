package models_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
)

// TestUnit_EnvironmentObject_ReadFromResponse_AirflowVariable_Preserve covers
// the AIRFLOW_VARIABLE preserve / round-trip rules:
//   - Non-secret values: API echoes them; preserve is ignored.
//   - Secret values: API returns empty (the redaction sentinel) and the model
//     falls back to the preserved plan value so secrets survive refresh.
//   - With no preserve (e.g. import flow), the API value wins as-is.
func TestUnit_EnvironmentObject_ReadFromResponse_AirflowVariable_Preserve(t *testing.T) {
	tests := []struct {
		name        string
		apiValue    string
		apiIsSecret bool
		preserve    *models.EnvironmentObjectPreserve
		wantValue   string
	}{
		{
			name:        "non-secret value: API echoes the value, preserve ignored",
			apiValue:    "from_api",
			apiIsSecret: false,
			preserve:    &models.EnvironmentObjectPreserve{AirflowVariableValue: lo.ToPtr("from_preserve")},
			wantValue:   "from_api",
		},
		{
			name:        "secret value: API returns empty redaction sentinel, preserve restores it",
			apiValue:    "",
			apiIsSecret: true,
			preserve:    &models.EnvironmentObjectPreserve{AirflowVariableValue: lo.ToPtr("the_secret")},
			wantValue:   "the_secret",
		},
		{
			name:        "secret value with nil preserve (import path): value stays whatever API returned",
			apiValue:    "",
			apiIsSecret: true,
			preserve:    nil,
			wantValue:   "",
		},
		{
			name:        "non-secret value with nil preserve: API value wins",
			apiValue:    "from_api",
			apiIsSecret: false,
			preserve:    nil,
			wantValue:   "from_api",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			obj := newAirflowVariableAPIResponse(tc.apiValue, tc.apiIsSecret)
			var data models.EnvironmentObject
			diags := data.ReadFromResponse(context.Background(), obj, tc.preserve)
			require.False(t, diags.HasError(), "ReadFromResponse returned errors: %v", diags)
			assert.Equal(t, tc.wantValue, data.Value.ValueString())
			assert.Equal(t, tc.apiIsSecret, data.IsSecret.ValueBool())
		})
	}
}

// TestUnit_EnvironmentObject_ReadFromResponse_Connection_Preserve covers the
// CONNECTION preserve rules for fields the API doesn't echo back on GET:
// password, auth_type_id, and extra (with its byte-exact JSON preservation
// requirement to avoid map-key reorder permadiffs).
func TestUnit_EnvironmentObject_ReadFromResponse_Connection_Preserve(t *testing.T) {
	t.Run("password preserved when API returns nil; preserve wins", func(t *testing.T) {
		obj := newConnectionAPIResponse(connectionAPI{})
		preserve := &models.EnvironmentObjectPreserve{Password: lo.ToPtr("from_plan")}
		var data models.EnvironmentObject
		diags := data.ReadFromResponse(context.Background(), obj, preserve)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Equal(t, "from_plan", data.Password.ValueString())
	})

	t.Run("password from API (when API does echo it) wins over preserve", func(t *testing.T) {
		obj := newConnectionAPIResponse(connectionAPI{Password: lo.ToPtr("from_api")})
		preserve := &models.EnvironmentObjectPreserve{Password: lo.ToPtr("from_plan")}
		var data models.EnvironmentObject
		diags := data.ReadFromResponse(context.Background(), obj, preserve)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Equal(t, "from_api", data.Password.ValueString())
	})

	t.Run("password nil-preserve + nil-API: state is null", func(t *testing.T) {
		obj := newConnectionAPIResponse(connectionAPI{})
		var data models.EnvironmentObject
		diags := data.ReadFromResponse(context.Background(), obj, nil)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.True(t, data.Password.IsNull(), "expected null password, got %q", data.Password.ValueString())
	})

	t.Run("auth_type_id always uses preserved plan value (API is write-only)", func(t *testing.T) {
		// API never echoes auth_type_id, only the resolved nested
		// connection_auth_type object. preserve is the only source.
		obj := newConnectionAPIResponse(connectionAPI{})
		preserve := &models.EnvironmentObjectPreserve{AuthTypeId: lo.ToPtr("snowflake-password")}
		var data models.EnvironmentObject
		diags := data.ReadFromResponse(context.Background(), obj, preserve)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Equal(t, "snowflake-password", data.AuthTypeId.ValueString())
	})

	t.Run("extra: preserved plan JSON wins over API-marshalled JSON to avoid key reorder drift", func(t *testing.T) {
		// API returns extra as a map[string]interface{}. If we json.Marshal it,
		// key order is non-deterministic and would cause a permadiff against the
		// user's original `jsonencode({...})` byte string. The preserve fallback
		// keeps the user's exact string.
		apiExtra := map[string]interface{}{"sslmode": "require", "timeout": 30.0}
		obj := newConnectionAPIResponse(connectionAPI{Extra: &apiExtra})
		userExtra := `{"timeout":30,"sslmode":"require"}` // user wrote keys in this order
		preserve := &models.EnvironmentObjectPreserve{Extra: lo.ToPtr(userExtra)}
		var data models.EnvironmentObject
		diags := data.ReadFromResponse(context.Background(), obj, preserve)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Equal(t, userExtra, data.Extra.ValueString(),
			"preserve should win to keep byte-exact user JSON")
	})

	t.Run("extra: nil preserve, API returns extra: state shows API JSON", func(t *testing.T) {
		apiExtra := map[string]interface{}{"sslmode": "require"}
		obj := newConnectionAPIResponse(connectionAPI{Extra: &apiExtra})
		var data models.EnvironmentObject
		diags := data.ReadFromResponse(context.Background(), obj, nil)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Equal(t, `{"sslmode":"require"}`, data.Extra.ValueString())
	})

	t.Run("extra: nil preserve, nil API: state is null", func(t *testing.T) {
		obj := newConnectionAPIResponse(connectionAPI{})
		var data models.EnvironmentObject
		diags := data.ReadFromResponse(context.Background(), obj, nil)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.True(t, data.Extra.IsNull())
	})
}

// TestUnit_EnvironmentObject_ReadFromResponse_MetricsExport_Preserve covers
// the METRICS_EXPORT preserve rules for fields the API doesn't echo back:
// password, basic_token, and auth_type.
func TestUnit_EnvironmentObject_ReadFromResponse_MetricsExport_Preserve(t *testing.T) {
	authToken := platform.EnvironmentObjectMetricsExportAuthTypeAUTHTOKEN

	t.Run("basic_token preserved when API returns nil", func(t *testing.T) {
		obj := newMetricsExportAPIResponse(nil, nil, nil)
		preserve := &models.EnvironmentObjectPreserve{BasicToken: lo.ToPtr("token_from_plan")}
		var data models.EnvironmentObject
		diags := data.ReadFromResponse(context.Background(), obj, preserve)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Equal(t, "token_from_plan", data.BasicToken.ValueString())
	})

	t.Run("auth_type preserved when API returns nil (write-only API behavior)", func(t *testing.T) {
		obj := newMetricsExportAPIResponse(nil, nil, nil)
		preserve := &models.EnvironmentObjectPreserve{MetricsExportAuthType: lo.ToPtr("AUTH_TOKEN")}
		var data models.EnvironmentObject
		diags := data.ReadFromResponse(context.Background(), obj, preserve)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Equal(t, "AUTH_TOKEN", data.AuthType.ValueString())
	})

	t.Run("auth_type from API wins over preserve when both set", func(t *testing.T) {
		obj := newMetricsExportAPIResponse(&authToken, nil, nil)
		preserve := &models.EnvironmentObjectPreserve{MetricsExportAuthType: lo.ToPtr("BASIC")}
		var data models.EnvironmentObject
		diags := data.ReadFromResponse(context.Background(), obj, preserve)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Equal(t, "AUTH_TOKEN", data.AuthType.ValueString())
	})

	t.Run("all write-only fields with nil preserve: state is null", func(t *testing.T) {
		obj := newMetricsExportAPIResponse(nil, nil, nil)
		var data models.EnvironmentObject
		diags := data.ReadFromResponse(context.Background(), obj, nil)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.True(t, data.AuthType.IsNull())
		assert.True(t, data.BasicToken.IsNull())
		assert.True(t, data.Password.IsNull())
	})
}

// TestUnit_EnvironmentObject_ReadFromResponse_LinkOverridesPreserve covers the
// per-link override preserve path: secrets, JSON, and auth_type must survive
// when the API doesn't echo them back. Lookup is keyed by scope:scope_entity_id.
func TestUnit_EnvironmentObject_ReadFromResponse_LinkOverridesPreserve(t *testing.T) {
	deploymentScope := platform.EnvironmentObjectLinkScopeDEPLOYMENT
	depId := "cmq6m73on0hnl01ktk3xftlwf"

	t.Run("airflow_variable link override: secret value falls back to preserve", func(t *testing.T) {
		obj := newAirflowVariableAPIResponse("ws_value", false)
		obj.Links = &[]platform.EnvironmentObjectLink{{
			Scope:                    deploymentScope,
			ScopeEntityId:            depId,
			AirflowVariableOverrides: &platform.EnvironmentObjectAirflowVariableOverrides{Value: ""},
		}}
		preserve := &models.EnvironmentObjectPreserve{
			LinkOverrides: map[string]*models.EnvironmentObjectLinkOverridePreserve{
				models.LinkPreserveKey("DEPLOYMENT", depId): {Value: lo.ToPtr("preserved_override")},
			},
		}
		var data models.EnvironmentObject
		diags := data.ReadFromResponse(context.Background(), obj, preserve)
		require.False(t, diags.HasError(), "diags: %v", diags)
		got := overrideValueFromLinks(t, data.Links, "value")
		assert.Equal(t, "preserved_override", got)
	})

	t.Run("metrics_export link override: auth_type + basic_token + password fall back to preserve", func(t *testing.T) {
		exporter := platform.EnvironmentObjectMetricsExportAuthTypeAUTHTOKEN
		_ = exporter
		obj := newMetricsExportAPIResponse(nil, nil, nil)
		obj.Links = &[]platform.EnvironmentObjectLink{{
			Scope:                  deploymentScope,
			ScopeEntityId:          depId,
			MetricsExportOverrides: &platform.EnvironmentObjectMetricsExportOverrides{
				Endpoint: lo.ToPtr("https://override.example.com/api/v1/write"),
				// API does NOT echo auth_type, basic_token, password
			},
		}}
		preserve := &models.EnvironmentObjectPreserve{
			LinkOverrides: map[string]*models.EnvironmentObjectLinkOverridePreserve{
				models.LinkPreserveKey("DEPLOYMENT", depId): {
					AuthType:   lo.ToPtr("AUTH_TOKEN"),
					BasicToken: lo.ToPtr("override_token"),
					Password:   lo.ToPtr("override_password"),
				},
			},
		}
		var data models.EnvironmentObject
		diags := data.ReadFromResponse(context.Background(), obj, preserve)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Equal(t, "AUTH_TOKEN", overrideValueFromLinks(t, data.Links, "auth_type"))
		assert.Equal(t, "override_token", overrideValueFromLinks(t, data.Links, "basic_token"))
		assert.Equal(t, "override_password", overrideValueFromLinks(t, data.Links, "password"))
	})

	t.Run("connection link override: extra preserve keeps byte-exact JSON", func(t *testing.T) {
		obj := newConnectionAPIResponse(connectionAPI{})
		obj.Links = &[]platform.EnvironmentObjectLink{{
			Scope:               deploymentScope,
			ScopeEntityId:       depId,
			ConnectionOverrides: &platform.EnvironmentObjectConnectionOverrides{
				Host: lo.ToPtr("override.example.com"),
			},
		}}
		userExtra := `{"sslmode":"prefer","timeout":15}`
		preserve := &models.EnvironmentObjectPreserve{
			LinkOverrides: map[string]*models.EnvironmentObjectLinkOverridePreserve{
				models.LinkPreserveKey("DEPLOYMENT", depId): {Extra: lo.ToPtr(userExtra)},
			},
		}
		var data models.EnvironmentObject
		diags := data.ReadFromResponse(context.Background(), obj, preserve)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.Equal(t, userExtra, overrideValueFromLinks(t, data.Links, "extra"))
	})

	t.Run("preserve key mismatch: link present but key doesn't match → fields stay null/empty", func(t *testing.T) {
		obj := newAirflowVariableAPIResponse("ws_value", false)
		obj.Links = &[]platform.EnvironmentObjectLink{{
			Scope:                    deploymentScope,
			ScopeEntityId:            depId,
			AirflowVariableOverrides: &platform.EnvironmentObjectAirflowVariableOverrides{Value: ""},
		}}
		preserve := &models.EnvironmentObjectPreserve{
			LinkOverrides: map[string]*models.EnvironmentObjectLinkOverridePreserve{
				// Key for a different deployment — should not be consulted
				models.LinkPreserveKey("DEPLOYMENT", "different_deployment_id"): {Value: lo.ToPtr("wrong_match")},
			},
		}
		var data models.EnvironmentObject
		diags := data.ReadFromResponse(context.Background(), obj, preserve)
		require.False(t, diags.HasError(), "diags: %v", diags)
		// API echoed empty value, no preserve match → value is empty (not "wrong_match")
		assert.Equal(t, "", overrideValueFromLinks(t, data.Links, "value"))
	})
}

// --- helpers ---

type connectionAPI struct {
	Host     *string
	Port     *int
	Login    *string
	Schema   *string
	Password *string
	Extra    *map[string]interface{}
}

func newAirflowVariableAPIResponse(value string, isSecret bool) *platform.EnvironmentObject {
	id := "cm6envobjid000airflow"
	scope := platform.EnvironmentObjectScopeWORKSPACE
	return &platform.EnvironmentObject{
		Id:            &id,
		ObjectKey:     "k",
		ObjectType:    platform.EnvironmentObjectObjectTypeAIRFLOWVARIABLE,
		Scope:         scope,
		ScopeEntityId: "ws_id",
		AirflowVariable: &platform.EnvironmentObjectAirflowVariable{
			Value:    value,
			IsSecret: isSecret,
		},
	}
}

func newConnectionAPIResponse(c connectionAPI) *platform.EnvironmentObject {
	id := "cm6envobjid000conn"
	scope := platform.EnvironmentObjectScopeWORKSPACE
	return &platform.EnvironmentObject{
		Id:            &id,
		ObjectKey:     "k",
		ObjectType:    platform.EnvironmentObjectObjectTypeCONNECTION,
		Scope:         scope,
		ScopeEntityId: "ws_id",
		Connection: &platform.EnvironmentObjectConnection{
			Type:     "postgres",
			Host:     c.Host,
			Port:     c.Port,
			Login:    c.Login,
			Schema:   c.Schema,
			Password: c.Password,
			Extra:    c.Extra,
		},
	}
}

func newMetricsExportAPIResponse(
	authType *platform.EnvironmentObjectMetricsExportAuthType,
	username *string,
	basicToken *string,
) *platform.EnvironmentObject {
	id := "cm6envobjid000metric"
	scope := platform.EnvironmentObjectScopeWORKSPACE
	return &platform.EnvironmentObject{
		Id:            &id,
		ObjectKey:     "k",
		ObjectType:    platform.EnvironmentObjectObjectTypeMETRICSEXPORT,
		Scope:         scope,
		ScopeEntityId: "ws_id",
		MetricsExport: &platform.EnvironmentObjectMetricsExport{
			Endpoint:     "https://prom.example.com/api/v1/write",
			ExporterType: platform.EnvironmentObjectMetricsExportExporterTypePROMETHEUS,
			AuthType:     authType,
			Username:     username,
			BasicToken:   basicToken,
		},
	}
}

// overrideValueFromLinks pulls a string attribute out of the first link's
// overrides nested object. Fails the test if the structure isn't as expected.
func overrideValueFromLinks(t *testing.T, links types.Set, attrName string) string {
	t.Helper()
	require.False(t, links.IsNull() || links.IsUnknown(), "links is null/unknown")
	elems := links.Elements()
	require.Len(t, elems, 1, "expected exactly 1 link")
	linkObj, ok := elems[0].(types.Object)
	require.True(t, ok, "link element is not types.Object")
	linkAttrs := linkObj.Attributes()
	overrides, ok := linkAttrs["overrides"].(types.Object)
	require.True(t, ok, "overrides is not types.Object")
	overrideAttrs := overrides.Attributes()
	val, ok := overrideAttrs[attrName]
	require.True(t, ok, "overrides[%s] not present", attrName)
	// String, Int, etc. — we only need string assertions for the cases above.
	if sv, ok := val.(types.String); ok {
		return sv.ValueString()
	}
	t.Fatalf("overrides[%s] is not types.String, got %T", attrName, val)
	return ""
}

// silence unused (in case the value extractor isn't needed in some compile mode)
var _ = attr.Value(types.StringNull())
var _ = schemas.EnvironmentObjectsElementAttributeTypes
