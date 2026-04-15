package resources

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &teamMembershipResource{}
var _ resource.ResourceWithImportState = &teamMembershipResource{}
var _ resource.ResourceWithConfigure = &teamMembershipResource{}

func NewTeamMembershipResource() resource.Resource {
	return &teamMembershipResource{}
}

type teamMembershipResource struct {
	iamClient      *iam.ClientWithResponses
	organizationId string
}

func (r *teamMembershipResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_team_membership"
}

func (r *teamMembershipResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a single user's membership in an Astro team. " +
			"Use this resource instead of the `member_ids` attribute on `astro_team` when you need to manage memberships " +
			"independently or across multiple state files.\n\n" +
			"**Conflict note:** Do not use both `astro_team.member_ids` and `astro_team_membership` for the same team. " +
			"Both write to the same API state and will conflict on apply.",
		Attributes: schemas.TeamMembershipResourceSchemaAttributes(),
	}
}

func (r *teamMembershipResource) Configure(
	ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
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

func (r *teamMembershipResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.TeamMembership

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamId := data.TeamId.ValueString()
	userId := data.UserId.ValueString()

	addResp, err := r.iamClient.AddTeamMembersWithResponse(
		ctx,
		r.organizationId,
		teamId,
		iam.AddTeamMembersJSONRequestBody{MemberIds: []string{userId}},
	)
	if err != nil {
		tflog.Error(ctx, "failed to add team member", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to add team member, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, addResp.HTTPResponse, addResp.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	data.ID = membershipID(teamId, userId)

	tflog.Trace(ctx, fmt.Sprintf("added user %v to team %v", userId, teamId))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamMembershipResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.TeamMembership

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamId := data.TeamId.ValueString()
	userId := data.UserId.ValueString()

	// Page through all members to handle teams with more than the default page size.
	pageSize := 1000
	offset := 0
	for {
		params := &iam.ListTeamMembersParams{
			Limit:  &pageSize,
			Offset: &offset,
		}
		membersResp, err := r.iamClient.ListTeamMembersWithResponse(
			ctx,
			r.organizationId,
			teamId,
			params,
		)
		if err != nil {
			tflog.Error(ctx, "failed to list team members", map[string]interface{}{"error": err})
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to list team members, got error: %s", err),
			)
			return
		}
		statusCode, diagnostic := clients.NormalizeAPIError(ctx, membersResp.HTTPResponse, membersResp.Body)
		if statusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		if diagnostic != nil {
			resp.Diagnostics.Append(diagnostic)
			return
		}
		if membersResp.JSON200 == nil {
			resp.Diagnostics.AddError("Client Error", "Unable to list team members, got nil response")
			return
		}

		for _, m := range membersResp.JSON200.TeamMembers {
			if m.UserId == userId {
				data.ID = membershipID(teamId, userId)
				tflog.Trace(ctx, fmt.Sprintf("read team_membership %v/%v", teamId, userId))
				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				return
			}
		}

		if membersResp.JSON200.TotalCount <= offset+len(membersResp.JSON200.TeamMembers) {
			break
		}
		offset += pageSize
	}

	// Member no longer exists — remove from state
	resp.State.RemoveResource(ctx)
}

// Update is a no-op: both team_id and user_id are RequiresReplace, so any change recreates.
func (r *teamMembershipResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
}

func (r *teamMembershipResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.TeamMembership

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamId := data.TeamId.ValueString()
	userId := data.UserId.ValueString()

	removeResp, err := r.iamClient.RemoveTeamMemberWithResponse(
		ctx,
		r.organizationId,
		teamId,
		userId,
	)
	if err != nil {
		tflog.Error(ctx, "failed to remove team member", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to remove team member, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, removeResp.HTTPResponse, removeResp.Body)
	if statusCode != http.StatusNotFound && diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("removed user %v from team %v", userId, teamId))
}

func (r *teamMembershipResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Import ID format: <team_id>/<user_id>
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in the format `<team_id>/<user_id>`",
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// membershipID returns the composite state ID for a team membership.
func membershipID(teamId, userId string) types.String {
	return types.StringValue(teamId + "/" + userId)
}
