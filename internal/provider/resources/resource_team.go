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

var _ resource.Resource = &TeamResource{}
var _ resource.ResourceWithImportState = &TeamResource{}
var _ resource.ResourceWithConfigure = &TeamResource{}
var _ resource.ResourceWithValidateConfig = &TeamResource{}

func NewTeamResource() resource.Resource {
	return &TeamResource{}
}

// TeamResource defines the resource implementation.
type TeamResource struct {
	IamClient      *iam.ClientWithResponses
	PlatformClient *platform.ClientWithResponses
	OrganizationId string
}

func (r *TeamResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *TeamResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Team resource",
		Attributes:          schemas.TeamResourceSchemaAttributes(),
	}
}

func (r *TeamResource) Configure(
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

func (r *TeamResource) MutateRoles(
	ctx context.Context,
	data *models.TeamResource,
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

	// Checks for missing workspace roles if deployment roles are provided and adds them to the workspace roles
	missingWorkspaceIds, diags := common.ValidateWorkspaceDeploymentRoles(ctx, common.ValidateWorkspaceDeploymentRolesInput{
		PlatformClient:  r.PlatformClient,
		OrganizationId:  r.OrganizationId,
		WorkspaceRoles:  workspaceRoles,
		DeploymentRoles: deploymentRoles,
	})
	if diags.HasError() {
		return diags
	}
	if missingWorkspaceIds != nil && len(*missingWorkspaceIds) > 0 {
		for _, id := range *missingWorkspaceIds {
			workspaceRoles = append(workspaceRoles, iam.WorkspaceRole{
				Role:        iam.WORKSPACEACCESSOR,
				WorkspaceId: id,
			})
		}
	}

	// Update team roles
	updateTeamRolesRequest := iam.UpdateTeamRolesJSONRequestBody{
		DeploymentRoles:  &deploymentRoles,
		OrganizationRole: iam.UpdateTeamRolesRequestOrganizationRole(data.OrganizationRole.ValueString()),
		WorkspaceRoles:   &workspaceRoles,
	}
	teamRoles, err := r.IamClient.UpdateTeamRolesWithResponse(
		ctx,
		r.OrganizationId,
		teamId,
		updateTeamRolesRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to mutate Team roles", map[string]interface{}{"error": err})
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to mutate Team roles, got error: %s", err),
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

func (r *TeamResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.TeamResource

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
	createTeamRequest := iam.CreateTeamRequest{
		Name:             data.Name.ValueString(),
		Description:      data.Description.ValueStringPointer(),
		MemberIds:        &memberIds,
		OrganizationRole: lo.ToPtr(iam.CreateTeamRequestOrganizationRole(data.OrganizationRole.ValueString())),
	}

	// Create the team
	team, err := r.IamClient.CreateTeamWithResponse(
		ctx,
		r.OrganizationId,
		createTeamRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to create Team", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create Team, got error: %s", err),
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
			return
		}
	}

	// Get Team and use this as data since it will have the correct roles
	teamResp, err := r.IamClient.GetTeamWithResponse(
		ctx,
		r.OrganizationId,
		teamId,
	)
	if err != nil {
		tflog.Error(ctx, "failed to create Team", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create and get Team, got error: %s", err),
		)
		return
	}

	diags = data.ReadFromResponse(ctx, teamResp.JSON200, &memberIds)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created a Team resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TeamResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.TeamResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// get request
	team, err := r.IamClient.GetTeamWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
	)

	if err != nil {
		tflog.Error(ctx, "failed to get Team", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get Team, got error: %s", err),
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

	tflog.Trace(ctx, fmt.Sprintf("read a Team resource: %v", data.Id.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TeamResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data models.TeamResource

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
	newMemberIds, diags := r.UpdateTeamMembers(ctx, data)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Update team
	updateTeamRequest := iam.UpdateTeamRequest{
		Name: data.Name.ValueString(),
	}

	if !data.Description.IsNull() {
		updateTeamRequest.Description = data.Description.ValueStringPointer()
	} else {
		updateTeamRequest.Description = lo.ToPtr("")
	}

	team, err := r.IamClient.UpdateTeamWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
		updateTeamRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to update Team", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update Team, got error: %s", err),
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

	// Get Team and use this as data since it will have the correct roles
	teamResp, err := r.IamClient.GetTeamWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to update Team", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update and get Team, got error: %s", err),
		)
		return
	}

	diags = data.ReadFromResponse(ctx, teamResp.JSON200, &newMemberIds)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated a Team resource: %v", data.Id.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TeamResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.TeamResource

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// delete request
	team, err := r.IamClient.DeleteTeamWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to delete Team", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete Team, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, team.HTTPResponse, team.Body)
	// It is recommended to ignore 404 Resource Not Found errors when deleting a resource
	if statusCode != http.StatusNotFound && diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted a Team resource: %v", data.Id.ValueString()))
}

func (r *TeamResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *TeamResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var data models.TeamResource

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate workspace roles
	workspaceRoles, diags := common.RequestWorkspaceRoles(ctx, data.WorkspaceRoles)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	for _, role := range workspaceRoles {
		if !common.ValidateRoleMatchesEntityType(string(role.Role), string(iam.WORKSPACE)) {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Role '%s' is not valid for role type '%s'", string(role.Role), string(iam.WORKSPACE)),
				fmt.Sprintf("Please provide a valid role for the type '%s'", string(iam.WORKSPACE)),
			)
			return
		}
	}

	duplicateWorkspaceIds := common.GetDuplicateWorkspaceIds(workspaceRoles)
	if len(duplicateWorkspaceIds) > 0 {
		resp.Diagnostics.AddError(
			"Invalid Configuration: Cannot have multiple roles with the same workspace id",
			fmt.Sprintf("Please provide a unique workspace id for each role. The following workspace ids are duplicated: %v", duplicateWorkspaceIds),
		)
		return
	}

	// Validate deployment roles
	deploymentRoles, diags := common.RequestDeploymentRoles(ctx, data.DeploymentRoles)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	for _, role := range deploymentRoles {
		if !common.ValidateRoleMatchesEntityType(role.Role, string(iam.DEPLOYMENT)) {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Role '%s' is not valid for role type '%s'", role.Role, string(iam.DEPLOYMENT)),
				fmt.Sprintf("Please provide a valid role for the type '%s'", string(iam.DEPLOYMENT)),
			)
			return
		}
	}

	duplicateDeploymentIds := common.GetDuplicateDeploymentIds(deploymentRoles)
	if len(duplicateDeploymentIds) > 0 {
		resp.Diagnostics.AddError(
			"Invalid Configuration: Cannot have multiple roles with the same deployment id",
			fmt.Sprintf("Please provide unique deployment id for each role. The following deployment ids are duplicated: %v", duplicateDeploymentIds),
		)
		return
	}
}

func (r *TeamResource) CheckOrganizationIsScim(ctx context.Context) diag.Diagnostics {
	// Validate if org isScimEnabled and return error if it is
	org, err := r.PlatformClient.GetOrganizationWithResponse(ctx, r.OrganizationId, nil)
	if err != nil {
		tflog.Error(ctx, "failed to validate Team", map[string]interface{}{"error": err})
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Client Error",
				fmt.Sprintf("Unable to validate Team, got error: %s", err),
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
				"Invalid Configuration: Cannot create, update or delete a Team resource when SCIM is enabled",
				"Please disable SCIM in the organization settings to manage Team resources",
			),
		}
	}
	return nil
}

func (r *TeamResource) UpdateTeamMembers(ctx context.Context, data models.TeamResource) ([]string, diag.Diagnostics) {
	// get existing team members
	teamMembersResp, err := r.IamClient.ListTeamMembersWithResponse(
		ctx,
		r.OrganizationId,
		data.Id.ValueString(),
		nil,
	)
	if err != nil {
		tflog.Error(ctx, "failed to update Team", map[string]interface{}{"error": err})
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Client Error",
				fmt.Sprintf("Unable to list existing Team members, got error: %s", err),
			),
		}
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, teamMembersResp.HTTPResponse, teamMembersResp.Body)
	if diagnostic != nil {
		return nil, diag.Diagnostics{diagnostic}
	}

	teamMembers := teamMembersResp.JSON200.TeamMembers
	memberIds := lo.Map(teamMembers, func(tm iam.TeamMember, _ int) string {
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
			removeTeamMemberResp, err := r.IamClient.RemoveTeamMemberWithResponse(
				ctx,
				r.OrganizationId,
				data.Id.ValueString(),
				id,
			)
			if err != nil {
				tflog.Error(ctx, "failed to update Team", map[string]interface{}{"error": err})
				return nil, diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Client Error",
						fmt.Sprintf("Unable to remove Team member, got error: %s", err),
					),
				}
			}
			_, diagnostic = clients.NormalizeAPIError(ctx, removeTeamMemberResp.HTTPResponse, removeTeamMemberResp.Body)
			if diagnostic != nil {
				return nil, diag.Diagnostics{diagnostic}
			}
		}
	}

	// add the members that are in the new list
	if len(addIds) > 0 {
		addTeamMembersRequest := iam.AddTeamMembersRequest{
			MemberIds: addIds,
		}
		addTeamMembersResp, err := r.IamClient.AddTeamMembersWithResponse(
			ctx,
			r.OrganizationId,
			data.Id.ValueString(),
			addTeamMembersRequest,
		)
		if err != nil {
			tflog.Error(ctx, "failed to update Team", map[string]interface{}{"error": err})
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Client Error",
					fmt.Sprintf("Unable to add Team members, got error: %s", err),
				),
			}
		}
		_, diagnostic = clients.NormalizeAPIError(ctx, addTeamMembersResp.HTTPResponse, addTeamMembersResp.Body)
		if diagnostic != nil {
			return nil, diag.Diagnostics{diagnostic}
		}
	}
	return newMemberIds, nil
}
