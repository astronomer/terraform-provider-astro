package resources

import (
	"context"
	"fmt"
	"net/http"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"

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

var _ resource.Resource = &CustomRoleResource{}
var _ resource.ResourceWithImportState = &CustomRoleResource{}
var _ resource.ResourceWithConfigure = &CustomRoleResource{}

func NewCustomRoleResource() resource.Resource {
	return &CustomRoleResource{}
}

// CustomRoleResource defines the resource implementation.
type CustomRoleResource struct {
	IamClient      *iam.ClientWithResponses
	PlatformClient *platform.ClientWithResponses
	OrganizationId string
}

func (r *CustomRoleResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *CustomRoleResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "CustomRole resource",
		Attributes:          schemas.CustomRoleResourceSchemaAttributes(),
	}
}

func (r *CustomRoleResource) Configure(
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
	r.PlatformClient = apiClients.PlatformClient
	r.OrganizationId = apiClients.OrganizationId
}

func (r *CustomRoleResource) MutateRoles(
	ctx context.Context,
	data *models.CustomRoleResource,
	teamId string,
) diag.Diagnostics {
	// Convert the models to the request types for the API
	workspaceRoles, diags := common.RequestWorkspaceRoles(ctx, data.WorkspaceRoles)
	if diags.HasError() {
		return diags
	}
	deploymentRoles, diags := common.RequestDeploymentRoles(ctx, data.DeploymentRoles)
	if diags.HasError() {
		return diags
	}

	// Validate the roles
	diags = common.ValidateRoles(workspaceRoles, deploymentRoles)
	if diags.HasError() {
		return diags
	}

	diags = common.ValidateWorkspaceDeploymentRoles(ctx, common.ValidateWorkspaceDeploymentRolesInput{
		PlatformClient:  r.PlatformClient,
		OrganizationId:  r.OrganizationId,
		WorkspaceRoles:  workspaceRoles,
		DeploymentRoles: deploymentRoles,
	})
	if diags.HasError() {
		return diags
	}

	// Update team roles
	updateCustomRoleRolesRequest := iam.UpdateCustomRoleRolesJSONRequestBody{
		DeploymentRoles:  &deploymentRoles,
		OrganizationRole: iam.UpdateCustomRoleRolesRequestOrganizationRole(data.OrganizationRole.ValueString()),
		WorkspaceRoles:   &workspaceRoles,
	}
	teamRoles, err := r.IamClient.UpdateCustomRoleRolesWithResponse(
		ctx,
		r.OrganizationId,
		teamId,
		updateCustomRoleRolesRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to mutate CustomRole roles", map[string]interface{}{"error": err})
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to mutate CustomRole roles, got error: %s", err),
		)
		return diags
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, teamRoles.HTTPResponse, teamRoles.Body)
	if diagnostic != nil {
		diags.Append(diagnostic)
		return diags
	}

	return nil
}

func (r *CustomRoleResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.CustomRoleResource

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var diags diag.Diagnostics

	// Check if the organization is SCIM enabled, if it is return an error
	diags = r.CheckOrganizationIsScim(ctx)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	memberIds, diags := utils.TypesSetToStringSlice(ctx, data.MemberIds)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Create the team request
	createCustomRoleRequest := iam.CreateCustomRoleRequest{
		Name:             data.Name.ValueString(),
		Description:      data.Description.ValueStringPointer(),
		MemberIds:        &memberIds,
		OrganizationRole: lo.ToPtr(iam.CreateCustomRoleRequestOrganizationRole(data.OrganizationRole.ValueString())),
	}

	// Create the team
	team, err := r.IamClient.CreateCustomRoleWithResponse(
		ctx,
		r.OrganizationId,
		createCustomRoleRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to create CustomRole", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create CustomRole, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, team.HTTPResponse, team.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	teamId := team.JSON200.Id

	// Update team roles
	if !data.WorkspaceRoles.IsNull() || !data.DeploymentRoles.IsNull() {
		diags = r.MutateRoles(ctx, &data, teamId)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)

			// if there is an error in creating team with workspace or deployment roles, delete the team
			team, err := r.IamClient.DeleteCustomRoleWithResponse(
				ctx,
				r.OrganizationId,
				teamId,
			)
			if err != nil {
				tflog.Error(ctx, "failed to delete CustomRole", map[string]interface{}{"error": err})
				resp.Diagnostics.AddError(
					"Client Error",
					fmt.Sprintf("Unable to delete CustomRole, got error: %s", err),
				)
				return
			}
			statusCode, diagnostic := clients.NormalizeAPIError(ctx, team.HTTPResponse, team.Body)
			// It is recommended to ignore 404 Resource Not Found errors when deleting a resource
			if statusCode != http.StatusNotFound && diagnostic != nil {
				resp.Diagnostics.Append(diagnostic)
				return
			}

			return
		}
	}

	// Get CustomRole and use this as data since it will have the correct roles
	teamResp, err := r.IamClient.GetCustomRoleWithResponse(
		ctx,
		r.OrganizationId,
		teamId,
	)
	if err != nil {
		tflog.Error(ctx, "failed to create CustomRole", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create and get CustomRole, got error: %s", err),
		)
		return
	}

	diags = data.ReadFromResponse(ctx, teamResp.JSON200, &memberIds)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created a CustomRole resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomRoleResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.CustomRoleResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// get request
	team, err := r.IamClient.GetCustomRoleWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
	)

	if err != nil {
		tflog.Error(ctx, "failed to get CustomRole", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get CustomRole, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, team.HTTPResponse, team.Body)
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

	memberIds, diags := utils.TypesSetToStringSlice(ctx, data.MemberIds)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = data.ReadFromResponse(ctx, team.JSON200, &memberIds)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read a CustomRole resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomRoleResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data models.CustomRoleResource

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var diags diag.Diagnostics

	// Check if the organization is SCIM enabled, if it is return an error
	diags = r.CheckOrganizationIsScim(ctx)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Update team members
	newMemberIds, diags := r.UpdateCustomRoleMembers(ctx, data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Update team
	updateCustomRoleRequest := iam.UpdateCustomRoleRequest{
		Name: data.Name.ValueString(),
	}

	if !data.Description.IsNull() {
		updateCustomRoleRequest.Description = data.Description.ValueStringPointer()
	} else {
		updateCustomRoleRequest.Description = lo.ToPtr("")
	}

	team, err := r.IamClient.UpdateCustomRoleWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
		updateCustomRoleRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to update CustomRole", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update CustomRole, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, team.HTTPResponse, team.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	// Update team roles
	if !data.WorkspaceRoles.IsNull() || !data.DeploymentRoles.IsNull() {
		diags = r.MutateRoles(ctx, &data, data.Id.ValueString())
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

	}

	// Get CustomRole and use this as data since it will have the correct roles
	teamResp, err := r.IamClient.GetCustomRoleWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to update CustomRole", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update and get CustomRole, got error: %s", err),
		)
		return
	}

	diags = data.ReadFromResponse(ctx, teamResp.JSON200, &newMemberIds)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated a CustomRole resource: %v", data.Id.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomRoleResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.CustomRoleResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// delete request
	team, err := r.IamClient.DeleteCustomRoleWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to delete CustomRole", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete CustomRole, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, team.HTTPResponse, team.Body)
	// It is recommended to ignore 404 Resource Not Found errors when deleting a resource
	if statusCode != http.StatusNotFound && diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted a CustomRole resource: %v", data.Id.ValueString()))
}

func (r *CustomRoleResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *CustomRoleResource) CheckOrganizationIsScim(ctx context.Context) diag.Diagnostics {
	// Validate if org isScimEnabled and return error if it is
	org, err := r.PlatformClient.GetOrganizationWithResponse(ctx, r.OrganizationId, nil)
	if err != nil {
		tflog.Error(ctx, "failed to validate CustomRole", map[string]interface{}{"error": err})
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Client Error",
				fmt.Sprintf("Unable to validate CustomRole, got error: %s", err),
			),
		}
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, org.HTTPResponse, org.Body)
	if diagnostic != nil {
		return diag.Diagnostics{diagnostic}
	}
	if org.JSON200 == nil {
		tflog.Error(ctx, "failed to get organization", map[string]interface{}{"error": "nil response"})
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Client Error",
				fmt.Sprintf("Unable to read organization %v, got nil response", r.OrganizationId)),
		}
	}
	if org.JSON200.IsScimEnabled {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Invalid Configuration: Cannot create, update or delete a CustomRole resource when SCIM is enabled",
				"Please disable SCIM in the organization settings to manage CustomRole resources",
			),
		}
	}
	return nil
}

func (r *CustomRoleResource) UpdateCustomRoleMembers(ctx context.Context, data models.CustomRoleResource) ([]string, diag.Diagnostics) {
	// get existing team members
	teamMembersResp, err := r.IamClient.ListCustomRoleMembersWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
		nil,
	)
	if err != nil {
		tflog.Error(ctx, "failed to update CustomRole", map[string]interface{}{"error": err})
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Client Error",
				fmt.Sprintf("Unable to list existing CustomRole members, got error: %s", err),
			),
		}
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, teamMembersResp.HTTPResponse, teamMembersResp.Body)
	if diagnostic != nil {
		return nil, diag.Diagnostics{diagnostic}
	}

	teamMembers := teamMembersResp.JSON200.CustomRoleMembers
	memberIds := lo.Map(teamMembers, func(tm iam.CustomRoleMember, _ int) string {
		return tm.UserId
	})

	// get list of new member ids
	newMemberIds, diags := utils.TypesSetToStringSlice(ctx, data.MemberIds)
	if diags.HasError() {
		return nil, diags
	}

	// find the difference between the two lists and update the team members
	deleteIds, addIds := lo.Difference(memberIds, newMemberIds)

	// delete the members that are not in the new list
	if len(deleteIds) > 0 {
		for _, id := range deleteIds {
			removeCustomRoleMemberResp, err := r.IamClient.RemoveCustomRoleMemberWithResponse(
				ctx,
				r.OrganizationId,
				data.Id.ValueString(),
				id,
			)
			if err != nil {
				tflog.Error(ctx, "failed to update CustomRole", map[string]interface{}{"error": err})
				return nil, diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Client Error",
						fmt.Sprintf("Unable to remove CustomRole member, got error: %s", err),
					),
				}
			}
			_, diagnostic = clients.NormalizeAPIError(ctx, removeCustomRoleMemberResp.HTTPResponse, removeCustomRoleMemberResp.Body)
			if diagnostic != nil {
				return nil, diag.Diagnostics{diagnostic}
			}
		}
	}

	// add the members that are in the new list
	if len(addIds) > 0 {
		addCustomRoleMembersRequest := iam.AddCustomRoleMembersRequest{
			MemberIds: addIds,
		}
		addCustomRoleMembersResp, err := r.IamClient.AddCustomRoleMembersWithResponse(
			ctx,
			r.OrganizationId,
			data.Id.ValueString(),
			addCustomRoleMembersRequest,
		)
		if err != nil {
			tflog.Error(ctx, "failed to update CustomRole", map[string]interface{}{"error": err})
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Client Error",
					fmt.Sprintf("Unable to add CustomRole members, got error: %s", err),
				),
			}
		}
		_, diagnostic = clients.NormalizeAPIError(ctx, addCustomRoleMembersResp.HTTPResponse, addCustomRoleMembersResp.Body)
		if diagnostic != nil {
			return nil, diag.Diagnostics{diagnostic}
		}
	}
	return newMemberIds, nil
}
