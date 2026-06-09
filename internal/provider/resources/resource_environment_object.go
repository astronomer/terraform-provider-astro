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

	// Build preserve struct from plan: captures sensitive fields and the user's exact
	// `extra` JSON string so state stays consistent after the API strips/normalizes them.
	preserve, diags := buildPreserveFromModel(ctx, &data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// create request
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

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *environmentObjectResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.EnvironmentObject

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build preserve struct from prior state so sensitive fields and the user's `extra`
	// JSON string survive the refresh (the API returns null/empty/reordered for these).
	preserve, diags := buildPreserveFromModel(ctx, &data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// get request
	envObj, err := r.platformClient.GetEnvironmentObjectWithResponse(ctx, r.organizationId, data.Id.ValueString())
	if err != nil {
		tflog.Error(ctx, "failed to get environment object", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get environment object, got error: %s", err))
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, envObj.HTTPResponse, envObj.Body)
	// If the resource no longer exists, it is recommended to ignore the errors
	// and call RemoveResource to remove the resource from the state. The next Terraform plan will recreate the resource.
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

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *environmentObjectResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data models.EnvironmentObject

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build preserve struct from plan: captures sensitive fields and the user's
	// `extra` JSON string for the round-trip.
	preserve, diags := buildPreserveFromModel(ctx, &data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// update request
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

	// Follow-up GET to ensure full state including created_by
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

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *environmentObjectResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.EnvironmentObject

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// delete request
	envObj, err := r.platformClient.DeleteEnvironmentObjectWithResponse(ctx, r.organizationId, data.Id.ValueString())
	if err != nil {
		tflog.Error(ctx, "failed to delete environment object", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete environment object, got error: %s", err))
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, envObj.HTTPResponse, envObj.Body)
	// It is recommended to ignore 404 Resource Not Found errors when deleting a resource
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

func buildCreateRequest(ctx context.Context, data *models.EnvironmentObject) (platform.CreateEnvironmentObjectJSONRequestBody, diag.Diagnostics) {
	req := platform.CreateEnvironmentObjectRequest{
		ObjectKey:           data.ObjectKey.ValueString(),
		ObjectType:          platform.CreateEnvironmentObjectRequestObjectType(data.ObjectType.ValueString()),
		Scope:               platform.CreateEnvironmentObjectRequestScope(data.Scope.ValueString()),
		ScopeEntityId:       data.ScopeEntityId.ValueString(),
		AutoLinkDeployments: data.AutoLinkDeployments.ValueBoolPointer(),
	}

	var diags diag.Diagnostics

	// Airflow Variable
	if !data.AirflowVariable.IsNull() && !data.AirflowVariable.IsUnknown() {
		var av models.EnvironmentObjectAirflowVariableInput
		diags = data.AirflowVariable.As(ctx, &av, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return req, diags
		}
		req.AirflowVariable = &platform.CreateEnvironmentObjectAirflowVariableRequest{
			Value:    av.Value.ValueStringPointer(),
			IsSecret: av.IsSecret.ValueBoolPointer(),
		}
	}

	// Airflow Connection
	if !data.AirflowConnection.IsNull() && !data.AirflowConnection.IsUnknown() {
		var ci models.EnvironmentObjectAirflowConnectionInput
		diags = data.AirflowConnection.As(ctx, &ci, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return req, diags
		}
		connReq := &platform.CreateEnvironmentObjectConnectionRequest{
			Type:       ci.Type.ValueString(),
			Host:       ci.Host.ValueStringPointer(),
			Login:      ci.Login.ValueStringPointer(),
			Password:   ci.Password.ValueStringPointer(),
			Schema:     ci.Schema.ValueStringPointer(),
			AuthTypeId: ci.AuthTypeId.ValueStringPointer(),
		}
		if !ci.Port.IsNull() && !ci.Port.IsUnknown() {
			connReq.Port = lo.ToPtr(int(ci.Port.ValueInt64()))
		}
		if !ci.Extra.IsNull() && !ci.Extra.IsUnknown() {
			var extra map[string]interface{}
			if err := json.Unmarshal([]byte(ci.Extra.ValueString()), &extra); err != nil {
				return req, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid Input", fmt.Sprintf("airflow_connection.extra must be valid JSON: %s", err))}
			}
			connReq.Extra = &extra
		}
		req.Connection = connReq
	}

	// Metrics Export
	if !data.MetricsExport.IsNull() && !data.MetricsExport.IsUnknown() {
		var me models.EnvironmentObjectMetricsExportInput
		diags = data.MetricsExport.As(ctx, &me, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return req, diags
		}
		meReq := &platform.CreateEnvironmentObjectMetricsExportRequest{
			Endpoint:     me.Endpoint.ValueString(),
			ExporterType: platform.CreateEnvironmentObjectMetricsExportRequestExporterType(me.ExporterType.ValueString()),
			BasicToken:   me.BasicToken.ValueStringPointer(),
			Username:     me.Username.ValueStringPointer(),
			Password:     me.Password.ValueStringPointer(),
		}
		if !me.AuthType.IsNull() && !me.AuthType.IsUnknown() {
			meReq.AuthType = lo.ToPtr(platform.CreateEnvironmentObjectMetricsExportRequestAuthType(me.AuthType.ValueString()))
		}
		if !me.Headers.IsNull() && !me.Headers.IsUnknown() {
			h := tfMapToStringMap(ctx, me.Headers)
			meReq.Headers = &h
		}
		if !me.Labels.IsNull() && !me.Labels.IsUnknown() {
			l := tfMapToStringMap(ctx, me.Labels)
			meReq.Labels = &l
		}
		req.MetricsExport = meReq
	}

	// Links
	if !data.Links.IsNull() && !data.Links.IsUnknown() {
		var linkInputs []models.EnvironmentObjectLinkInput
		diags = data.Links.ElementsAs(ctx, &linkInputs, false)
		if diags.HasError() {
			return req, diags
		}
		createLinks := make([]platform.CreateEnvironmentObjectLinkRequest, len(linkInputs))
		for i, li := range linkInputs {
			createLinks[i] = platform.CreateEnvironmentObjectLinkRequest{
				Scope:         platform.CreateEnvironmentObjectLinkRequestScope(li.Scope.ValueString()),
				ScopeEntityId: li.ScopeEntityId.ValueString(),
			}
			overrides, d := buildCreateOverrides(ctx, li.Overrides)
			if d.HasError() {
				return req, d
			}
			if overrides != nil {
				createLinks[i].Overrides = overrides
			}
		}
		req.Links = &createLinks
	}

	// Exclude Links
	if !data.ExcludeLinks.IsNull() && !data.ExcludeLinks.IsUnknown() {
		var elInputs []models.EnvironmentObjectExcludeLinkInput
		diags = data.ExcludeLinks.ElementsAs(ctx, &elInputs, false)
		if diags.HasError() {
			return req, diags
		}
		excludeLinks := make([]platform.ExcludeLinkEnvironmentObjectRequest, len(elInputs))
		for i, el := range elInputs {
			excludeLinks[i] = platform.ExcludeLinkEnvironmentObjectRequest{
				Scope:         platform.ExcludeLinkEnvironmentObjectRequestScope(el.Scope.ValueString()),
				ScopeEntityId: el.ScopeEntityId.ValueString(),
			}
		}
		req.ExcludeLinks = &excludeLinks
	}

	return req, nil
}

func buildUpdateRequest(ctx context.Context, data *models.EnvironmentObject) (platform.UpdateEnvironmentObjectJSONRequestBody, diag.Diagnostics) {
	req := platform.UpdateEnvironmentObjectRequest{
		AutoLinkDeployments: data.AutoLinkDeployments.ValueBoolPointer(),
	}

	var diags diag.Diagnostics

	// Airflow Variable
	if !data.AirflowVariable.IsNull() && !data.AirflowVariable.IsUnknown() {
		var av models.EnvironmentObjectAirflowVariableInput
		diags = data.AirflowVariable.As(ctx, &av, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return req, diags
		}
		req.AirflowVariable = &platform.UpdateEnvironmentObjectAirflowVariableRequest{
			Value: av.Value.ValueStringPointer(),
		}
	}

	// Airflow Connection
	if !data.AirflowConnection.IsNull() && !data.AirflowConnection.IsUnknown() {
		var ci models.EnvironmentObjectAirflowConnectionInput
		diags = data.AirflowConnection.As(ctx, &ci, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return req, diags
		}
		connReq := &platform.UpdateEnvironmentObjectConnectionRequest{
			Type:       ci.Type.ValueString(),
			Host:       ci.Host.ValueStringPointer(),
			Login:      ci.Login.ValueStringPointer(),
			Password:   ci.Password.ValueStringPointer(),
			Schema:     ci.Schema.ValueStringPointer(),
			AuthTypeId: ci.AuthTypeId.ValueStringPointer(),
		}
		if !ci.Port.IsNull() && !ci.Port.IsUnknown() {
			connReq.Port = lo.ToPtr(int(ci.Port.ValueInt64()))
		}
		if !ci.Extra.IsNull() && !ci.Extra.IsUnknown() {
			var extra map[string]interface{}
			if err := json.Unmarshal([]byte(ci.Extra.ValueString()), &extra); err != nil {
				return req, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid Input", fmt.Sprintf("airflow_connection.extra must be valid JSON: %s", err))}
			}
			connReq.Extra = &extra
		}
		req.Connection = connReq
	}

	// Metrics Export
	if !data.MetricsExport.IsNull() && !data.MetricsExport.IsUnknown() {
		var me models.EnvironmentObjectMetricsExportInput
		diags = data.MetricsExport.As(ctx, &me, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return req, diags
		}
		meReq := &platform.UpdateEnvironmentObjectMetricsExportRequest{
			Endpoint:   me.Endpoint.ValueStringPointer(),
			BasicToken: me.BasicToken.ValueStringPointer(),
			Username:   me.Username.ValueStringPointer(),
			Password:   me.Password.ValueStringPointer(),
		}
		if !me.ExporterType.IsNull() && !me.ExporterType.IsUnknown() {
			meReq.ExporterType = lo.ToPtr(platform.UpdateEnvironmentObjectMetricsExportRequestExporterType(me.ExporterType.ValueString()))
		}
		if !me.AuthType.IsNull() && !me.AuthType.IsUnknown() {
			meReq.AuthType = lo.ToPtr(platform.UpdateEnvironmentObjectMetricsExportRequestAuthType(me.AuthType.ValueString()))
		}
		if !me.Headers.IsNull() && !me.Headers.IsUnknown() {
			h := tfMapToStringMap(ctx, me.Headers)
			meReq.Headers = &h
		}
		if !me.Labels.IsNull() && !me.Labels.IsUnknown() {
			l := tfMapToStringMap(ctx, me.Labels)
			meReq.Labels = &l
		}
		req.MetricsExport = meReq
	}

	// Links
	if !data.Links.IsNull() && !data.Links.IsUnknown() {
		var linkInputs []models.EnvironmentObjectLinkInput
		diags = data.Links.ElementsAs(ctx, &linkInputs, false)
		if diags.HasError() {
			return req, diags
		}
		updateLinks := make([]platform.UpdateEnvironmentObjectLinkRequest, len(linkInputs))
		for i, li := range linkInputs {
			updateLinks[i] = platform.UpdateEnvironmentObjectLinkRequest{
				Scope:         platform.UpdateEnvironmentObjectLinkRequestScope(li.Scope.ValueString()),
				ScopeEntityId: li.ScopeEntityId.ValueString(),
			}
			overrides, d := buildUpdateOverrides(ctx, li.Overrides)
			if d.HasError() {
				return req, d
			}
			if overrides != nil {
				updateLinks[i].Overrides = overrides
			}
		}
		req.Links = &updateLinks
	}

	// Exclude Links
	if !data.ExcludeLinks.IsNull() && !data.ExcludeLinks.IsUnknown() {
		var elInputs []models.EnvironmentObjectExcludeLinkInput
		diags = data.ExcludeLinks.ElementsAs(ctx, &elInputs, false)
		if diags.HasError() {
			return req, diags
		}
		excludeLinks := make([]platform.ExcludeLinkEnvironmentObjectRequest, len(elInputs))
		for i, el := range elInputs {
			excludeLinks[i] = platform.ExcludeLinkEnvironmentObjectRequest{
				Scope:         platform.ExcludeLinkEnvironmentObjectRequestScope(el.Scope.ValueString()),
				ScopeEntityId: el.ScopeEntityId.ValueString(),
			}
		}
		req.ExcludeLinks = &excludeLinks
	}

	return req, nil
}

// buildCreateOverrides unpacks the per-link `overrides` wrapper into the API's
// CreateEnvironmentObjectOverridesRequest. Returns nil when no sub-block is set.
func buildCreateOverrides(ctx context.Context, overrides types.Object) (*platform.CreateEnvironmentObjectOverridesRequest, diag.Diagnostics) {
	if overrides.IsNull() || overrides.IsUnknown() {
		return nil, nil
	}
	var ov models.EnvironmentObjectOverridesInput
	if d := overrides.As(ctx, &ov, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); d.HasError() {
		return nil, d
	}

	out := &platform.CreateEnvironmentObjectOverridesRequest{}
	hasAny := false

	if !ov.AirflowVariable.IsNull() && !ov.AirflowVariable.IsUnknown() {
		var avo models.EnvironmentObjectAirflowVariableOverridesInput
		if d := ov.AirflowVariable.As(ctx, &avo, basetypes.ObjectAsOptions{}); d.HasError() {
			return nil, d
		}
		out.AirflowVariable = &platform.CreateEnvironmentObjectAirflowVariableOverridesRequest{
			Value: avo.Value.ValueStringPointer(),
		}
		hasAny = true
	}

	if !ov.AirflowConnection.IsNull() && !ov.AirflowConnection.IsUnknown() {
		var co models.EnvironmentObjectAirflowConnectionOverridesInput
		if d := ov.AirflowConnection.As(ctx, &co, basetypes.ObjectAsOptions{}); d.HasError() {
			return nil, d
		}
		connOvr := &platform.CreateEnvironmentObjectConnectionOverridesRequest{
			Type:     co.Type.ValueStringPointer(),
			Host:     co.Host.ValueStringPointer(),
			Login:    co.Login.ValueStringPointer(),
			Password: co.Password.ValueStringPointer(),
			Schema:   co.Schema.ValueStringPointer(),
		}
		if !co.Port.IsNull() && !co.Port.IsUnknown() {
			connOvr.Port = lo.ToPtr(int(co.Port.ValueInt64()))
		}
		if !co.Extra.IsNull() && !co.Extra.IsUnknown() {
			var extra map[string]interface{}
			if err := json.Unmarshal([]byte(co.Extra.ValueString()), &extra); err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid Input", fmt.Sprintf("overrides.airflow_connection.extra must be valid JSON: %s", err))}
			}
			connOvr.Extra = &extra
		}
		out.Connection = connOvr
		hasAny = true
	}

	if !ov.MetricsExport.IsNull() && !ov.MetricsExport.IsUnknown() {
		var mo models.EnvironmentObjectMetricsExportOverridesInput
		if d := ov.MetricsExport.As(ctx, &mo, basetypes.ObjectAsOptions{}); d.HasError() {
			return nil, d
		}
		meOvr := &platform.CreateEnvironmentObjectMetricsExportOverridesRequest{
			Endpoint:   mo.Endpoint.ValueStringPointer(),
			BasicToken: mo.BasicToken.ValueStringPointer(),
			Username:   mo.Username.ValueStringPointer(),
			Password:   mo.Password.ValueStringPointer(),
		}
		if !mo.AuthType.IsNull() && !mo.AuthType.IsUnknown() {
			meOvr.AuthType = lo.ToPtr(platform.CreateEnvironmentObjectMetricsExportOverridesRequestAuthType(mo.AuthType.ValueString()))
		}
		if !mo.ExporterType.IsNull() && !mo.ExporterType.IsUnknown() {
			meOvr.ExporterType = lo.ToPtr(platform.CreateEnvironmentObjectMetricsExportOverridesRequestExporterType(mo.ExporterType.ValueString()))
		}
		if !mo.Headers.IsNull() && !mo.Headers.IsUnknown() {
			h := tfMapToStringMap(ctx, mo.Headers)
			meOvr.Headers = &h
		}
		if !mo.Labels.IsNull() && !mo.Labels.IsUnknown() {
			l := tfMapToStringMap(ctx, mo.Labels)
			meOvr.Labels = &l
		}
		out.MetricsExport = meOvr
		hasAny = true
	}

	if !hasAny {
		return nil, nil
	}
	return out, nil
}

// buildUpdateOverrides is the Update counterpart to buildCreateOverrides.
func buildUpdateOverrides(ctx context.Context, overrides types.Object) (*platform.UpdateEnvironmentObjectOverridesRequest, diag.Diagnostics) {
	if overrides.IsNull() || overrides.IsUnknown() {
		return nil, nil
	}
	var ov models.EnvironmentObjectOverridesInput
	if d := overrides.As(ctx, &ov, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); d.HasError() {
		return nil, d
	}

	out := &platform.UpdateEnvironmentObjectOverridesRequest{}
	hasAny := false

	if !ov.AirflowVariable.IsNull() && !ov.AirflowVariable.IsUnknown() {
		var avo models.EnvironmentObjectAirflowVariableOverridesInput
		if d := ov.AirflowVariable.As(ctx, &avo, basetypes.ObjectAsOptions{}); d.HasError() {
			return nil, d
		}
		out.AirflowVariable = &platform.UpdateEnvironmentObjectAirflowVariableOverridesRequest{
			Value: avo.Value.ValueStringPointer(),
		}
		hasAny = true
	}

	if !ov.AirflowConnection.IsNull() && !ov.AirflowConnection.IsUnknown() {
		var co models.EnvironmentObjectAirflowConnectionOverridesInput
		if d := ov.AirflowConnection.As(ctx, &co, basetypes.ObjectAsOptions{}); d.HasError() {
			return nil, d
		}
		connOvr := &platform.UpdateEnvironmentObjectConnectionOverridesRequest{
			Type:     co.Type.ValueStringPointer(),
			Host:     co.Host.ValueStringPointer(),
			Login:    co.Login.ValueStringPointer(),
			Password: co.Password.ValueStringPointer(),
			Schema:   co.Schema.ValueStringPointer(),
		}
		if !co.Port.IsNull() && !co.Port.IsUnknown() {
			connOvr.Port = lo.ToPtr(int(co.Port.ValueInt64()))
		}
		if !co.Extra.IsNull() && !co.Extra.IsUnknown() {
			var extra map[string]interface{}
			if err := json.Unmarshal([]byte(co.Extra.ValueString()), &extra); err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid Input", fmt.Sprintf("overrides.airflow_connection.extra must be valid JSON: %s", err))}
			}
			connOvr.Extra = &extra
		}
		out.Connection = connOvr
		hasAny = true
	}

	if !ov.MetricsExport.IsNull() && !ov.MetricsExport.IsUnknown() {
		var mo models.EnvironmentObjectMetricsExportOverridesInput
		if d := ov.MetricsExport.As(ctx, &mo, basetypes.ObjectAsOptions{}); d.HasError() {
			return nil, d
		}
		meOvr := &platform.UpdateEnvironmentObjectMetricsExportOverridesRequest{
			Endpoint:   mo.Endpoint.ValueStringPointer(),
			BasicToken: mo.BasicToken.ValueStringPointer(),
			Username:   mo.Username.ValueStringPointer(),
			Password:   mo.Password.ValueStringPointer(),
		}
		if !mo.AuthType.IsNull() && !mo.AuthType.IsUnknown() {
			meOvr.AuthType = lo.ToPtr(platform.UpdateEnvironmentObjectMetricsExportOverridesRequestAuthType(mo.AuthType.ValueString()))
		}
		if !mo.ExporterType.IsNull() && !mo.ExporterType.IsUnknown() {
			meOvr.ExporterType = lo.ToPtr(platform.UpdateEnvironmentObjectMetricsExportOverridesRequestExporterType(mo.ExporterType.ValueString()))
		}
		if !mo.Headers.IsNull() && !mo.Headers.IsUnknown() {
			h := tfMapToStringMap(ctx, mo.Headers)
			meOvr.Headers = &h
		}
		if !mo.Labels.IsNull() && !mo.Labels.IsUnknown() {
			l := tfMapToStringMap(ctx, mo.Labels)
			meOvr.Labels = &l
		}
		out.MetricsExport = meOvr
		hasAny = true
	}

	if !hasAny {
		return nil, nil
	}
	return out, nil
}

func tfMapToStringMap(ctx context.Context, m types.Map) map[string]string {
	result := make(map[string]string, len(m.Elements()))
	for k, v := range m.Elements() {
		if sv, ok := v.(types.String); ok {
			result[k] = sv.ValueString()
		}
	}
	return result
}

// buildPreserveFromModel walks the model and extracts every value the API does
// not echo back on GET — sensitive fields plus the user's exact `extra` JSON
// string. Used to repopulate state without losing user input on refresh.
// Diagnostics from .As() are surfaced; silent failure would clobber state.
func buildPreserveFromModel(ctx context.Context, data *models.EnvironmentObject) (*models.EnvironmentObjectPreserve, diag.Diagnostics) {
	preserve := &models.EnvironmentObjectPreserve{}
	var diags diag.Diagnostics

	// Airflow Connection
	if !data.AirflowConnection.IsNull() && !data.AirflowConnection.IsUnknown() {
		var ci models.EnvironmentObjectAirflowConnectionInput
		d := data.AirflowConnection.As(ctx, &ci, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		if d.HasError() {
			diags.Append(d...)
			return nil, diags
		}
		preserve.AirflowConnectionPassword = ci.Password.ValueStringPointer()
		preserve.AirflowConnectionAuthTypeId = ci.AuthTypeId.ValueStringPointer()
		preserve.AirflowConnectionExtra = ci.Extra.ValueStringPointer()
	}

	// Airflow Variable
	if !data.AirflowVariable.IsNull() && !data.AirflowVariable.IsUnknown() {
		var av models.EnvironmentObjectAirflowVariableInput
		d := data.AirflowVariable.As(ctx, &av, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		if d.HasError() {
			diags.Append(d...)
			return nil, diags
		}
		// Preserve value when secret (API returns empty for secrets) OR when caller
		// supplied a value (handles is_secret toggle edge cases).
		if av.IsSecret.ValueBool() || !av.Value.IsNull() {
			preserve.AirflowVariableValue = av.Value.ValueStringPointer()
		}
	}

	// Metrics Export
	if !data.MetricsExport.IsNull() && !data.MetricsExport.IsUnknown() {
		var me models.EnvironmentObjectMetricsExportInput
		d := data.MetricsExport.As(ctx, &me, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		if d.HasError() {
			diags.Append(d...)
			return nil, diags
		}
		preserve.MetricsExportPassword = me.Password.ValueStringPointer()
		preserve.MetricsExportBasicToken = me.BasicToken.ValueStringPointer()
	}

	// Per-link overrides
	if !data.Links.IsNull() && !data.Links.IsUnknown() {
		var linkInputs []models.EnvironmentObjectLinkInput
		d := data.Links.ElementsAs(ctx, &linkInputs, false)
		if d.HasError() {
			diags.Append(d...)
			return nil, diags
		}
		if len(linkInputs) > 0 {
			preserve.LinkOverrides = make(map[string]*models.EnvironmentObjectLinkOverridePreserve, len(linkInputs))
			for _, li := range linkInputs {
				lop, d := extractLinkOverridePreserve(ctx, li.Overrides)
				if d.HasError() {
					diags.Append(d...)
					return nil, diags
				}
				preserve.LinkOverrides[models.LinkPreserveKey(li.Scope.ValueString(), li.ScopeEntityId.ValueString())] = lop
			}
		}
	}

	return preserve, nil
}

// extractLinkOverridePreserve pulls per-link sensitive fields out of the
// collapsed `overrides` wrapper.
func extractLinkOverridePreserve(ctx context.Context, overrides types.Object) (*models.EnvironmentObjectLinkOverridePreserve, diag.Diagnostics) {
	lop := &models.EnvironmentObjectLinkOverridePreserve{}
	if overrides.IsNull() || overrides.IsUnknown() {
		return lop, nil
	}
	var ov models.EnvironmentObjectOverridesInput
	if d := overrides.As(ctx, &ov, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); d.HasError() {
		return nil, d
	}

	if !ov.AirflowVariable.IsNull() && !ov.AirflowVariable.IsUnknown() {
		var avo models.EnvironmentObjectAirflowVariableOverridesInput
		if d := ov.AirflowVariable.As(ctx, &avo, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); d.HasError() {
			return nil, d
		}
		lop.AirflowVariableValue = avo.Value.ValueStringPointer()
	}
	if !ov.AirflowConnection.IsNull() && !ov.AirflowConnection.IsUnknown() {
		var co models.EnvironmentObjectAirflowConnectionOverridesInput
		if d := ov.AirflowConnection.As(ctx, &co, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); d.HasError() {
			return nil, d
		}
		lop.AirflowConnectionPassword = co.Password.ValueStringPointer()
		lop.AirflowConnectionExtra = co.Extra.ValueStringPointer()
	}
	if !ov.MetricsExport.IsNull() && !ov.MetricsExport.IsUnknown() {
		var mo models.EnvironmentObjectMetricsExportOverridesInput
		if d := ov.MetricsExport.As(ctx, &mo, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); d.HasError() {
			return nil, d
		}
		lop.MetricsExportPassword = mo.Password.ValueStringPointer()
		lop.MetricsExportBasicToken = mo.BasicToken.ValueStringPointer()
	}
	return lop, nil
}

// ValidateConfig validates the configuration of the resource as a whole before any operations are performed.
// It enforces the documented invariants that the schema alone cannot express: the right block must be set for
// each object_type, and auto_link_deployments / links / exclude_links only apply to WORKSPACE scope.
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

	// object_type ↔ block mutual exclusivity. Skip when object_type is unknown
	// (rare — typically interpolated from another resource).
	if !data.ObjectType.IsUnknown() && !data.ObjectType.IsNull() {
		objectType := platform.CreateEnvironmentObjectRequestObjectType(data.ObjectType.ValueString())
		switch objectType {
		case platform.CreateEnvironmentObjectRequestObjectTypeAIRFLOWVARIABLE:
			if data.AirflowVariable.IsNull() {
				resp.Diagnostics.AddAttributeError(path.Root("airflow_variable"),
					"Missing required block",
					"object_type=AIRFLOW_VARIABLE requires an airflow_variable block")
			}
			if !data.AirflowConnection.IsNull() && !data.AirflowConnection.IsUnknown() {
				resp.Diagnostics.AddAttributeError(path.Root("airflow_connection"),
					"Conflicting block",
					"airflow_connection is not allowed when object_type=AIRFLOW_VARIABLE")
			}
			if !data.MetricsExport.IsNull() && !data.MetricsExport.IsUnknown() {
				resp.Diagnostics.AddAttributeError(path.Root("metrics_export"),
					"Conflicting block",
					"metrics_export is not allowed when object_type=AIRFLOW_VARIABLE")
			}
		case platform.CreateEnvironmentObjectRequestObjectTypeCONNECTION:
			if data.AirflowConnection.IsNull() {
				resp.Diagnostics.AddAttributeError(path.Root("airflow_connection"),
					"Missing required block",
					"object_type=CONNECTION requires an airflow_connection block")
			}
			if !data.AirflowVariable.IsNull() && !data.AirflowVariable.IsUnknown() {
				resp.Diagnostics.AddAttributeError(path.Root("airflow_variable"),
					"Conflicting block",
					"airflow_variable is not allowed when object_type=CONNECTION")
			}
			if !data.MetricsExport.IsNull() && !data.MetricsExport.IsUnknown() {
				resp.Diagnostics.AddAttributeError(path.Root("metrics_export"),
					"Conflicting block",
					"metrics_export is not allowed when object_type=CONNECTION")
			}
		case platform.CreateEnvironmentObjectRequestObjectTypeMETRICSEXPORT:
			if data.MetricsExport.IsNull() {
				resp.Diagnostics.AddAttributeError(path.Root("metrics_export"),
					"Missing required block",
					"object_type=METRICS_EXPORT requires a metrics_export block")
			}
			if !data.AirflowVariable.IsNull() && !data.AirflowVariable.IsUnknown() {
				resp.Diagnostics.AddAttributeError(path.Root("airflow_variable"),
					"Conflicting block",
					"airflow_variable is not allowed when object_type=METRICS_EXPORT")
			}
			if !data.AirflowConnection.IsNull() && !data.AirflowConnection.IsUnknown() {
				resp.Diagnostics.AddAttributeError(path.Root("airflow_connection"),
					"Conflicting block",
					"airflow_connection is not allowed when object_type=METRICS_EXPORT")
			}
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
