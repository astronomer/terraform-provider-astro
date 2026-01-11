package resources

import (
	"context"
	"fmt"
	"net/http"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &customRoleResource{}
var _ resource.ResourceWithImportState = &customRoleResource{}
var _ resource.ResourceWithConfigure = &customRoleResource{}

func NewCustomRoleResource() resource.Resource {
	return &customRoleResource{}
}

// customRoleResource defines the resource implementation.
type customRoleResource struct {
	IamClient      *iam.ClientWithResponses
	OrganizationId string
}

func (r *customRoleResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_custom_role"
}

func (r *customRoleResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Custom role resource",
		Attributes:          schemas.CustomRoleResourceSchemaAttributes(),
	}
}

func (r *customRoleResource) Configure(
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

	r.IamClient = apiClients.IamClient
	r.OrganizationId = apiClients.OrganizationId
}

func (r *customRoleResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.CustomRole

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert permissions Set to slice
	var permissions []string
	resp.Diagnostics.Append(data.Permissions.ElementsAs(ctx, &permissions, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert restricted workspace IDs Set to slice (optional)
	var restrictedWorkspaceIds []string
	if !data.RestrictedWorkspaceIds.IsNull() && !data.RestrictedWorkspaceIds.IsUnknown() {
		resp.Diagnostics.Append(data.RestrictedWorkspaceIds.ElementsAs(ctx, &restrictedWorkspaceIds, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Create request
	createCustomRoleRequest := iam.CreateCustomRoleRequest{
		Name:        data.Name.ValueString(),
		Permissions: permissions,
		ScopeType:   iam.CreateCustomRoleRequestScopeType(data.ScopeType.ValueString()),
	}

	// Set optional fields
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		createCustomRoleRequest.Description = lo.ToPtr(data.Description.ValueString())
	}

	if len(restrictedWorkspaceIds) > 0 {
		createCustomRoleRequest.RestrictedWorkspaceIds = &restrictedWorkspaceIds
	}

	customRole, err := r.IamClient.CreateCustomRoleWithResponse(
		ctx,
		r.OrganizationId,
		createCustomRoleRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to create custom role", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create custom role, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, customRole.HTTPResponse, customRole.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}
	if customRole.JSON200 == nil {
		tflog.Error(ctx, "failed to create custom role", map[string]interface{}{"error": "nil response"})
		resp.Diagnostics.AddError(
			"Client Error",
			"Unable to create custom role, got nil response",
		)
		return
	}

	diags := data.ReadFromResponse(ctx, customRole.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created a custom role resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *customRoleResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.CustomRole

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get request
	customRole, err := r.IamClient.GetCustomRoleWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to get custom role", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get custom role, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, customRole.HTTPResponse, customRole.Body)
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
	if customRole.JSON200 == nil {
		tflog.Error(ctx, "failed to get custom role", map[string]interface{}{"error": "nil response"})
		resp.Diagnostics.AddError(
			"Client Error",
			"Unable to get custom role, got nil response",
		)
		return
	}

	diags := data.ReadFromResponse(ctx, customRole.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read a custom role resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *customRoleResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data models.CustomRole

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert permissions Set to slice
	var permissions []string
	resp.Diagnostics.Append(data.Permissions.ElementsAs(ctx, &permissions, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert restricted workspace IDs Set to slice (optional)
	var restrictedWorkspaceIds []string
	if !data.RestrictedWorkspaceIds.IsNull() && !data.RestrictedWorkspaceIds.IsUnknown() {
		resp.Diagnostics.Append(data.RestrictedWorkspaceIds.ElementsAs(ctx, &restrictedWorkspaceIds, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Update request
	updateCustomRoleRequest := iam.UpdateCustomRoleRequest{
		Name:        data.Name.ValueString(),
		Permissions: permissions,
	}

	// Set optional fields
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		updateCustomRoleRequest.Description = lo.ToPtr(data.Description.ValueString())
	}

	if len(restrictedWorkspaceIds) > 0 {
		updateCustomRoleRequest.RestrictedWorkspaceIds = &restrictedWorkspaceIds
	}

	customRole, err := r.IamClient.UpdateCustomRoleWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
		updateCustomRoleRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to update custom role", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update custom role, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, customRole.HTTPResponse, customRole.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}
	if customRole.JSON200 == nil {
		tflog.Error(ctx, "failed to update custom role", map[string]interface{}{"error": "nil response"})
		resp.Diagnostics.AddError(
			"Client Error",
			"Unable to update custom role, got nil response",
		)
		return
	}

	diags := data.ReadFromResponse(ctx, customRole.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated a custom role resource: %v", data.Id.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *customRoleResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.CustomRole

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete request
	customRole, err := r.IamClient.DeleteCustomRoleWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to delete custom role", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete custom role, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, customRole.HTTPResponse, customRole.Body)
	// It is recommended to ignore 404 Resource Not Found errors when deleting a resource
	if statusCode != http.StatusNotFound && diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted a custom role resource: %v", data.Id.ValueString()))
}

func (r *customRoleResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
