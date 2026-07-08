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
var _ resource.Resource = &alertResource{}
var _ resource.ResourceWithImportState = &alertResource{}
var _ resource.ResourceWithConfigure = &alertResource{}

func NewAlertResource() resource.Resource {
	return &alertResource{}
}

// alertResource defines the resource implementation.
type alertResource struct {
	platformClient *platform.ClientWithResponses
	organizationId string
}

func (r *alertResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_alert"
}

func (r *alertResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Manage a single alert.\n\n" +
			"~> **Deprecated** Use the `astro_alerts` resource instead, which manages one or many alerts in a " +
			"single resource. Do not manage the same alert with both `astro_alert` and `astro_alerts` — they " +
			"conflict on every apply.",
		DeprecationMessage: "Use the astro_alerts resource instead, which manages one or many alerts in a " +
			"single resource. Do not manage the same alert with both astro_alert and astro_alerts — they will " +
			"conflict on every apply. astro_alert may be removed in a future major version.",
		Attributes: schemas.AlertResourceSchemaAttributes(),
	}
}

func (r *alertResource) Configure(
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

func (r *alertResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.AlertResource

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createAlertRequest, diags := BuildCreateAlertRequest(ctx, data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Call platform to create
	alertResp, err := r.platformClient.CreateAlertWithResponse(ctx, r.organizationId, createAlertRequest)
	if err != nil {
		tflog.Error(ctx, "failed to create alert", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create alert: %s", err))
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, alertResp.HTTPResponse, alertResp.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	// Map response into state
	diags = data.ReadFromResponse(ctx, alertResp.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created alert resource %s", data.Id.ValueString()))

	// Save to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *alertResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.AlertResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// get request
	alert, err := r.platformClient.GetAlertWithResponse(
		ctx,
		r.organizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to get alert", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get alert, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, alert.HTTPResponse, alert.Body)
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

	diags := data.ReadFromResponse(ctx, alert.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read a alert resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *alertResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data models.AlertResource

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateBody, diags := BuildUpdateAlertRequest(ctx, data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Call platform update
	alertResp, err := r.platformClient.UpdateAlertWithResponse(ctx, r.organizationId, data.Id.ValueString(), updateBody)
	if err != nil {
		tflog.Error(ctx, "failed to update alert", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update alert: %s", err))
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, alertResp.HTTPResponse, alertResp.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	// Map updated response
	diags = data.ReadFromResponse(ctx, alertResp.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated alert resource %s", data.Id.ValueString()))

	// Save to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *alertResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.AlertResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// delete request
	alert, err := r.platformClient.DeleteAlertWithResponse(
		ctx,
		r.organizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to delete alert", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete alert, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, alert.HTTPResponse, alert.Body)
	// It is recommended to ignore 404 Resource Not Found errors when deleting a resource
	if statusCode != http.StatusNotFound && diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted a alert resource: %v", data.Id.ValueString()))
}

func (r *alertResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
