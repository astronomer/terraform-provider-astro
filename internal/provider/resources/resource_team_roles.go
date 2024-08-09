package resources

import (
	"context"
	"fmt"
	"net/http"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"

	"github.com/astronomer/terraform-provider-astro/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/samber/lo"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &teamRolesResource{}
var _ resource.ResourceWithImportState = &teamRolesResource{}
var _ resource.ResourceWithConfigure = &teamRolesResource{}

func NewTeamRolesResource() resource.Resource {
	return &teamRolesResource{}
}

// teamRolesResource defines the resource implementation.
type teamRolesResource struct {
	iamClient      *iam.ClientWithResponses
	platformClient *platform.ClientWithResponses
	organizationId string
}

func (r *teamRolesResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_team_roles"
}

func (r *teamRolesResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Team Roles resource",
		Attributes:          schemas.ResourceTeamRolesSchemaAttributes(),
	}
}

func (r *teamRolesResource) Configure(
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
	r.platformClient = apiClients.PlatformClient
	r.organizationId = apiClients.OrganizationId
}

func (r *teamRolesResource) MutateRoles(
	ctx context.Context,
	data *models.TeamRoles,
) diag.Diagnostics {
	teamId := data.TeamId.ValueString()

	// Then convert the models to the request types for the API
	workspaceRoles, diags := common.RequestWorkspaceRoles(ctx, data.WorkspaceRoles)
	if diags.HasError() {
		return diags
	}
	deploymentRoles, diags := common.RequestDeploymentRoles(ctx, data.DeploymentRoles)
	if diags.HasError() {
		return diags
	}

	// Validate the roles
	diags = common.ValidateWorkspaceDeploymentRoles(ctx, common.ValidateWorkspaceDeploymentRolesInput{
		PlatformClient:  r.platformClient,
		OrganizationId:  r.organizationId,
		WorkspaceRoles:  workspaceRoles,
		DeploymentRoles: deploymentRoles,
	})
	if diags.HasError() {
		return diags
	}

	// create request
	updateTeamRolesRequest := iam.UpdateTeamRolesJSONRequestBody{
		DeploymentRoles:  &deploymentRoles,
		OrganizationRole: iam.UpdateTeamRolesRequestOrganizationRole(data.OrganizationRole.ValueString()),
		WorkspaceRoles:   &workspaceRoles,
	}
	teamRoles, err := r.iamClient.UpdateTeamRolesWithResponse(
		ctx,
		r.organizationId,
		teamId,
		updateTeamRolesRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to mutate team_roles", map[string]interface{}{"error": err})
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to mutate team_roles, got error: %s", err),
		)
		return diags
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, teamRoles.HTTPResponse, teamRoles.Body)
	if diagnostic != nil {
		diags.Append(diagnostic)
		return diags
	}

	diags = data.ReadFromResponse(ctx, teamId, teamRoles.JSON200)
	if diags.HasError() {
		return diags
	}

	return nil
}

func (r *teamRolesResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.TeamRoles

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
	tflog.Trace(ctx, fmt.Sprintf("created a team_roles resource for team '%v'", data.TeamId.ValueString()))
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamRolesResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.TeamRoles

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamId := data.TeamId.ValueString()

	// get request
	teamRoles, err := r.iamClient.GetTeamWithResponse(
		ctx,
		r.organizationId,
		teamId,
	)
	if err != nil {
		tflog.Error(ctx, "failed to get team_roles", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get team_roles, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, teamRoles.HTTPResponse, teamRoles.Body)
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

	// Generate subjectRoles from the get team API response
	subjectRoles := iam.SubjectRoles{
		OrganizationRole: lo.ToPtr(iam.SubjectRolesOrganizationRole(teamRoles.JSON200.OrganizationRole)),
		WorkspaceRoles:   teamRoles.JSON200.WorkspaceRoles,
		DeploymentRoles:  teamRoles.JSON200.DeploymentRoles,
	}
	diags := data.ReadFromResponse(ctx, teamId, &subjectRoles)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read a team_roles resource: %v", teamId))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamRolesResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data models.TeamRoles

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
	tflog.Trace(ctx, fmt.Sprintf("updated a team_roles resource for team '%v'", data.TeamId.ValueString()))
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamRolesResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.TeamRoles

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// delete request
	teamId := data.TeamId.ValueString()

	// update request with no workspace roles, no deployment roles and lowest organization role
	updateTeamRolesRequest := iam.UpdateTeamRolesJSONRequestBody{
		DeploymentRoles:  nil,
		OrganizationRole: iam.UpdateTeamRolesRequestOrganizationRole(iam.ORGANIZATIONMEMBER),
		WorkspaceRoles:   nil,
	}
	teamRoles, err := r.iamClient.UpdateTeamRolesWithResponse(
		ctx,
		r.organizationId,
		teamId,
		updateTeamRolesRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to delete team_roles", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete team_roles, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, teamRoles.HTTPResponse, teamRoles.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	diags := data.ReadFromResponse(ctx, teamId, teamRoles.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted a team_roles resource for team '%v'", teamId))
}

func (r *teamRolesResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("team_id"), req, resp)
}
