package models

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EnvironmentObject describes the resource and data source data model.
type EnvironmentObject struct {
	Id                  types.String `tfsdk:"id"`
	ObjectKey           types.String `tfsdk:"object_key"`
	ObjectType          types.String `tfsdk:"object_type"`
	Scope               types.String `tfsdk:"scope"`
	ScopeEntityId       types.String `tfsdk:"scope_entity_id"`
	SourceScope         types.String `tfsdk:"source_scope"`
	SourceScopeEntityId types.String `tfsdk:"source_scope_entity_id"`
	AutoLinkDeployments types.Bool   `tfsdk:"auto_link_deployments"`
	AirflowVariable     types.Object `tfsdk:"airflow_variable"`
	ConnectionConfig    types.Object `tfsdk:"connection_config"`
	MetricsExport       types.Object `tfsdk:"metrics_export"`
	Links               types.Set    `tfsdk:"links"`
	ExcludeLinks        types.Set    `tfsdk:"exclude_links"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
	CreatedBy           types.Object `tfsdk:"created_by"`
	UpdatedBy           types.Object `tfsdk:"updated_by"`
}

// EnvironmentObjectConnectionInput is the tfsdk-tagged struct used to (de)serialize
// the connection_config attribute. Resource handlers unmarshal data.ConnectionConfig
// into this type before building API requests or extracting preserve values.
type EnvironmentObjectConnectionInput struct {
	AuthTypeId         types.String `tfsdk:"auth_type_id"`
	ConnectionAuthType types.Object `tfsdk:"connection_auth_type"`
	Type               types.String `tfsdk:"type"`
	Host               types.String `tfsdk:"host"`
	Port               types.Int64  `tfsdk:"port"`
	Schema             types.String `tfsdk:"schema"`
	Login              types.String `tfsdk:"login"`
	Password           types.String `tfsdk:"password"`
	Extra              types.String `tfsdk:"extra"`
}

// EnvironmentObjectAirflowVariableInput is the tfsdk-tagged struct for the
// airflow_variable attribute.
type EnvironmentObjectAirflowVariableInput struct {
	Value    types.String `tfsdk:"value"`
	IsSecret types.Bool   `tfsdk:"is_secret"`
}

// EnvironmentObjectMetricsExportInput is the tfsdk-tagged struct for the
// metrics_export attribute.
type EnvironmentObjectMetricsExportInput struct {
	AuthType     types.String `tfsdk:"auth_type"`
	Endpoint     types.String `tfsdk:"endpoint"`
	BasicToken   types.String `tfsdk:"basic_token"`
	ExporterType types.String `tfsdk:"exporter_type"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
	Headers      types.Map    `tfsdk:"headers"`
	Labels       types.Map    `tfsdk:"labels"`
}

// EnvironmentObjectExcludeLinkInput is the tfsdk-tagged struct for an
// exclude_links element.
type EnvironmentObjectExcludeLinkInput struct {
	Scope         types.String `tfsdk:"scope"`
	ScopeEntityId types.String `tfsdk:"scope_entity_id"`
}

// EnvironmentObjectLinkInput is the tfsdk-tagged struct for a links element.
type EnvironmentObjectLinkInput struct {
	Scope                    types.String `tfsdk:"scope"`
	ScopeEntityId            types.String `tfsdk:"scope_entity_id"`
	AirflowVariableOverrides types.Object `tfsdk:"airflow_variable_overrides"`
	ConnectionOverrides      types.Object `tfsdk:"connection_overrides"`
	MetricsExportOverrides   types.Object `tfsdk:"metrics_export_overrides"`
}

// EnvironmentObjectConnectionOverridesInput is the tfsdk-tagged struct for the
// per-link connection_overrides attribute.
type EnvironmentObjectConnectionOverridesInput struct {
	Type     types.String `tfsdk:"type"`
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
	Schema   types.String `tfsdk:"schema"`
	Login    types.String `tfsdk:"login"`
	Password types.String `tfsdk:"password"`
	Extra    types.String `tfsdk:"extra"`
}

// EnvironmentObjectMetricsExportOverridesInput is the tfsdk-tagged struct for the
// per-link metrics_export_overrides attribute.
type EnvironmentObjectMetricsExportOverridesInput struct {
	AuthType     types.String `tfsdk:"auth_type"`
	Endpoint     types.String `tfsdk:"endpoint"`
	BasicToken   types.String `tfsdk:"basic_token"`
	ExporterType types.String `tfsdk:"exporter_type"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
	Headers      types.Map    `tfsdk:"headers"`
	Labels       types.Map    `tfsdk:"labels"`
}

// EnvironmentObjectAirflowVariableOverridesInput is the tfsdk-tagged struct for the
// per-link airflow_variable_overrides attribute.
type EnvironmentObjectAirflowVariableOverridesInput struct {
	Value types.String `tfsdk:"value"`
}

// EnvironmentObjectPreserve carries values from the prior plan/state that must
// survive a refresh because the API does not echo them back on GET. Fields are
// nil when the caller has nothing to preserve (e.g. import, or data sources).
type EnvironmentObjectPreserve struct {
	// Top-level
	ConnectionPassword      *string
	ConnectionAuthTypeId    *string
	ConnectionExtra         *string // user's exact JSON string (avoids map round-trip drift)
	AirflowVariableValue    *string // only meaningful when is_secret=true
	MetricsExportPassword   *string
	MetricsExportBasicToken *string
	// Per-link overrides, keyed by LinkPreserveKey(scope, scope_entity_id).
	LinkOverrides map[string]*EnvironmentObjectLinkOverridePreserve
}

// EnvironmentObjectLinkOverridePreserve carries per-link override secrets/JSON
// that the API strips on GET.
type EnvironmentObjectLinkOverridePreserve struct {
	AirflowVariableValue    *string
	ConnectionPassword      *string
	ConnectionExtra         *string
	MetricsExportPassword   *string
	MetricsExportBasicToken *string
}

// LinkPreserveKey builds the composite key used to look up per-link preserve
// entries. Exported so the resource layer can populate the map with matching
// keys before calling ReadFromResponse.
func LinkPreserveKey(scope, scopeEntityId string) string {
	return scope + ":" + scopeEntityId
}

func (data *EnvironmentObject) ReadFromResponse(ctx context.Context, obj *platform.EnvironmentObject, preserve *EnvironmentObjectPreserve) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Id = types.StringPointerValue(obj.Id)
	data.ObjectKey = types.StringValue(obj.ObjectKey)
	data.ObjectType = types.StringValue(string(obj.ObjectType))
	data.Scope = types.StringValue(string(obj.Scope))
	data.ScopeEntityId = types.StringValue(obj.ScopeEntityId)

	if obj.SourceScope != nil {
		data.SourceScope = types.StringValue(string(*obj.SourceScope))
	} else {
		data.SourceScope = types.StringNull()
	}
	data.SourceScopeEntityId = types.StringPointerValue(obj.SourceScopeEntityId)
	data.AutoLinkDeployments = types.BoolPointerValue(obj.AutoLinkDeployments)

	data.CreatedAt = types.StringPointerValue(obj.CreatedAt)
	data.UpdatedAt = types.StringPointerValue(obj.UpdatedAt)

	if obj.CreatedBy != nil {
		data.CreatedBy, diags = SubjectProfileTypesObject(ctx, obj.CreatedBy)
		if diags.HasError() {
			return diags
		}
	} else {
		data.CreatedBy = types.ObjectNull(schemas.SubjectProfileAttributeTypes())
	}

	if obj.UpdatedBy != nil {
		data.UpdatedBy, diags = SubjectProfileTypesObject(ctx, obj.UpdatedBy)
		if diags.HasError() {
			return diags
		}
	} else {
		data.UpdatedBy = types.ObjectNull(schemas.SubjectProfileAttributeTypes())
	}

	// Airflow Variable — when is_secret, the API returns an empty value. Substitute the
	// preserved value (from plan or prior state) so Terraform state stays consistent.
	if obj.AirflowVariable != nil {
		value := obj.AirflowVariable.Value
		if preserve != nil && preserve.AirflowVariableValue != nil && obj.AirflowVariable.IsSecret {
			value = *preserve.AirflowVariableValue
		}
		data.AirflowVariable, diags = types.ObjectValue(schemas.EnvironmentObjectAirflowVariableAttributeTypes(), map[string]attr.Value{
			"value":     types.StringValue(value),
			"is_secret": types.BoolValue(obj.AirflowVariable.IsSecret),
		})
		if diags.HasError() {
			return diags
		}
	} else {
		data.AirflowVariable = types.ObjectNull(schemas.EnvironmentObjectAirflowVariableAttributeTypes())
	}

	// Connection
	if obj.Connection != nil {
		data.ConnectionConfig, diags = EnvironmentObjectConnectionTypesObject(ctx, obj.Connection, preserve)
		if diags.HasError() {
			return diags
		}
	} else {
		data.ConnectionConfig = types.ObjectNull(schemas.EnvironmentObjectConnectionAttributeTypes())
	}

	// Metrics Export
	if obj.MetricsExport != nil {
		data.MetricsExport, diags = EnvironmentObjectMetricsExportTypesObject(ctx, obj.MetricsExport, preserve)
		if diags.HasError() {
			return diags
		}
	} else {
		data.MetricsExport = types.ObjectNull(schemas.EnvironmentObjectMetricsExportAttributeTypes())
	}

	// Links — preserve the nil-vs-empty distinction from the API so state reflects
	// reality (an empty array stays an empty set, not null).
	linkObjType := types.ObjectType{AttrTypes: schemas.EnvironmentObjectLinkAttributeTypes()}
	if obj.Links != nil {
		linkObjects := make([]attr.Value, len(*obj.Links))
		for i, link := range *obj.Links {
			var linkPreserve *EnvironmentObjectLinkOverridePreserve
			if preserve != nil && preserve.LinkOverrides != nil {
				linkPreserve = preserve.LinkOverrides[LinkPreserveKey(string(link.Scope), link.ScopeEntityId)]
			}
			linkObjects[i], diags = EnvironmentObjectLinkTypesObject(ctx, &link, linkPreserve)
			if diags.HasError() {
				return diags
			}
		}
		data.Links, diags = types.SetValue(linkObjType, linkObjects)
		if diags.HasError() {
			return diags
		}
	} else {
		data.Links = types.SetNull(linkObjType)
	}

	// Exclude Links
	excludeLinkObjType := types.ObjectType{AttrTypes: schemas.EnvironmentObjectExcludeLinkAttributeTypes()}
	if obj.ExcludeLinks != nil {
		excludeLinkObjects := make([]attr.Value, len(*obj.ExcludeLinks))
		for i, el := range *obj.ExcludeLinks {
			excludeLinkObjects[i], diags = types.ObjectValue(schemas.EnvironmentObjectExcludeLinkAttributeTypes(), map[string]attr.Value{
				"scope":           types.StringValue(string(el.Scope)),
				"scope_entity_id": types.StringValue(el.ScopeEntityId),
			})
			if diags.HasError() {
				return diags
			}
		}
		data.ExcludeLinks, diags = types.SetValue(excludeLinkObjType, excludeLinkObjects)
		if diags.HasError() {
			return diags
		}
	} else {
		data.ExcludeLinks = types.SetNull(excludeLinkObjType)
	}

	return nil
}

// EnvironmentObjectConnectionTypesObject converts a platform.EnvironmentObjectConnection
// into a types.Object matching the connection_config schema. The preserve argument
// supplies the user's auth_type_id / password / extra values when the API does not
// echo them back; pass nil from data sources or other read-only contexts.
func EnvironmentObjectConnectionTypesObject(ctx context.Context, conn *platform.EnvironmentObjectConnection, preserve *EnvironmentObjectPreserve) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	var authTypeObj types.Object
	if conn.ConnectionAuthType != nil {
		authTypeObj, diags = EnvironmentObjectConnectionAuthTypeTypesObject(ctx, conn.ConnectionAuthType)
		if diags.HasError() {
			return types.Object{}, diags
		}
	} else {
		authTypeObj = types.ObjectNull(schemas.EnvironmentObjectConnectionAuthTypeAttributeTypes())
	}

	// auth_type_id is provided by the user on create/update but is not returned by the
	// API on GET (the resolved object lives on connection_auth_type). Preserve the user
	// value if present; otherwise fall back to the resolved connection_auth_type.id so
	// import populates a stable value.
	var authTypeIdVal types.String
	switch {
	case preserve != nil && preserve.ConnectionAuthTypeId != nil:
		authTypeIdVal = types.StringValue(*preserve.ConnectionAuthTypeId)
	case conn.ConnectionAuthType != nil:
		authTypeIdVal = types.StringValue(conn.ConnectionAuthType.Id)
	default:
		authTypeIdVal = types.StringNull()
	}

	// extra: keep the caller's exact JSON string when provided. Go's json.Marshal
	// reorders keys and re-formats numbers, which causes a permadiff against the
	// user's original jsonencode(...) string.
	var extraVal types.String
	switch {
	case preserve != nil && preserve.ConnectionExtra != nil:
		extraVal = types.StringValue(*preserve.ConnectionExtra)
	case conn.Extra != nil:
		extraBytes, err := json.Marshal(conn.Extra)
		if err != nil {
			return types.Object{}, diag.Diagnostics{diag.NewErrorDiagnostic("Internal Error", fmt.Sprintf("Failed to marshal connection extra to JSON: %s", err))}
		}
		extraVal = types.StringValue(string(extraBytes))
	default:
		extraVal = types.StringNull()
	}

	var portVal types.Int64
	if conn.Port != nil {
		portVal = types.Int64Value(int64(*conn.Port))
	} else {
		portVal = types.Int64Null()
	}

	passwordVal := preserveSecret(conn.Password, ifPreserve(preserve, func(p *EnvironmentObjectPreserve) *string { return p.ConnectionPassword }))

	return types.ObjectValue(schemas.EnvironmentObjectConnectionAttributeTypes(), map[string]attr.Value{
		"auth_type_id":         authTypeIdVal,
		"connection_auth_type": authTypeObj,
		"type":                 types.StringValue(conn.Type),
		"host":                 types.StringPointerValue(conn.Host),
		"port":                 portVal,
		"schema":               types.StringPointerValue(conn.Schema),
		"login":                types.StringPointerValue(conn.Login),
		"password":             passwordVal,
		"extra":                extraVal,
	})
}

// EnvironmentObjectConnectionAuthTypeTypesObject converts a platform.ConnectionAuthType
// into a types.Object matching the connection_auth_type schema.
func EnvironmentObjectConnectionAuthTypeTypesObject(ctx context.Context, cat *platform.ConnectionAuthType) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	paramAttrType := schemas.EnvironmentObjectConnectionAuthTypeParameterAttributeTypes()
	paramObjects := make([]attr.Value, len(cat.Parameters))
	for i, p := range cat.Parameters {
		paramObjects[i], diags = types.ObjectValue(paramAttrType, map[string]attr.Value{
			"airflow_param_name": types.StringValue(p.AirflowParamName),
			"friendly_name":      types.StringValue(p.FriendlyName),
			"data_type":          types.StringValue(p.DataType),
			"is_required":        types.BoolValue(p.IsRequired),
			"is_secret":          types.BoolValue(p.IsSecret),
			"description":        types.StringValue(p.Description),
			"example":            types.StringPointerValue(p.Example),
			"is_in_extra":        types.BoolValue(p.IsInExtra),
			"pattern":            types.StringPointerValue(p.Pattern),
		})
		if diags.HasError() {
			return types.Object{}, diags
		}
	}

	paramsList, diags := types.ListValue(types.ObjectType{AttrTypes: paramAttrType}, paramObjects)
	if diags.HasError() {
		return types.Object{}, diags
	}

	return types.ObjectValue(schemas.EnvironmentObjectConnectionAuthTypeAttributeTypes(), map[string]attr.Value{
		"parameters":            paramsList,
		"id":                    types.StringValue(cat.Id),
		"name":                  types.StringValue(cat.Name),
		"auth_method_name":      types.StringValue(cat.AuthMethodName),
		"airflow_type":          types.StringValue(cat.AirflowType),
		"description":           types.StringValue(cat.Description),
		"provider_package_name": types.StringValue(cat.ProviderPackageName),
		"provider_logo":         types.StringPointerValue(cat.ProviderLogo),
		"guide_path":            types.StringPointerValue(cat.GuidePath),
	})
}

// EnvironmentObjectMetricsExportTypesObject converts a
// platform.EnvironmentObjectMetricsExport into a types.Object matching the
// metrics_export schema. The preserve argument supplies the user's secrets when
// the API does not echo them back.
func EnvironmentObjectMetricsExportTypesObject(ctx context.Context, me *platform.EnvironmentObjectMetricsExport, preserve *EnvironmentObjectPreserve) (types.Object, diag.Diagnostics) {
	var authTypeVal types.String
	if me.AuthType != nil {
		authTypeVal = types.StringValue(string(*me.AuthType))
	} else {
		authTypeVal = types.StringNull()
	}

	headers, diags := stringMapToTFMap(me.Headers)
	if diags.HasError() {
		return types.Object{}, diags
	}

	labels, diags := stringMapToTFMap(me.Labels)
	if diags.HasError() {
		return types.Object{}, diags
	}

	basicToken := preserveSecret(me.BasicToken, ifPreserve(preserve, func(p *EnvironmentObjectPreserve) *string { return p.MetricsExportBasicToken }))
	password := preserveSecret(me.Password, ifPreserve(preserve, func(p *EnvironmentObjectPreserve) *string { return p.MetricsExportPassword }))

	return types.ObjectValue(schemas.EnvironmentObjectMetricsExportAttributeTypes(), map[string]attr.Value{
		"auth_type":     authTypeVal,
		"endpoint":      types.StringValue(me.Endpoint),
		"basic_token":   basicToken,
		"exporter_type": types.StringValue(string(me.ExporterType)),
		"username":      types.StringPointerValue(me.Username),
		"password":      password,
		"headers":       headers,
		"labels":        labels,
	})
}

// EnvironmentObjectLinkTypesObject converts a platform.EnvironmentObjectLink into a
// types.Object matching the link schema. The preserve argument supplies any
// per-link override secrets/extra JSON that the API does not echo back.
func EnvironmentObjectLinkTypesObject(ctx context.Context, link *platform.EnvironmentObjectLink, preserve *EnvironmentObjectLinkOverridePreserve) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Airflow variable overrides
	var avOverrides types.Object
	if link.AirflowVariableOverrides != nil {
		value := link.AirflowVariableOverrides.Value
		if preserve != nil && preserve.AirflowVariableValue != nil {
			value = *preserve.AirflowVariableValue
		}
		avOverrides, diags = types.ObjectValue(schemas.EnvironmentObjectAirflowVariableOverridesAttributeTypes(), map[string]attr.Value{
			"value": types.StringValue(value),
		})
		if diags.HasError() {
			return types.Object{}, diags
		}
	} else {
		avOverrides = types.ObjectNull(schemas.EnvironmentObjectAirflowVariableOverridesAttributeTypes())
	}

	// Connection overrides
	var connOverrides types.Object
	if link.ConnectionOverrides != nil {
		co := link.ConnectionOverrides
		var extraVal types.String
		switch {
		case preserve != nil && preserve.ConnectionExtra != nil:
			extraVal = types.StringValue(*preserve.ConnectionExtra)
		case co.Extra != nil:
			extraBytes, err := json.Marshal(co.Extra)
			if err != nil {
				return types.Object{}, diag.Diagnostics{diag.NewErrorDiagnostic("Internal Error", fmt.Sprintf("Failed to marshal connection overrides extra: %s", err))}
			}
			extraVal = types.StringValue(string(extraBytes))
		default:
			extraVal = types.StringNull()
		}

		var portVal types.Int64
		if co.Port != nil {
			portVal = types.Int64Value(int64(*co.Port))
		} else {
			portVal = types.Int64Null()
		}

		passwordVal := preserveSecretLink(co.Password, preserve, func(p *EnvironmentObjectLinkOverridePreserve) *string { return p.ConnectionPassword })

		connOverrides, diags = types.ObjectValue(schemas.EnvironmentObjectConnectionOverridesAttributeTypes(), map[string]attr.Value{
			"type":     types.StringPointerValue(co.Type),
			"host":     types.StringPointerValue(co.Host),
			"port":     portVal,
			"schema":   types.StringPointerValue(co.Schema),
			"login":    types.StringPointerValue(co.Login),
			"password": passwordVal,
			"extra":    extraVal,
		})
		if diags.HasError() {
			return types.Object{}, diags
		}
	} else {
		connOverrides = types.ObjectNull(schemas.EnvironmentObjectConnectionOverridesAttributeTypes())
	}

	// Metrics export overrides
	var meOverrides types.Object
	if link.MetricsExportOverrides != nil {
		mo := link.MetricsExportOverrides
		var authTypeVal types.String
		if mo.AuthType != nil {
			authTypeVal = types.StringValue(string(*mo.AuthType))
		} else {
			authTypeVal = types.StringNull()
		}
		var exporterTypeVal types.String
		if mo.ExporterType != nil {
			exporterTypeVal = types.StringValue(string(*mo.ExporterType))
		} else {
			exporterTypeVal = types.StringNull()
		}
		headers, d := stringMapToTFMap(mo.Headers)
		if d.HasError() {
			return types.Object{}, d
		}
		labels, d := stringMapToTFMap(mo.Labels)
		if d.HasError() {
			return types.Object{}, d
		}

		basicToken := preserveSecretLink(mo.BasicToken, preserve, func(p *EnvironmentObjectLinkOverridePreserve) *string { return p.MetricsExportBasicToken })
		password := preserveSecretLink(mo.Password, preserve, func(p *EnvironmentObjectLinkOverridePreserve) *string { return p.MetricsExportPassword })

		meOverrides, diags = types.ObjectValue(schemas.EnvironmentObjectMetricsExportOverridesAttributeTypes(), map[string]attr.Value{
			"auth_type":     authTypeVal,
			"endpoint":      types.StringPointerValue(mo.Endpoint),
			"basic_token":   basicToken,
			"exporter_type": exporterTypeVal,
			"username":      types.StringPointerValue(mo.Username),
			"password":      password,
			"headers":       headers,
			"labels":        labels,
		})
		if diags.HasError() {
			return types.Object{}, diags
		}
	} else {
		meOverrides = types.ObjectNull(schemas.EnvironmentObjectMetricsExportOverridesAttributeTypes())
	}

	return types.ObjectValue(schemas.EnvironmentObjectLinkAttributeTypes(), map[string]attr.Value{
		"scope":                      types.StringValue(string(link.Scope)),
		"scope_entity_id":            types.StringValue(link.ScopeEntityId),
		"airflow_variable_overrides": avOverrides,
		"connection_overrides":       connOverrides,
		"metrics_export_overrides":   meOverrides,
	})
}

func stringMapToTFMap(m *map[string]string) (types.Map, diag.Diagnostics) {
	if m == nil {
		return types.MapNull(types.StringType), nil
	}
	elems := make(map[string]attr.Value, len(*m))
	for k, v := range *m {
		elems[k] = types.StringValue(v)
	}
	return types.MapValue(types.StringType, elems)
}

// preserveSecret returns a types.String for a sensitive field. When the API
// returns nil or an empty string (both are common "redacted" signals) AND the
// caller supplied a preserved value, the preserved value wins. Otherwise the
// API value is used.
func preserveSecret(apiVal, preserveVal *string) types.String {
	if preserveVal != nil && (apiVal == nil || *apiVal == "") {
		return types.StringValue(*preserveVal)
	}
	return types.StringPointerValue(apiVal)
}

// preserveSecretLink is the per-link variant of preserveSecret.
func preserveSecretLink(apiVal *string, preserve *EnvironmentObjectLinkOverridePreserve, pick func(*EnvironmentObjectLinkOverridePreserve) *string) types.String {
	if preserve != nil {
		return preserveSecret(apiVal, pick(preserve))
	}
	return types.StringPointerValue(apiVal)
}

// ifPreserve guards a preserve-field lookup against a nil EnvironmentObjectPreserve.
func ifPreserve(p *EnvironmentObjectPreserve, pick func(*EnvironmentObjectPreserve) *string) *string {
	if p == nil {
		return nil
	}
	return pick(p)
}
