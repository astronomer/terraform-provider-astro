package resources

import (
	"context"
	"fmt"
	"net/http"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &environmnetObjectResource{}
var _ resource.ResourceWithImportState = &environmnetObjectResource{}
var _ resource.ResourceWithConfigure = &environmnetObjectResource{}

func NewEnvironmentObjectResource() resource.Resource {
	return &environmnetObjectResource{}
}

// environmnetObjectResource defines the resource implementation.
type environmnetObjectResource struct {
	platformClient *platform.ClientWithResponses
	organizationId string
}

func (r *environmnetObjectResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_environmnetObject"
}

func (r *environmnetObjectResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "EnvironmentObject resource",
		Attributes:          schemas.EnvironmentObjectResourceSchemaAttributes(),
	}
}

func (r *environmnetObjectResource) Configure(
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

func (r *environmnetObjectResource) Create(
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

	//type CreateEnvironmentObjectRequest struct {
	//	AirflowVariable *CreateEnvironmentObjectAirflowVariableRequest `json:"airflowVariable,omitempty"`
	//
	//	// AutoLinkDeployments Whether or not to automatically link Deployments to the environment object. Only applicable for WORKSPACE scope
	//	AutoLinkDeployments *bool                                     `json:"autoLinkDeployments,omitempty"`
	//	Connection          *CreateEnvironmentObjectConnectionRequest `json:"connection,omitempty"`
	//
	//	// ExcludeLinks The links to exclude from the environment object. Only applicable for WORKSPACE scope
	//	ExcludeLinks *[]ExcludeLinkEnvironmentObjectRequest `json:"excludeLinks,omitempty"`
	//
	//	// Links The Deployments that Astro links to the environment object. Only applicable for WORKSPACE scope
	//	Links         *[]CreateEnvironmentObjectLinkRequest        `json:"links,omitempty"`
	//	MetricsExport *CreateEnvironmentObjectMetricsExportRequest `json:"metricsExport,omitempty"`
	//
	//	// ObjectKey The key for the environment object
	//	ObjectKey string `json:"objectKey"`
	//
	//	// ObjectType The type of environment object
	//	ObjectType CreateEnvironmentObjectRequestObjectType `json:"objectType"`
	//
	//	// Scope The scope of the environment object
	//	Scope CreateEnvironmentObjectRequestScope `json:"scope"`
	//
	//	// ScopeEntityId The ID of the scope entity where the environment object is created
	//	ScopeEntityId string `json:"scopeEntityId"`
	//}

	// create request
	createEnvironmentObjectRequest := platform.CreateEnvironmentObjectJSONRequestBody{
		CicdEnforcedDefault: data.CicdEnforcedDefault.ValueBoolPointer(),
		Description:         data.Description.ValueStringPointer(),
		Name:                data.Name.ValueString(),
	}
	environmnetObject, err := r.platformClient.CreateEnvironmentObjectWithResponse(
		ctx,
		r.organizationId,
		createEnvironmentObjectRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to create environmnetObject", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create environmnetObject, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, environmnetObject.HTTPResponse, environmnetObject.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	diags := data.ReadFromResponse(ctx, environmnetObject.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created a environmnetObject resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *environmnetObjectResource) Read(
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
	environmnetObject, err := r.platformClient.GetEnvironmentObjectWithResponse(
		ctx,
		r.organizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to get environmnetObject", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get environmnetObject, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, environmnetObject.HTTPResponse, environmnetObject.Body)
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

	diags := data.ReadFromResponse(ctx, environmnetObject.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read a environmnetObject resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *environmnetObjectResource) Update(
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
	updateEnvironmentObjectRequest := platform.UpdateEnvironmentObjectJSONRequestBody{
		CicdEnforcedDefault: data.CicdEnforcedDefault.ValueBool(),
		Description:         data.Description.ValueString(),
		Name:                data.Name.ValueString(),
	}
	environmnetObject, err := r.platformClient.UpdateEnvironmentObjectWithResponse(
		ctx,
		r.organizationId,
		data.Id.ValueString(),
		updateEnvironmentObjectRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to update environmnetObject", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update environmnetObject, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, environmnetObject.HTTPResponse, environmnetObject.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	diags := data.ReadFromResponse(ctx, environmnetObject.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated a environmnetObject resource: %v", data.Id.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *environmnetObjectResource) Delete(
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
	environmnetObject, err := r.platformClient.DeleteEnvironmentObjectWithResponse(
		ctx,
		r.organizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to delete environmnetObject", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete environmnetObject, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, environmnetObject.HTTPResponse, environmnetObject.Body)
	// It is recommended to ignore 404 Resource Not Found errors when deleting a resource
	if statusCode != http.StatusNotFound && diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted a environmnetObject resource: %v", data.Id.ValueString()))
}

func (r *environmnetObjectResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
