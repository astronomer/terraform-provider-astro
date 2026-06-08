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

	// create request
	createReq, diags := buildCreateRequest(ctx, &data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	createResp, err := r.platformClient.CreateEnvironmentObjectWithResponse(ctx, r.organizationId, createReq)
	if err != nil {
		tflog.Error(ctx, "failed to create environment object", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create environment object: %s", err))
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, createResp.HTTPResponse, createResp.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	// Create only returns the ID, do a follow-up GET to populate full state
	getResp, err := r.platformClient.GetEnvironmentObjectWithResponse(ctx, r.organizationId, createResp.JSON200.Id)
	if err != nil {
		tflog.Error(ctx, "failed to get environment object after create", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read environment object after create: %s", err))
		return
	}
	_, diagnostic = clients.NormalizeAPIError(ctx, getResp.HTTPResponse, getResp.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	diags = data.ReadFromResponse(ctx, getResp.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created a environment object resource: %v", data.Id.ValueString()))

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

	// get request
	envObj, err := r.platformClient.GetEnvironmentObjectWithResponse(ctx, r.organizationId, data.Id.ValueString())
	if err != nil {
		tflog.Error(ctx, "failed to get environment object", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get environment object: %s", err))
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

	diags := data.ReadFromResponse(ctx, envObj.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read a environment object resource: %v", data.Id.ValueString()))

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

	// update request
	updateReq, diags := buildUpdateRequest(ctx, &data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	envObj, err := r.platformClient.UpdateEnvironmentObjectWithResponse(ctx, r.organizationId, data.Id.ValueString(), updateReq)
	if err != nil {
		tflog.Error(ctx, "failed to update environment object", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update environment object: %s", err))
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, envObj.HTTPResponse, envObj.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	diags = data.ReadFromResponse(ctx, envObj.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated a environment object resource: %v", data.Id.ValueString()))

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
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete environment object: %s", err))
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, envObj.HTTPResponse, envObj.Body)
	// It is recommended to ignore 404 Resource Not Found errors when deleting a resource
	if statusCode != http.StatusNotFound && diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted a environment object resource: %v", data.Id.ValueString()))
}

func (r *environmentObjectResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type connectionInput struct {
	AuthTypeId types.String `tfsdk:"auth_type_id"`
	Type       types.String `tfsdk:"type"`
	Host       types.String `tfsdk:"host"`
	Port       types.Int64  `tfsdk:"port"`
	Schema     types.String `tfsdk:"schema"`
	Login      types.String `tfsdk:"login"`
	Password   types.String `tfsdk:"password"`
	Extra      types.String `tfsdk:"extra"`
}

type airflowVariableInput struct {
	Value    types.String `tfsdk:"value"`
	IsSecret types.Bool   `tfsdk:"is_secret"`
}

type metricsExportInput struct {
	AuthType     types.String `tfsdk:"auth_type"`
	Endpoint     types.String `tfsdk:"endpoint"`
	BasicToken   types.String `tfsdk:"basic_token"`
	ExporterType types.String `tfsdk:"exporter_type"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
	Headers      types.Map    `tfsdk:"headers"`
	Labels       types.Map    `tfsdk:"labels"`
}

type excludeLinkInput struct {
	Scope         types.String `tfsdk:"scope"`
	ScopeEntityId types.String `tfsdk:"scope_entity_id"`
}

type linkInput struct {
	Scope                    types.String `tfsdk:"scope"`
	ScopeEntityId            types.String `tfsdk:"scope_entity_id"`
	AirflowVariableOverrides types.Object `tfsdk:"airflow_variable_overrides"`
	ConnectionOverrides      types.Object `tfsdk:"connection_overrides"`
	MetricsExportOverrides   types.Object `tfsdk:"metrics_export_overrides"`
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
		var av airflowVariableInput
		diags = data.AirflowVariable.As(ctx, &av, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return req, diags
		}
		req.AirflowVariable = &platform.CreateEnvironmentObjectAirflowVariableRequest{
			Value:    av.Value.ValueStringPointer(),
			IsSecret: av.IsSecret.ValueBoolPointer(),
		}
	}

	// Connection
	if !data.ConnectionConfig.IsNull() && !data.ConnectionConfig.IsUnknown() {
		var ci connectionInput
		diags = data.ConnectionConfig.As(ctx, &ci, basetypes.ObjectAsOptions{})
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
				return req, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid Input", fmt.Sprintf("connection.extra must be valid JSON: %s", err))}
			}
			connReq.Extra = &extra
		}
		req.Connection = connReq
	}

	// Metrics Export
	if !data.MetricsExport.IsNull() && !data.MetricsExport.IsUnknown() {
		var me metricsExportInput
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
		var linkInputs []linkInput
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
			overrides, d := buildCreateOverrides(ctx, &li)
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
		var elInputs []excludeLinkInput
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
		var av airflowVariableInput
		diags = data.AirflowVariable.As(ctx, &av, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return req, diags
		}
		req.AirflowVariable = &platform.UpdateEnvironmentObjectAirflowVariableRequest{
			Value: av.Value.ValueStringPointer(),
		}
	}

	// Connection
	if !data.ConnectionConfig.IsNull() && !data.ConnectionConfig.IsUnknown() {
		var ci connectionInput
		diags = data.ConnectionConfig.As(ctx, &ci, basetypes.ObjectAsOptions{})
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
				return req, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid Input", fmt.Sprintf("connection.extra must be valid JSON: %s", err))}
			}
			connReq.Extra = &extra
		}
		req.Connection = connReq
	}

	// Metrics Export
	if !data.MetricsExport.IsNull() && !data.MetricsExport.IsUnknown() {
		var me metricsExportInput
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
		var linkInputs []linkInput
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
			overrides, d := buildUpdateOverrides(ctx, &li)
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
		var elInputs []excludeLinkInput
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

type connectionOverridesInput struct {
	Type     types.String `tfsdk:"type"`
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
	Schema   types.String `tfsdk:"schema"`
	Login    types.String `tfsdk:"login"`
	Password types.String `tfsdk:"password"`
	Extra    types.String `tfsdk:"extra"`
}

type metricsExportOverridesInput struct {
	AuthType     types.String `tfsdk:"auth_type"`
	Endpoint     types.String `tfsdk:"endpoint"`
	BasicToken   types.String `tfsdk:"basic_token"`
	ExporterType types.String `tfsdk:"exporter_type"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
	Headers      types.Map    `tfsdk:"headers"`
	Labels       types.Map    `tfsdk:"labels"`
}

type airflowVariableOverridesInput struct {
	Value types.String `tfsdk:"value"`
}

func buildCreateOverrides(ctx context.Context, li *linkInput) (*platform.CreateEnvironmentObjectOverridesRequest, diag.Diagnostics) {
	overrides := &platform.CreateEnvironmentObjectOverridesRequest{}
	hasOverrides := false

	if !li.AirflowVariableOverrides.IsNull() && !li.AirflowVariableOverrides.IsUnknown() {
		var avo airflowVariableOverridesInput
		diags := li.AirflowVariableOverrides.As(ctx, &avo, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
		}
		overrides.AirflowVariable = &platform.CreateEnvironmentObjectAirflowVariableOverridesRequest{
			Value: avo.Value.ValueStringPointer(),
		}
		hasOverrides = true
	}

	if !li.ConnectionOverrides.IsNull() && !li.ConnectionOverrides.IsUnknown() {
		var co connectionOverridesInput
		diags := li.ConnectionOverrides.As(ctx, &co, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
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
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid Input", "connection overrides extra must be valid JSON")}
			}
			connOvr.Extra = &extra
		}
		overrides.Connection = connOvr
		hasOverrides = true
	}

	if !li.MetricsExportOverrides.IsNull() && !li.MetricsExportOverrides.IsUnknown() {
		var mo metricsExportOverridesInput
		diags := li.MetricsExportOverrides.As(ctx, &mo, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
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
		overrides.MetricsExport = meOvr
		hasOverrides = true
	}

	if !hasOverrides {
		return nil, nil
	}
	return overrides, nil
}

func buildUpdateOverrides(ctx context.Context, li *linkInput) (*platform.UpdateEnvironmentObjectOverridesRequest, diag.Diagnostics) {
	overrides := &platform.UpdateEnvironmentObjectOverridesRequest{}
	hasOverrides := false

	if !li.AirflowVariableOverrides.IsNull() && !li.AirflowVariableOverrides.IsUnknown() {
		var avo airflowVariableOverridesInput
		diags := li.AirflowVariableOverrides.As(ctx, &avo, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
		}
		overrides.AirflowVariable = &platform.UpdateEnvironmentObjectAirflowVariableOverridesRequest{
			Value: avo.Value.ValueStringPointer(),
		}
		hasOverrides = true
	}

	if !li.ConnectionOverrides.IsNull() && !li.ConnectionOverrides.IsUnknown() {
		var co connectionOverridesInput
		diags := li.ConnectionOverrides.As(ctx, &co, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
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
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid Input", "connection overrides extra must be valid JSON")}
			}
			connOvr.Extra = &extra
		}
		overrides.Connection = connOvr
		hasOverrides = true
	}

	if !li.MetricsExportOverrides.IsNull() && !li.MetricsExportOverrides.IsUnknown() {
		var mo metricsExportOverridesInput
		diags := li.MetricsExportOverrides.As(ctx, &mo, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
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
		overrides.MetricsExport = meOvr
		hasOverrides = true
	}

	if !hasOverrides {
		return nil, nil
	}
	return overrides, nil
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
