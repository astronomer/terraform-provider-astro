package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &environmentObjectResource{}
var _ resource.ResourceWithImportState = &environmentObjectResource{}
var _ resource.ResourceWithConfigure = &environmentObjectResource{}
var _ resource.ResourceWithValidateConfig = &environmentObjectResource{}

func NewEnvironmentObjectResource() resource.Resource {
	return &environmentObjectResource{}
}

// environmentObjectResource defines the resource implementation.
type environmentObjectResource struct {
	platformClient *platform.ClientWithResponses
	organizationId string
}

func (r *environmentObjectResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_environment_object"
}

func (r *environmentObjectResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Environment Object resource. Manages Airflow connections, variables, and metrics exports scoped to a Workspace or Deployment.",
		Attributes:          schemas.EnvironmentObjectResourceSchemaAttributes(),
	}
}

func (r *environmentObjectResource) Configure(
	ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	apiClients, ok := req.ProviderData.(models.ApiClientsModel)
	if !ok {
		utils.ResourceApiClientConfigureError(ctx, req, resp)
		return
	}

	r.platformClient = apiClients.PlatformClient
	r.organizationId = apiClients.OrganizationId
}

func (r *environmentObjectResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.EnvironmentObject

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	preserve, diags := buildPreserveFromModel(ctx, &data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	createReq, diags := buildCreateRequest(ctx, &data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	createResp, err := r.platformClient.CreateEnvironmentObjectWithResponse(ctx, r.organizationId, createReq)
	if err != nil {
		tflog.Error(ctx, "failed to create environment object", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create environment object, got error: %s", err))
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, createResp.HTTPResponse, createResp.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}
	if createResp.JSON200 == nil {
		tflog.Error(ctx, "failed to create environment object", map[string]interface{}{"error": "nil response"})
		resp.Diagnostics.AddError("Client Error", "Unable to create environment object, got nil response")
		return
	}

	// Persist the ID immediately so a failure in the follow-up GET leaves a
	// refreshable row in state instead of orphaning the just-created object.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), createResp.JSON200.Id)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create only returns the ID, do a follow-up GET to populate full state
	getResp, err := r.platformClient.GetEnvironmentObjectWithResponse(ctx, r.organizationId, createResp.JSON200.Id)
	if err != nil {
		tflog.Error(ctx, "failed to get environment object after create", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get environment object after create, got error: %s", err))
		return
	}
	_, diagnostic = clients.NormalizeAPIError(ctx, getResp.HTTPResponse, getResp.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}
	if getResp.JSON200 == nil {
		tflog.Error(ctx, "failed to get environment object after create", map[string]interface{}{"error": "nil response"})
		resp.Diagnostics.AddError("Client Error", "Unable to get environment object after create, got nil response")
		return
	}

	diags = data.ReadFromResponse(ctx, getResp.JSON200, preserve)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created an environment object resource: %v", data.Id.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *environmentObjectResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.EnvironmentObject

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	preserve, diags := buildPreserveFromModel(ctx, &data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	envObj, err := r.platformClient.GetEnvironmentObjectWithResponse(ctx, r.organizationId, data.Id.ValueString())
	if err != nil {
		tflog.Error(ctx, "failed to get environment object", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get environment object, got error: %s", err))
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, envObj.HTTPResponse, envObj.Body)
	if statusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}
	if envObj.JSON200 == nil {
		tflog.Error(ctx, "failed to get environment object", map[string]interface{}{"error": "nil response"})
		resp.Diagnostics.AddError("Client Error", "Unable to get environment object, got nil response")
		return
	}

	diags = data.ReadFromResponse(ctx, envObj.JSON200, preserve)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read an environment object resource: %v", data.Id.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *environmentObjectResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data models.EnvironmentObject

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	preserve, diags := buildPreserveFromModel(ctx, &data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	updateReq, diags := buildUpdateRequest(ctx, &data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	updateResp, err := r.platformClient.UpdateEnvironmentObjectWithResponse(ctx, r.organizationId, data.Id.ValueString(), updateReq)
	if err != nil {
		tflog.Error(ctx, "failed to update environment object", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update environment object, got error: %s", err))
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, updateResp.HTTPResponse, updateResp.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	getResp, err := r.platformClient.GetEnvironmentObjectWithResponse(ctx, r.organizationId, data.Id.ValueString())
	if err != nil {
		tflog.Error(ctx, "failed to get environment object after update", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get environment object after update, got error: %s", err))
		return
	}
	_, diagnostic = clients.NormalizeAPIError(ctx, getResp.HTTPResponse, getResp.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}
	if getResp.JSON200 == nil {
		tflog.Error(ctx, "failed to get environment object after update", map[string]interface{}{"error": "nil response"})
		resp.Diagnostics.AddError("Client Error", "Unable to get environment object after update, got nil response")
		return
	}

	diags = data.ReadFromResponse(ctx, getResp.JSON200, preserve)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated an environment object resource: %v", data.Id.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *environmentObjectResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.EnvironmentObject

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	envObj, err := r.platformClient.DeleteEnvironmentObjectWithResponse(ctx, r.organizationId, data.Id.ValueString())
	if err != nil {
		tflog.Error(ctx, "failed to delete environment object", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete environment object, got error: %s", err))
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, envObj.HTTPResponse, envObj.Body)
	if statusCode != http.StatusNotFound && diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted an environment object resource: %v", data.Id.ValueString()))
}

func (r *environmentObjectResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// --- Request builders ---

func buildCreateRequest(ctx context.Context, data *models.EnvironmentObject) (platform.CreateEnvironmentObjectJSONRequestBody, diag.Diagnostics) {
	req := platform.CreateEnvironmentObjectRequest{
		ObjectKey:           data.ObjectKey.ValueString(),
		ObjectType:          platform.CreateEnvironmentObjectRequestObjectType(data.ObjectType.ValueString()),
		Scope:               platform.CreateEnvironmentObjectRequestScope(data.Scope.ValueString()),
		ScopeEntityId:       data.ScopeEntityId.ValueString(),
		AutoLinkDeployments: data.AutoLinkDeployments.ValueBoolPointer(),
	}

	switch platform.CreateEnvironmentObjectRequestObjectType(data.ObjectType.ValueString()) {
	case platform.CreateEnvironmentObjectRequestObjectTypeAIRFLOWVARIABLE:
		req.AirflowVariable = &platform.CreateEnvironmentObjectAirflowVariableRequest{
			Value:    data.Value.ValueStringPointer(),
			IsSecret: data.IsSecret.ValueBoolPointer(),
		}
	case platform.CreateEnvironmentObjectRequestObjectTypeCONNECTION:
		connReq := &platform.CreateEnvironmentObjectConnectionRequest{
			Type:       data.Type.ValueString(),
			Host:       data.Host.ValueStringPointer(),
			Login:      data.Login.ValueStringPointer(),
			Password:   data.Password.ValueStringPointer(),
			Schema:     data.Schema.ValueStringPointer(),
			AuthTypeId: data.AuthTypeId.ValueStringPointer(),
		}
		if !data.Port.IsNull() && !data.Port.IsUnknown() {
			connReq.Port = lo.ToPtr(int(data.Port.ValueInt64()))
		}
		if !data.Extra.IsNull() && !data.Extra.IsUnknown() {
			var extra map[string]interface{}
			if err := json.Unmarshal([]byte(data.Extra.ValueString()), &extra); err != nil {
				return req, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid Input", fmt.Sprintf("extra must be valid JSON: %s", err))}
			}
			connReq.Extra = &extra
		}
		req.Connection = connReq
	case platform.CreateEnvironmentObjectRequestObjectTypeMETRICSEXPORT:
		meReq := &platform.CreateEnvironmentObjectMetricsExportRequest{
			Endpoint:     data.Endpoint.ValueString(),
			ExporterType: platform.CreateEnvironmentObjectMetricsExportRequestExporterType(data.ExporterType.ValueString()),
			BasicToken:   data.BasicToken.ValueStringPointer(),
			Username:     data.Username.ValueStringPointer(),
			Password:     data.Password.ValueStringPointer(),
		}
		if !data.AuthType.IsNull() && !data.AuthType.IsUnknown() {
			meReq.AuthType = lo.ToPtr(platform.CreateEnvironmentObjectMetricsExportRequestAuthType(data.AuthType.ValueString()))
		}
		if !data.Headers.IsNull() && !data.Headers.IsUnknown() {
			h := tfMapToStringMap(data.Headers)
			meReq.Headers = &h
		}
		if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
			l := tfMapToStringMap(data.Labels)
			meReq.Labels = &l
		}
		req.MetricsExport = meReq
	}

	if d := buildCreateLinks(ctx, data, &req); d.HasError() {
		return req, d
	}
	if d := buildCreateExcludeLinks(ctx, data, &req); d.HasError() {
		return req, d
	}

	return req, nil
}

func buildCreateLinks(ctx context.Context, data *models.EnvironmentObject, req *platform.CreateEnvironmentObjectRequest) diag.Diagnostics {
	if data.Links.IsNull() || data.Links.IsUnknown() {
		return nil
	}
	var linkInputs []models.EnvironmentObjectLinkInput
	if d := data.Links.ElementsAs(ctx, &linkInputs, false); d.HasError() {
		return d
	}
	createLinks := make([]platform.CreateEnvironmentObjectLinkRequest, len(linkInputs))
	objectType := platform.CreateEnvironmentObjectRequestObjectType(data.ObjectType.ValueString())
	for i, li := range linkInputs {
		createLinks[i] = platform.CreateEnvironmentObjectLinkRequest{
			Scope:         platform.CreateEnvironmentObjectLinkRequestScope(li.Scope.ValueString()),
			ScopeEntityId: li.ScopeEntityId.ValueString(),
		}
		overrides, d := buildCreateOverrides(ctx, li.Overrides, objectType)
		if d.HasError() {
			return d
		}
		if overrides != nil {
			createLinks[i].Overrides = overrides
		}
	}
	req.Links = &createLinks
	return nil
}

func buildCreateExcludeLinks(ctx context.Context, data *models.EnvironmentObject, req *platform.CreateEnvironmentObjectRequest) diag.Diagnostics {
	if data.ExcludeLinks.IsNull() || data.ExcludeLinks.IsUnknown() {
		return nil
	}
	var elInputs []models.EnvironmentObjectExcludeLinkInput
	if d := data.ExcludeLinks.ElementsAs(ctx, &elInputs, false); d.HasError() {
		return d
	}
	excludeLinks := make([]platform.ExcludeLinkEnvironmentObjectRequest, len(elInputs))
	for i, el := range elInputs {
		excludeLinks[i] = platform.ExcludeLinkEnvironmentObjectRequest{
			Scope:         platform.ExcludeLinkEnvironmentObjectRequestScope(el.Scope.ValueString()),
			ScopeEntityId: el.ScopeEntityId.ValueString(),
		}
	}
	req.ExcludeLinks = &excludeLinks
	return nil
}

func buildUpdateRequest(ctx context.Context, data *models.EnvironmentObject) (platform.UpdateEnvironmentObjectJSONRequestBody, diag.Diagnostics) {
	req := platform.UpdateEnvironmentObjectRequest{
		AutoLinkDeployments: data.AutoLinkDeployments.ValueBoolPointer(),
	}

	switch platform.CreateEnvironmentObjectRequestObjectType(data.ObjectType.ValueString()) {
	case platform.CreateEnvironmentObjectRequestObjectTypeAIRFLOWVARIABLE:
		req.AirflowVariable = &platform.UpdateEnvironmentObjectAirflowVariableRequest{
			Value: data.Value.ValueStringPointer(),
		}
	case platform.CreateEnvironmentObjectRequestObjectTypeCONNECTION:
		connReq := &platform.UpdateEnvironmentObjectConnectionRequest{
			Type:       data.Type.ValueString(),
			Host:       data.Host.ValueStringPointer(),
			Login:      data.Login.ValueStringPointer(),
			Password:   data.Password.ValueStringPointer(),
			Schema:     data.Schema.ValueStringPointer(),
			AuthTypeId: data.AuthTypeId.ValueStringPointer(),
		}
		if !data.Port.IsNull() && !data.Port.IsUnknown() {
			connReq.Port = lo.ToPtr(int(data.Port.ValueInt64()))
		}
		if !data.Extra.IsNull() && !data.Extra.IsUnknown() {
			var extra map[string]interface{}
			if err := json.Unmarshal([]byte(data.Extra.ValueString()), &extra); err != nil {
				return req, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid Input", fmt.Sprintf("extra must be valid JSON: %s", err))}
			}
			connReq.Extra = &extra
		}
		req.Connection = connReq
	case platform.CreateEnvironmentObjectRequestObjectTypeMETRICSEXPORT:
		meReq := &platform.UpdateEnvironmentObjectMetricsExportRequest{
			Endpoint:   data.Endpoint.ValueStringPointer(),
			BasicToken: data.BasicToken.ValueStringPointer(),
			Username:   data.Username.ValueStringPointer(),
			Password:   data.Password.ValueStringPointer(),
		}
		if !data.ExporterType.IsNull() && !data.ExporterType.IsUnknown() {
			meReq.ExporterType = lo.ToPtr(platform.UpdateEnvironmentObjectMetricsExportRequestExporterType(data.ExporterType.ValueString()))
		}
		if !data.AuthType.IsNull() && !data.AuthType.IsUnknown() {
			meReq.AuthType = lo.ToPtr(platform.UpdateEnvironmentObjectMetricsExportRequestAuthType(data.AuthType.ValueString()))
		}
		if !data.Headers.IsNull() && !data.Headers.IsUnknown() {
			h := tfMapToStringMap(data.Headers)
			meReq.Headers = &h
		}
		if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
			l := tfMapToStringMap(data.Labels)
			meReq.Labels = &l
		}
		req.MetricsExport = meReq
	}

	if d := buildUpdateLinks(ctx, data, &req); d.HasError() {
		return req, d
	}
	if d := buildUpdateExcludeLinks(ctx, data, &req); d.HasError() {
		return req, d
	}

	return req, nil
}

func buildUpdateLinks(ctx context.Context, data *models.EnvironmentObject, req *platform.UpdateEnvironmentObjectRequest) diag.Diagnostics {
	if data.Links.IsNull() || data.Links.IsUnknown() {
		return nil
	}
	var linkInputs []models.EnvironmentObjectLinkInput
	if d := data.Links.ElementsAs(ctx, &linkInputs, false); d.HasError() {
		return d
	}
	updateLinks := make([]platform.UpdateEnvironmentObjectLinkRequest, len(linkInputs))
	objectType := platform.CreateEnvironmentObjectRequestObjectType(data.ObjectType.ValueString())
	for i, li := range linkInputs {
		updateLinks[i] = platform.UpdateEnvironmentObjectLinkRequest{
			Scope:         platform.UpdateEnvironmentObjectLinkRequestScope(li.Scope.ValueString()),
			ScopeEntityId: li.ScopeEntityId.ValueString(),
		}
		overrides, d := buildUpdateOverrides(ctx, li.Overrides, objectType)
		if d.HasError() {
			return d
		}
		if overrides != nil {
			updateLinks[i].Overrides = overrides
		}
	}
	req.Links = &updateLinks
	return nil
}

func buildUpdateExcludeLinks(ctx context.Context, data *models.EnvironmentObject, req *platform.UpdateEnvironmentObjectRequest) diag.Diagnostics {
	if data.ExcludeLinks.IsNull() || data.ExcludeLinks.IsUnknown() {
		return nil
	}
	var elInputs []models.EnvironmentObjectExcludeLinkInput
	if d := data.ExcludeLinks.ElementsAs(ctx, &elInputs, false); d.HasError() {
		return d
	}
	excludeLinks := make([]platform.ExcludeLinkEnvironmentObjectRequest, len(elInputs))
	for i, el := range elInputs {
		excludeLinks[i] = platform.ExcludeLinkEnvironmentObjectRequest{
			Scope:         platform.ExcludeLinkEnvironmentObjectRequestScope(el.Scope.ValueString()),
			ScopeEntityId: el.ScopeEntityId.ValueString(),
		}
	}
	req.ExcludeLinks = &excludeLinks
	return nil
}

// buildCreateOverrides unpacks the flat per-link `overrides` block into the
// API's CreateEnvironmentObjectOverridesRequest, picking the right sub-struct
// based on the parent object_type. Returns nil when overrides is null/unknown.
func buildCreateOverrides(ctx context.Context, overrides types.Object, objectType platform.CreateEnvironmentObjectRequestObjectType) (*platform.CreateEnvironmentObjectOverridesRequest, diag.Diagnostics) {
	if overrides.IsNull() || overrides.IsUnknown() {
		return nil, nil
	}
	var ov models.EnvironmentObjectOverridesInput
	if d := overrides.As(ctx, &ov, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); d.HasError() {
		return nil, d
	}

	out := &platform.CreateEnvironmentObjectOverridesRequest{}

	switch objectType {
	case platform.CreateEnvironmentObjectRequestObjectTypeAIRFLOWVARIABLE:
		if ov.Value.IsNull() {
			return nil, nil
		}
		out.AirflowVariable = &platform.CreateEnvironmentObjectAirflowVariableOverridesRequest{
			Value: ov.Value.ValueStringPointer(),
		}
	case platform.CreateEnvironmentObjectRequestObjectTypeCONNECTION:
		connOvr := &platform.CreateEnvironmentObjectConnectionOverridesRequest{
			Type:     ov.Type.ValueStringPointer(),
			Host:     ov.Host.ValueStringPointer(),
			Login:    ov.Login.ValueStringPointer(),
			Password: ov.Password.ValueStringPointer(),
			Schema:   ov.Schema.ValueStringPointer(),
		}
		if !ov.Port.IsNull() && !ov.Port.IsUnknown() {
			connOvr.Port = lo.ToPtr(int(ov.Port.ValueInt64()))
		}
		if !ov.Extra.IsNull() && !ov.Extra.IsUnknown() {
			var extra map[string]interface{}
			if err := json.Unmarshal([]byte(ov.Extra.ValueString()), &extra); err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid Input", fmt.Sprintf("overrides.extra must be valid JSON: %s", err))}
			}
			connOvr.Extra = &extra
		}
		out.Connection = connOvr
	case platform.CreateEnvironmentObjectRequestObjectTypeMETRICSEXPORT:
		meOvr := &platform.CreateEnvironmentObjectMetricsExportOverridesRequest{
			Endpoint:   ov.Endpoint.ValueStringPointer(),
			BasicToken: ov.BasicToken.ValueStringPointer(),
			Username:   ov.Username.ValueStringPointer(),
			Password:   ov.Password.ValueStringPointer(),
		}
		if !ov.AuthType.IsNull() && !ov.AuthType.IsUnknown() {
			meOvr.AuthType = lo.ToPtr(platform.CreateEnvironmentObjectMetricsExportOverridesRequestAuthType(ov.AuthType.ValueString()))
		}
		if !ov.ExporterType.IsNull() && !ov.ExporterType.IsUnknown() {
			meOvr.ExporterType = lo.ToPtr(platform.CreateEnvironmentObjectMetricsExportOverridesRequestExporterType(ov.ExporterType.ValueString()))
		}
		if !ov.Headers.IsNull() && !ov.Headers.IsUnknown() {
			h := tfMapToStringMap(ov.Headers)
			meOvr.Headers = &h
		}
		if !ov.Labels.IsNull() && !ov.Labels.IsUnknown() {
			l := tfMapToStringMap(ov.Labels)
			meOvr.Labels = &l
		}
		out.MetricsExport = meOvr
	}

	return out, nil
}

// buildUpdateOverrides is the Update counterpart to buildCreateOverrides.
func buildUpdateOverrides(ctx context.Context, overrides types.Object, objectType platform.CreateEnvironmentObjectRequestObjectType) (*platform.UpdateEnvironmentObjectOverridesRequest, diag.Diagnostics) {
	if overrides.IsNull() || overrides.IsUnknown() {
		return nil, nil
	}
	var ov models.EnvironmentObjectOverridesInput
	if d := overrides.As(ctx, &ov, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); d.HasError() {
		return nil, d
	}

	out := &platform.UpdateEnvironmentObjectOverridesRequest{}

	switch objectType {
	case platform.CreateEnvironmentObjectRequestObjectTypeAIRFLOWVARIABLE:
		// Always emit the AirflowVariable sub-struct so the user can clear a
		// previously-set value override (Value=nil → API clears). The Create
		// counterpart still early-returns nil because "create with no value"
		// is unambiguous.
		out.AirflowVariable = &platform.UpdateEnvironmentObjectAirflowVariableOverridesRequest{
			Value: ov.Value.ValueStringPointer(),
		}
	case platform.CreateEnvironmentObjectRequestObjectTypeCONNECTION:
		connOvr := &platform.UpdateEnvironmentObjectConnectionOverridesRequest{
			Type:     ov.Type.ValueStringPointer(),
			Host:     ov.Host.ValueStringPointer(),
			Login:    ov.Login.ValueStringPointer(),
			Password: ov.Password.ValueStringPointer(),
			Schema:   ov.Schema.ValueStringPointer(),
		}
		if !ov.Port.IsNull() && !ov.Port.IsUnknown() {
			connOvr.Port = lo.ToPtr(int(ov.Port.ValueInt64()))
		}
		if !ov.Extra.IsNull() && !ov.Extra.IsUnknown() {
			var extra map[string]interface{}
			if err := json.Unmarshal([]byte(ov.Extra.ValueString()), &extra); err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid Input", fmt.Sprintf("overrides.extra must be valid JSON: %s", err))}
			}
			connOvr.Extra = &extra
		}
		out.Connection = connOvr
	case platform.CreateEnvironmentObjectRequestObjectTypeMETRICSEXPORT:
		meOvr := &platform.UpdateEnvironmentObjectMetricsExportOverridesRequest{
			Endpoint:   ov.Endpoint.ValueStringPointer(),
			BasicToken: ov.BasicToken.ValueStringPointer(),
			Username:   ov.Username.ValueStringPointer(),
			Password:   ov.Password.ValueStringPointer(),
		}
		if !ov.AuthType.IsNull() && !ov.AuthType.IsUnknown() {
			meOvr.AuthType = lo.ToPtr(platform.UpdateEnvironmentObjectMetricsExportOverridesRequestAuthType(ov.AuthType.ValueString()))
		}
		if !ov.ExporterType.IsNull() && !ov.ExporterType.IsUnknown() {
			meOvr.ExporterType = lo.ToPtr(platform.UpdateEnvironmentObjectMetricsExportOverridesRequestExporterType(ov.ExporterType.ValueString()))
		}
		if !ov.Headers.IsNull() && !ov.Headers.IsUnknown() {
			h := tfMapToStringMap(ov.Headers)
			meOvr.Headers = &h
		}
		if !ov.Labels.IsNull() && !ov.Labels.IsUnknown() {
			l := tfMapToStringMap(ov.Labels)
			meOvr.Labels = &l
		}
		out.MetricsExport = meOvr
	}

	return out, nil
}

func tfMapToStringMap(m types.Map) map[string]string {
	result := make(map[string]string, len(m.Elements()))
	for k, v := range m.Elements() {
		sv, ok := v.(types.String)
		if !ok || sv.IsNull() || sv.IsUnknown() {
			// Schema declares ElementType=StringType so the framework normally
			// blocks non-string values; skip null/unknown so they don't get
			// silently coerced to "" and sent to the API as empty values.
			continue
		}
		result[k] = sv.ValueString()
	}
	return result
}

// buildPreserveFromModel pulls every field the API does not echo back on GET
// — sensitive fields plus the user's exact `extra` JSON string — directly off
// the flat model. Used to repopulate state without losing user input on refresh.
func buildPreserveFromModel(ctx context.Context, data *models.EnvironmentObject) (*models.EnvironmentObjectPreserve, diag.Diagnostics) {
	preserve := &models.EnvironmentObjectPreserve{}

	switch platform.CreateEnvironmentObjectRequestObjectType(data.ObjectType.ValueString()) {
	case platform.CreateEnvironmentObjectRequestObjectTypeAIRFLOWVARIABLE:
		// Preserve value when secret (API returns empty for secrets) OR when caller
		// supplied a value (handles is_secret toggle edge cases).
		if data.IsSecret.ValueBool() || !data.Value.IsNull() {
			preserve.AirflowVariableValue = data.Value.ValueStringPointer()
		}
	case platform.CreateEnvironmentObjectRequestObjectTypeCONNECTION:
		preserve.Password = data.Password.ValueStringPointer()
		preserve.AuthTypeId = data.AuthTypeId.ValueStringPointer()
		preserve.Extra = data.Extra.ValueStringPointer()
	case platform.CreateEnvironmentObjectRequestObjectTypeMETRICSEXPORT:
		preserve.Password = data.Password.ValueStringPointer()
		preserve.BasicToken = data.BasicToken.ValueStringPointer()
		preserve.MetricsExportAuthType = data.AuthType.ValueStringPointer()
	}

	// Per-link overrides
	if !data.Links.IsNull() && !data.Links.IsUnknown() {
		var linkInputs []models.EnvironmentObjectLinkInput
		if d := data.Links.ElementsAs(ctx, &linkInputs, false); d.HasError() {
			return nil, d
		}
		if len(linkInputs) > 0 {
			preserve.LinkOverrides = make(map[string]*models.EnvironmentObjectLinkOverridePreserve, len(linkInputs))
			for _, li := range linkInputs {
				lop, d := extractLinkOverridePreserve(ctx, li.Overrides)
				if d.HasError() {
					return nil, d
				}
				preserve.LinkOverrides[models.LinkPreserveKey(li.Scope.ValueString(), li.ScopeEntityId.ValueString())] = lop
			}
		}
	}

	return preserve, nil
}

func extractLinkOverridePreserve(ctx context.Context, overrides types.Object) (*models.EnvironmentObjectLinkOverridePreserve, diag.Diagnostics) {
	lop := &models.EnvironmentObjectLinkOverridePreserve{}
	if overrides.IsNull() || overrides.IsUnknown() {
		return lop, nil
	}
	var ov models.EnvironmentObjectOverridesInput
	if d := overrides.As(ctx, &ov, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); d.HasError() {
		return nil, d
	}
	lop.Value = ov.Value.ValueStringPointer()
	lop.Password = ov.Password.ValueStringPointer()
	lop.Extra = ov.Extra.ValueStringPointer()
	lop.BasicToken = ov.BasicToken.ValueStringPointer()
	lop.AuthType = ov.AuthType.ValueStringPointer()
	return lop, nil
}

// --- ValidateConfig ---

// ValidateConfig enforces the field-level invariants that the flat schema
// can't express in attribute metadata: each type-specific field is only valid
// for its associated object_type, and certain fields are required for their
// type. Mirrors the per-field gating used in
// internal/provider/validators/notification_channel_definition_validator.go.
func (r *environmentObjectResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var data models.EnvironmentObject

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip type-level validation when object_type is unknown (rare —
	// typically interpolated from another resource).
	if !data.ObjectType.IsUnknown() && !data.ObjectType.IsNull() {
		switch platform.CreateEnvironmentObjectRequestObjectType(data.ObjectType.ValueString()) {
		case platform.CreateEnvironmentObjectRequestObjectTypeAIRFLOWVARIABLE:
			resp.Diagnostics.Append(validateAirflowVariableFields(&data)...)
		case platform.CreateEnvironmentObjectRequestObjectTypeCONNECTION:
			resp.Diagnostics.Append(validateConnectionFields(&data)...)
		case platform.CreateEnvironmentObjectRequestObjectTypeMETRICSEXPORT:
			resp.Diagnostics.Append(validateMetricsExportFields(&data)...)
		}
	}

	// scope=DEPLOYMENT can't carry workspace-only attributes.
	if !data.Scope.IsUnknown() && !data.Scope.IsNull() &&
		data.Scope.ValueString() == string(platform.CreateEnvironmentObjectRequestScopeDEPLOYMENT) {
		if !data.AutoLinkDeployments.IsNull() && !data.AutoLinkDeployments.IsUnknown() {
			resp.Diagnostics.AddAttributeError(path.Root("auto_link_deployments"),
				"Conflicting attribute",
				"auto_link_deployments is only valid when scope=WORKSPACE")
		}
		if !data.Links.IsNull() && !data.Links.IsUnknown() {
			resp.Diagnostics.AddAttributeError(path.Root("links"),
				"Conflicting attribute",
				"links is only valid when scope=WORKSPACE")
		}
		if !data.ExcludeLinks.IsNull() && !data.ExcludeLinks.IsUnknown() {
			resp.Diagnostics.AddAttributeError(path.Root("exclude_links"),
				"Conflicting attribute",
				"exclude_links is only valid when scope=WORKSPACE")
		}
	}
}

// validateAirflowVariableFields enforces: `value` is required, and connection
// + metrics_export fields are absent. `is_secret` defaults to false via the API.
func validateAirflowVariableFields(data *models.EnvironmentObject) diag.Diagnostics {
	var diags diag.Diagnostics
	if data.Value.IsNull() && !data.Value.IsUnknown() {
		diags.AddAttributeError(path.Root("value"), "Missing required field",
			"value is required when object_type=AIRFLOW_VARIABLE")
	}
	for _, f := range connectionOnlyFields(data) {
		diags.AddAttributeError(path.Root(f.name), "Conflicting field",
			fmt.Sprintf("%s is only valid when object_type=CONNECTION", f.name))
	}
	for _, f := range metricsExportOnlyFields(data) {
		diags.AddAttributeError(path.Root(f.name), "Conflicting field",
			fmt.Sprintf("%s is only valid when object_type=METRICS_EXPORT", f.name))
	}
	return diags
}

// validateConnectionFields enforces: airflow_variable + metrics_export fields
// are absent, `type` is required, and `extra` parses as JSON object at plan time
// (otherwise it would fail mid-apply when the request builder calls
// json.Unmarshal).
func validateConnectionFields(data *models.EnvironmentObject) diag.Diagnostics {
	var diags diag.Diagnostics
	if data.Type.IsNull() && !data.Type.IsUnknown() {
		diags.AddAttributeError(path.Root("type"), "Missing required field",
			"type is required when object_type=CONNECTION")
	}
	if isUserSet(data.Extra) {
		var probe map[string]interface{}
		if err := json.Unmarshal([]byte(data.Extra.ValueString()), &probe); err != nil {
			diags.AddAttributeError(path.Root("extra"), "Invalid extra JSON",
				fmt.Sprintf("extra must be a JSON object string (use jsonencode({...})). Parse error: %s", err))
		}
	}
	for _, f := range airflowVariableOnlyFields(data) {
		diags.AddAttributeError(path.Root(f.name), "Conflicting field",
			fmt.Sprintf("%s is only valid when object_type=AIRFLOW_VARIABLE", f.name))
	}
	for _, f := range metricsExportOnlyFields(data) {
		diags.AddAttributeError(path.Root(f.name), "Conflicting field",
			fmt.Sprintf("%s is only valid when object_type=METRICS_EXPORT", f.name))
	}
	return diags
}

// validateMetricsExportFields enforces: airflow_variable + connection fields
// are absent, and `endpoint` + `exporter_type` are required.
func validateMetricsExportFields(data *models.EnvironmentObject) diag.Diagnostics {
	var diags diag.Diagnostics
	if data.Endpoint.IsNull() && !data.Endpoint.IsUnknown() {
		diags.AddAttributeError(path.Root("endpoint"), "Missing required field",
			"endpoint is required when object_type=METRICS_EXPORT")
	}
	if data.ExporterType.IsNull() && !data.ExporterType.IsUnknown() {
		diags.AddAttributeError(path.Root("exporter_type"), "Missing required field",
			"exporter_type is required when object_type=METRICS_EXPORT")
	}
	for _, f := range airflowVariableOnlyFields(data) {
		diags.AddAttributeError(path.Root(f.name), "Conflicting field",
			fmt.Sprintf("%s is only valid when object_type=AIRFLOW_VARIABLE", f.name))
	}
	for _, f := range connectionOnlyFields(data) {
		diags.AddAttributeError(path.Root(f.name), "Conflicting field",
			fmt.Sprintf("%s is only valid when object_type=CONNECTION", f.name))
	}
	return diags
}

// namedField is a tiny helper for the field-presence loops.
type namedField struct {
	name string
	set  bool
}

func airflowVariableOnlyFields(data *models.EnvironmentObject) []namedField {
	var out []namedField
	if isUserSet(data.Value) {
		out = append(out, namedField{name: "value", set: true})
	}
	// is_secret defaults to false / Computed — only flag if explicitly true.
	// We can't distinguish "user wrote is_secret=false" from "Computed default",
	// so we only flag when set to true.
	if data.IsSecret.ValueBool() {
		out = append(out, namedField{name: "is_secret", set: true})
	}
	return out
}

func connectionOnlyFields(data *models.EnvironmentObject) []namedField {
	checks := []struct {
		name string
		v    types.String
	}{
		{"type", data.Type},
		{"host", data.Host},
		{"schema", data.Schema},
		{"login", data.Login},
		{"extra", data.Extra},
		{"auth_type_id", data.AuthTypeId},
	}
	var out []namedField
	for _, c := range checks {
		if isUserSet(c.v) {
			out = append(out, namedField{name: c.name, set: true})
		}
	}
	if !data.Port.IsNull() && !data.Port.IsUnknown() {
		out = append(out, namedField{name: "port", set: true})
	}
	return out
}

func metricsExportOnlyFields(data *models.EnvironmentObject) []namedField {
	checks := []struct {
		name string
		v    types.String
	}{
		{"auth_type", data.AuthType},
		{"endpoint", data.Endpoint},
		{"basic_token", data.BasicToken},
		{"exporter_type", data.ExporterType},
		{"username", data.Username},
	}
	var out []namedField
	for _, c := range checks {
		if isUserSet(c.v) {
			out = append(out, namedField{name: c.name, set: true})
		}
	}
	if !data.Headers.IsNull() && !data.Headers.IsUnknown() {
		out = append(out, namedField{name: "headers", set: true})
	}
	if !data.Labels.IsNull() && !data.Labels.IsUnknown() {
		out = append(out, namedField{name: "labels", set: true})
	}
	return out
}

// isUserSet returns true when the user explicitly set a Computed-eligible
// String attribute (i.e. it's neither null nor unknown).
func isUserSet(v types.String) bool {
	return !v.IsNull() && !v.IsUnknown()
}
