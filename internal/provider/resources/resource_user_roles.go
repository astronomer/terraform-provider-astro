package resources

import (
	"context"
	"fmt"
	"net/http"

	"github.com/astronomer/terraform-provider-astro/internal/provider/common"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &userRolesResource{}
var _ resource.ResourceWithImportState = &userRolesResource{}
var _ resource.ResourceWithConfigure = &userRolesResource{}

func NewUserRolesResource() resource.Resource {
	return &userRolesResource{}
}

// userRolesResource defines the resource implementation.
type userRolesResource struct {
	iamClient      *iam.ClientWithResponses
	organizationId string
}

func (r *userRolesResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_user_roles"
}

func (r *userRolesResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "User Roles resource",
		Attributes:          schemas.ResourceUserRolesSchemaAttributes(),
	}
}

func (r *userRolesResource) Configure(
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

	r.iamClient = apiClients.IamClient
	r.organizationId = apiClients.OrganizationId
}

func (r *userRolesResource) MutateRoles(
	ctx context.Context,
	data *models.UserRoles,
) diag.Diagnostics {
	userId := data.UserId.ValueString()

	// Then convert the models to the request types for the API
	workspaceRoles, diags := common.RequestWorkspaceRoles(ctx, data.WorkspaceRoles)
	if diags.HasError() {
		return diags
	}
	deploymentRoles, diags := common.RequestDeploymentRoles(ctx, data.DeploymentRoles)
	if diags.HasError() {
		return diags
	}

	// create request
	updateUserRolesRequest := iam.UpdateUserRolesJSONRequestBody{
		DeploymentRoles:  &deploymentRoles,
		OrganizationRole: lo.ToPtr(iam.UpdateUserRolesRequestOrganizationRole(data.OrganizationRole.ValueString())),
		WorkspaceRoles:   &workspaceRoles,
	}
	userRoles, err := r.iamClient.UpdateUserRolesWithResponse(
		ctx,
		r.organizationId,
		userId,
		updateUserRolesRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to mutate user_roles", map[string]interface{}{"error": err})
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to mutate user_roles, got error: %s", err),
		)
		return diags
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, userRoles.HTTPResponse, userRoles.Body)
	if diagnostic != nil {
		diags.Append(diagnostic)
		return diags
	}

	diags = data.ReadFromResponse(ctx, userId, userRoles.JSON200)
	if diags.HasError() {
		return diags
	}

	return nil
}

func (r *userRolesResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.UserRoles

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := r.MutateRoles(ctx, &data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("created a user_roles resource for user '%v'", data.UserId.ValueString()))
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userRolesResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.UserRoles

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userId := data.UserId.ValueString()

	// get request
	userRoles, err := r.iamClient.GetUserWithResponse(
		ctx,
		r.organizationId,
		userId,
	)
	if err != nil {
		tflog.Error(ctx, "failed to get user_roles", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get user_roles, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, userRoles.HTTPResponse, userRoles.Body)
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

	// Generate subjectRoles from the get user API response
	subjectRoles := iam.SubjectRoles{
		OrganizationRole: lo.ToPtr(iam.SubjectRolesOrganizationRole(*(*string)(userRoles.JSON200.OrganizationRole))),
		WorkspaceRoles:   userRoles.JSON200.WorkspaceRoles,
		DeploymentRoles:  userRoles.JSON200.DeploymentRoles,
	}
	diags := data.ReadFromResponse(ctx, userId, &subjectRoles)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read a user_roles resource: %v", userId))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userRolesResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data models.UserRoles

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := r.MutateRoles(ctx, &data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("updated a user_roles resource for user '%v'", data.UserId.ValueString()))
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userRolesResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.UserRoles

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// delete request
	userId := data.UserId.ValueString()

	// update request with no workspace roles, no deployment roles and lowest organization role
	updateUserRolesRequest := iam.UpdateUserRolesJSONRequestBody{
		DeploymentRoles:  nil,
		OrganizationRole: lo.ToPtr(iam.UpdateUserRolesRequestOrganizationRole(iam.ORGANIZATIONMEMBER)),
		WorkspaceRoles:   nil,
	}
	userRoles, err := r.iamClient.UpdateUserRolesWithResponse(
		ctx,
		r.organizationId,
		userId,
		updateUserRolesRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to delete user_roles", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete user_roles, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, userRoles.HTTPResponse, userRoles.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	diags := data.ReadFromResponse(ctx, userId, userRoles.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted a user_roles resource for user '%v'", userId))
}

func (r *userRolesResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("user_id"), req, resp)
}
