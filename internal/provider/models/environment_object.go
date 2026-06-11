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

// EnvironmentObject describes the resource and data source data model. Type-specific
// fields sit directly on the struct; the parent `object_type` discriminates which
// fields are populated for any given object. Matches the flat-polymorphic shape
// used by astro_notification_channel.
type EnvironmentObject struct {
	// Identity / common
	Id                  types.String `tfsdk:"id"`
	ObjectKey           types.String `tfsdk:"object_key"`
	ObjectType          types.String `tfsdk:"object_type"`
	Scope               types.String `tfsdk:"scope"`
	ScopeEntityId       types.String `tfsdk:"scope_entity_id"`
	SourceScope         types.String `tfsdk:"source_scope"`
	SourceScopeEntityId types.String `tfsdk:"source_scope_entity_id"`
	AutoLinkDeployments types.Bool   `tfsdk:"auto_link_deployments"`
	// AIRFLOW_VARIABLE
	Value    types.String `tfsdk:"value"`
	IsSecret types.Bool   `tfsdk:"is_secret"`
	// CONNECTION
	Type               types.String `tfsdk:"type"`
	Host               types.String `tfsdk:"host"`
	Port               types.Int64  `tfsdk:"port"`
	Schema             types.String `tfsdk:"schema"`
	Login              types.String `tfsdk:"login"`
	Extra              types.String `tfsdk:"extra"`
	AuthTypeId         types.String `tfsdk:"auth_type_id"`
	ConnectionAuthType types.Object `tfsdk:"connection_auth_type"`
	// METRICS_EXPORT
	AuthType     types.String `tfsdk:"auth_type"`
	Endpoint     types.String `tfsdk:"endpoint"`
	BasicToken   types.String `tfsdk:"basic_token"`
	ExporterType types.String `tfsdk:"exporter_type"`
	Username     types.String `tfsdk:"username"`
	Headers      types.Map    `tfsdk:"headers"`
	Labels       types.Map    `tfsdk:"labels"`
	// Polymorphic (CONNECTION + METRICS_EXPORT)
	Password types.String `tfsdk:"password"`
	// Links
	Links        types.Set `tfsdk:"links"`
	ExcludeLinks types.Set `tfsdk:"exclude_links"`
	// Metadata
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	CreatedBy types.Object `tfsdk:"created_by"`
	UpdatedBy types.Object `tfsdk:"updated_by"`
}

// EnvironmentObjectExcludeLinkInput is the tfsdk-tagged struct for an
// exclude_links element.
type EnvironmentObjectExcludeLinkInput struct {
	Scope         types.String `tfsdk:"scope"`
	ScopeEntityId types.String `tfsdk:"scope_entity_id"`
}

// EnvironmentObjectLinkInput is the tfsdk-tagged struct for a links element.
// Overrides are a single flat object (one field per overridable attribute),
// discriminated by the parent's object_type.
type EnvironmentObjectLinkInput struct {
	Scope         types.String `tfsdk:"scope"`
	ScopeEntityId types.String `tfsdk:"scope_entity_id"`
	Overrides     types.Object `tfsdk:"overrides"`
}

// EnvironmentObjectOverridesInput is the tfsdk-tagged struct for the flat
// per-link `overrides` block.
type EnvironmentObjectOverridesInput struct {
	// AIRFLOW_VARIABLE
	Value types.String `tfsdk:"value"`
	// CONNECTION
	Type   types.String `tfsdk:"type"`
	Host   types.String `tfsdk:"host"`
	Port   types.Int64  `tfsdk:"port"`
	Schema types.String `tfsdk:"schema"`
	Login  types.String `tfsdk:"login"`
	Extra  types.String `tfsdk:"extra"`
	// METRICS_EXPORT
	AuthType     types.String `tfsdk:"auth_type"`
	Endpoint     types.String `tfsdk:"endpoint"`
	BasicToken   types.String `tfsdk:"basic_token"`
	ExporterType types.String `tfsdk:"exporter_type"`
	Username     types.String `tfsdk:"username"`
	Headers      types.Map    `tfsdk:"headers"`
	Labels       types.Map    `tfsdk:"labels"`
	// Polymorphic
	Password types.String `tfsdk:"password"`
}

// EnvironmentObjectPreserve carries values from the prior plan/state that must
// survive a refresh because the API does not echo them back on GET. Fields are
// nil when the caller has nothing to preserve (e.g. import, or data sources).
//
// Connection vs MetricsExport passwords are stored under distinct keys here
// because the preserve struct is built from the *typed* plan field — the
// resource layer knows which kind of password it's preserving based on the
// parent object_type.
type EnvironmentObjectPreserve struct {
	Password              *string // CONNECTION password OR METRICS_EXPORT basic-auth password (object_type discriminates)
	AuthTypeId            *string // CONNECTION
	Extra                 *string // CONNECTION — user's exact JSON string (avoids map round-trip drift)
	AirflowVariableValue  *string // AIRFLOW_VARIABLE — only meaningful when is_secret=true
	BasicToken            *string // METRICS_EXPORT
	MetricsExportAuthType *string // METRICS_EXPORT — API does not echo it back on GET
	// Per-link overrides, keyed by LinkPreserveKey(scope, scope_entity_id).
	LinkOverrides map[string]*EnvironmentObjectLinkOverridePreserve
}

// EnvironmentObjectLinkOverridePreserve carries per-link override secrets/JSON
// that the API strips on GET. Same polymorphism rules as the parent.
type EnvironmentObjectLinkOverridePreserve struct {
	Password   *string
	Extra      *string
	BasicToken *string
	Value      *string
	AuthType   *string // METRICS_EXPORT override — API does not echo it back
}

// LinkPreserveKey builds the composite key used to look up per-link preserve
// entries. Exported so the resource layer can populate the map with matching
// keys before calling ReadFromResponse.
func LinkPreserveKey(scope, scopeEntityId string) string {
	return scope + ":" + scopeEntityId
}

func (data *EnvironmentObject) ReadFromResponse(ctx context.Context, obj *platform.EnvironmentObject, preserve *EnvironmentObjectPreserve) diag.Diagnostics {
	var diags diag.Diagnostics

	// Common / metadata
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
	// Normalize: the Astro API echoes back nil for auto_link_deployments=false,
	// so an explicit `false` in config would otherwise drift to null in state.
	// Always project to a concrete bool here; the schema's UseStateForUnknown
	// plan modifier prevents the omit case from reflooding as drift on refresh.
	if obj.AutoLinkDeployments != nil {
		data.AutoLinkDeployments = types.BoolValue(*obj.AutoLinkDeployments)
	} else {
		data.AutoLinkDeployments = types.BoolValue(false)
	}

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

	// Type-specific population. Every field is set on every refresh so that
	// state matches the schema shape exactly (no "extra attribute"/"missing
	// attribute" errors). Unused fields are explicitly null.
	data.nullAllTypeSpecific()

	if obj.AirflowVariable != nil {
		value := obj.AirflowVariable.Value
		if preserve != nil && preserve.AirflowVariableValue != nil && obj.AirflowVariable.IsSecret {
			value = *preserve.AirflowVariableValue
		}
		data.Value = types.StringValue(value)
		data.IsSecret = types.BoolValue(obj.AirflowVariable.IsSecret)
	}

	if obj.Connection != nil {
		diags = data.populateConnection(ctx, obj.Connection, preserve)
		if diags.HasError() {
			return diags
		}
	}

	if obj.MetricsExport != nil {
		if d := data.populateMetricsExport(obj.MetricsExport, preserve); d.HasError() {
			return d
		}
	}

	// Links — preserve the nil-vs-empty distinction from the API so state
	// reflects reality (an empty array stays an empty set, not null).
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

// nullAllTypeSpecific sets every type-specific field to its typed null so that
// fields not relevant for the current object_type are explicitly absent from
// state. populateXxx calls below overwrite the relevant subset.
func (data *EnvironmentObject) nullAllTypeSpecific() {
	// AIRFLOW_VARIABLE
	data.Value = types.StringNull()
	data.IsSecret = types.BoolNull()
	// CONNECTION
	data.Type = types.StringNull()
	data.Host = types.StringNull()
	data.Port = types.Int64Null()
	data.Schema = types.StringNull()
	data.Login = types.StringNull()
	data.Extra = types.StringNull()
	data.AuthTypeId = types.StringNull()
	data.ConnectionAuthType = types.ObjectNull(schemas.EnvironmentObjectConnectionAuthTypeAttributeTypes())
	// METRICS_EXPORT
	data.AuthType = types.StringNull()
	data.Endpoint = types.StringNull()
	data.BasicToken = types.StringNull()
	data.ExporterType = types.StringNull()
	data.Username = types.StringNull()
	data.Headers = types.MapNull(types.StringType)
	data.Labels = types.MapNull(types.StringType)
	// Polymorphic
	data.Password = types.StringNull()
}

func (data *EnvironmentObject) populateConnection(ctx context.Context, conn *platform.EnvironmentObjectConnection, preserve *EnvironmentObjectPreserve) diag.Diagnostics {
	var diags diag.Diagnostics

	// connection_auth_type (Computed nested object)
	if conn.ConnectionAuthType != nil {
		data.ConnectionAuthType, diags = EnvironmentObjectConnectionAuthTypeTypesObject(ctx, conn.ConnectionAuthType)
		if diags.HasError() {
			return diags
		}
	} else {
		data.ConnectionAuthType = types.ObjectNull(schemas.EnvironmentObjectConnectionAuthTypeAttributeTypes())
	}

	// auth_type_id is write-only on the API: it's accepted on Create/Update but
	// never echoed back on GET. Preserve the user-supplied value when we have it;
	// otherwise stay null. Do NOT fall back to connection_auth_type.id from the
	// API — that would write a non-null value into an Optional-only field for
	// users who never set auth_type_id, breaking the plan-vs-state consistency
	// check. The resolved auth-type id is still readable via the already-Computed
	// connection_auth_type.id nested object.
	if preserve != nil && preserve.AuthTypeId != nil {
		data.AuthTypeId = types.StringValue(*preserve.AuthTypeId)
	} else {
		data.AuthTypeId = types.StringNull()
	}

	// extra — keep the caller's exact JSON string when provided to avoid
	// json.Marshal key reordering causing permadiffs.
	switch {
	case preserve != nil && preserve.Extra != nil:
		data.Extra = types.StringValue(*preserve.Extra)
	case conn.Extra != nil:
		extraBytes, err := json.Marshal(conn.Extra)
		if err != nil {
			return diag.Diagnostics{diag.NewErrorDiagnostic("Internal Error", fmt.Sprintf("Failed to marshal connection extra to JSON: %s", err))}
		}
		data.Extra = types.StringValue(string(extraBytes))
	default:
		data.Extra = types.StringNull()
	}

	if conn.Port != nil {
		data.Port = types.Int64Value(int64(*conn.Port))
	} else {
		data.Port = types.Int64Null()
	}

	data.Type = types.StringValue(conn.Type)
	data.Host = types.StringPointerValue(conn.Host)
	data.Schema = types.StringPointerValue(conn.Schema)
	data.Login = types.StringPointerValue(conn.Login)
	data.Password = preserveSecret(conn.Password, ifPreserve(preserve, func(p *EnvironmentObjectPreserve) *string { return p.Password }))

	return nil
}

func (data *EnvironmentObject) populateMetricsExport(me *platform.EnvironmentObjectMetricsExport, preserve *EnvironmentObjectPreserve) diag.Diagnostics {
	// auth_type is not echoed back on GET, so fall back to the user's prior
	// plan value when we have one (same pattern as auth_type_id for CONNECTION).
	switch {
	case me.AuthType != nil:
		data.AuthType = types.StringValue(string(*me.AuthType))
	case preserve != nil && preserve.MetricsExportAuthType != nil:
		data.AuthType = types.StringValue(*preserve.MetricsExportAuthType)
	default:
		data.AuthType = types.StringNull()
	}
	data.Endpoint = types.StringValue(me.Endpoint)
	data.ExporterType = types.StringValue(string(me.ExporterType))
	data.Username = types.StringPointerValue(me.Username)

	headers, diags := stringMapToTFMap(me.Headers)
	if diags.HasError() {
		return diags
	}
	data.Headers = headers
	labels, diags := stringMapToTFMap(me.Labels)
	if diags.HasError() {
		return diags
	}
	data.Labels = labels

	data.BasicToken = preserveSecret(me.BasicToken, ifPreserve(preserve, func(p *EnvironmentObjectPreserve) *string { return p.BasicToken }))
	data.Password = preserveSecret(me.Password, ifPreserve(preserve, func(p *EnvironmentObjectPreserve) *string { return p.Password }))
	return nil
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

// EnvironmentObjectLinkTypesObject converts a platform.EnvironmentObjectLink into a
// types.Object matching the link schema with a flat `overrides` block. The
// preserve argument supplies any per-link override secrets/extra JSON that the
// API does not echo back.
func EnvironmentObjectLinkTypesObject(ctx context.Context, link *platform.EnvironmentObjectLink, preserve *EnvironmentObjectLinkOverridePreserve) (types.Object, diag.Diagnostics) {
	overrides, diags := environmentObjectOverridesTypesObject(link, preserve)
	if diags.HasError() {
		return types.Object{}, diags
	}

	return types.ObjectValue(schemas.EnvironmentObjectLinkAttributeTypes(), map[string]attr.Value{
		"scope":           types.StringValue(string(link.Scope)),
		"scope_entity_id": types.StringValue(link.ScopeEntityId),
		"overrides":       overrides,
	})
}

// environmentObjectOverridesTypesObject builds the flat `overrides` value for
// a link by selecting fields from whichever of AirflowVariableOverrides /
// ConnectionOverrides / MetricsExportOverrides is non-nil on the API response.
func environmentObjectOverridesTypesObject(link *platform.EnvironmentObjectLink, preserve *EnvironmentObjectLinkOverridePreserve) (types.Object, diag.Diagnostics) {
	overridesAttrTypes := schemas.EnvironmentObjectOverridesAttributeTypes()

	hasAny := link.AirflowVariableOverrides != nil || link.ConnectionOverrides != nil || link.MetricsExportOverrides != nil
	if !hasAny {
		return types.ObjectNull(overridesAttrTypes), nil
	}

	// Start with everything null; populate per type.
	values := map[string]attr.Value{
		"value":         types.StringNull(),
		"type":          types.StringNull(),
		"host":          types.StringNull(),
		"port":          types.Int64Null(),
		"schema":        types.StringNull(),
		"login":         types.StringNull(),
		"extra":         types.StringNull(),
		"password":      types.StringNull(),
		"auth_type":     types.StringNull(),
		"endpoint":      types.StringNull(),
		"basic_token":   types.StringNull(),
		"exporter_type": types.StringNull(),
		"username":      types.StringNull(),
		"headers":       types.MapNull(types.StringType),
		"labels":        types.MapNull(types.StringType),
	}

	if link.AirflowVariableOverrides != nil {
		value := link.AirflowVariableOverrides.Value
		// EnvironmentObjectAirflowVariableOverrides has no IsSecret field, so we
		// can't mirror the top-level handler's `obj.AirflowVariable.IsSecret`
		// guard exactly. Approximate it: only fall back to the preserved plan
		// value when the API returned empty (the redaction sentinel). For
		// non-secret overrides the API returns the real value, which then wins —
		// so server-side edits surface as drift instead of being silently masked.
		if value == "" && preserve != nil && preserve.Value != nil {
			value = *preserve.Value
		}
		values["value"] = types.StringValue(value)
	}

	if link.ConnectionOverrides != nil {
		co := link.ConnectionOverrides
		switch {
		case preserve != nil && preserve.Extra != nil:
			values["extra"] = types.StringValue(*preserve.Extra)
		case co.Extra != nil:
			extraBytes, err := json.Marshal(co.Extra)
			if err != nil {
				return types.Object{}, diag.Diagnostics{diag.NewErrorDiagnostic("Internal Error", fmt.Sprintf("Failed to marshal connection overrides extra: %s", err))}
			}
			values["extra"] = types.StringValue(string(extraBytes))
		}
		if co.Port != nil {
			values["port"] = types.Int64Value(int64(*co.Port))
		}
		values["type"] = types.StringPointerValue(co.Type)
		values["host"] = types.StringPointerValue(co.Host)
		values["schema"] = types.StringPointerValue(co.Schema)
		values["login"] = types.StringPointerValue(co.Login)
		values["password"] = preserveSecretLink(co.Password, preserve, func(p *EnvironmentObjectLinkOverridePreserve) *string { return p.Password })
	}

	if link.MetricsExportOverrides != nil {
		mo := link.MetricsExportOverrides
		// auth_type is not echoed back on GET — fall back to preserved plan value
		switch {
		case mo.AuthType != nil:
			values["auth_type"] = types.StringValue(string(*mo.AuthType))
		case preserve != nil && preserve.AuthType != nil:
			values["auth_type"] = types.StringValue(*preserve.AuthType)
		}
		if mo.ExporterType != nil {
			values["exporter_type"] = types.StringValue(string(*mo.ExporterType))
		}
		headers, diags := stringMapToTFMap(mo.Headers)
		if diags.HasError() {
			return types.Object{}, diags
		}
		labels, diags := stringMapToTFMap(mo.Labels)
		if diags.HasError() {
			return types.Object{}, diags
		}
		values["endpoint"] = types.StringPointerValue(mo.Endpoint)
		values["username"] = types.StringPointerValue(mo.Username)
		values["headers"] = headers
		values["labels"] = labels
		values["basic_token"] = preserveSecretLink(mo.BasicToken, preserve, func(p *EnvironmentObjectLinkOverridePreserve) *string { return p.BasicToken })
		values["password"] = preserveSecretLink(mo.Password, preserve, func(p *EnvironmentObjectLinkOverridePreserve) *string { return p.Password })
	}

	return types.ObjectValue(overridesAttrTypes, values)
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

// preserveSecret returns a types.String for a sensitive field. The Astro API
// strips secrets on GET, returning either nil or an empty string as the
// redaction sentinel. When the caller has a preserved plan value, use it.
// Otherwise treat empty-string from the API as null — writing a non-null empty
// string into an Optional-only attribute (password / basic_token) would trip
// the framework's "Provider produced inconsistent result after apply" check
// for users who never set the field.
func preserveSecret(apiVal, preserveVal *string) types.String {
	if preserveVal != nil && (apiVal == nil || *apiVal == "") {
		return types.StringValue(*preserveVal)
	}
	if apiVal == nil || *apiVal == "" {
		return types.StringNull()
	}
	return types.StringValue(*apiVal)
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
