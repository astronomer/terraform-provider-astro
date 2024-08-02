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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

var _ resource.Resource = &UserInviteResource{}
var _ resource.ResourceWithConfigure = &UserInviteResource{}

func NewUserInviteResource() resource.Resource {
	return &UserInviteResource{}
}

// UserInviteResource defines the resource implementation.
type UserInviteResource struct {
	IamClient      *iam.ClientWithResponses
	OrganizationId string
}

func (r *UserInviteResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_user_invite"
}

func (r *UserInviteResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "User Invite resource",
		Attributes:          schemas.UserInviteResourceSchemaAttributes(),
	}
}

func (r *UserInviteResource) Configure(
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

func (r *UserInviteResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data models.UserInvite

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var diags diag.Diagnostics

	// Create the user invite request
	createUserInviteRequest := iam.CreateUserInviteRequest{
		InviteeEmail: data.Email.ValueString(),
		Role:         iam.CreateUserInviteRequestRole(data.Role.ValueString()),
	}

	// Create the user invite
	userInvite, err := r.IamClient.CreateUserInviteWithResponse(
		ctx,
		r.OrganizationId,
		createUserInviteRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to create User Invite", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create User Invite, got error: %s", err),
		)
		return
	}
	_, diagnostic := clients.NormalizeAPIError(ctx, userInvite.HTTPResponse, userInvite.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	diags = data.ReadFromResponse(ctx, userInvite.JSON200, data.Email.ValueString(), data.Role.ValueString())
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created a User Invite resource: %v", data.InviteId.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserInviteResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data models.UserInvite

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Extract the invitee from the invitee object
	var invitee models.SubjectProfile
	if !data.Invitee.IsUnknown() && !data.Invitee.IsNull() {
		diags := data.Invitee.As(ctx, &invitee, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    false,
			UnhandledUnknownAsEmpty: false,
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Extract the inviter from the inviter object
	var inviter models.SubjectProfile
	if !data.Inviter.IsUnknown() && !data.Inviter.IsNull() {
		diags := data.Inviter.As(ctx, &inviter, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    false,
			UnhandledUnknownAsEmpty: false,
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Get the user invite
	user, err := r.IamClient.GetUserWithResponse(
		ctx,
		r.OrganizationId,
		invitee.Id.ValueString(),
	)
	if err != nil {
		tflog.Error(ctx, "failed to get User Invite", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get User Invite, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, user.HTTPResponse, user.Body)
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

	// Generate userInvite from the get user API response
	userInvite := iam.Invite{
		ExpiresAt: data.ExpiresAt.ValueString(),
		InviteId:  data.InviteId.ValueString(),
		Invitee: iam.BasicSubjectProfile{
			ApiTokenName: invitee.ApiTokenName.ValueStringPointer(),
			AvatarUrl:    invitee.AvatarUrl.ValueStringPointer(),
			FullName:     invitee.FullName.ValueStringPointer(),
			Id:           invitee.Id.ValueString(),
			SubjectType:  lo.ToPtr(iam.BasicSubjectProfileSubjectType(invitee.SubjectType.ValueString())),
			Username:     invitee.Username.ValueStringPointer(),
		},
		Inviter: iam.BasicSubjectProfile{
			ApiTokenName: inviter.ApiTokenName.ValueStringPointer(),
			AvatarUrl:    inviter.AvatarUrl.ValueStringPointer(),
			FullName:     inviter.FullName.ValueStringPointer(),
			Id:           inviter.Id.ValueString(),
			SubjectType:  lo.ToPtr(iam.BasicSubjectProfileSubjectType(inviter.SubjectType.ValueString())),
			Username:     inviter.Username.ValueStringPointer(),
		},
		OrganizationId: r.OrganizationId,
		UserId:         lo.ToPtr(user.JSON200.Id),
	}

	diags := data.ReadFromResponse(ctx, &userInvite, data.Email.ValueString(), data.Role.ValueString())
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read a User Invite resource: %v", data.InviteId.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the User Invite resource by deleting the existing user invite and creating a new user invite.
func (r *UserInviteResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data models.UserInvite

	// Delete existing user invite

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var diags diag.Diagnostics

	existingInviteId := data.InviteId.ValueString()

	// Delete the existing user invite
	deletedUserInvite, err := r.IamClient.DeleteUserInviteWithResponse(
		ctx,
		r.OrganizationId,
		existingInviteId,
	)
	if err != nil {
		tflog.Error(ctx, "failed to update User Invite", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update and delete User Invite, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, deletedUserInvite.HTTPResponse, deletedUserInvite.Body)
	// It is recommended to ignore 404 Resource Not Found errors when deleting a resource
	if statusCode != http.StatusNotFound && diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	// Create a new user invite

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new user invite request
	createUserInviteRequest := iam.CreateUserInviteRequest{
		InviteeEmail: data.Email.ValueString(),
		Role:         iam.CreateUserInviteRequestRole(data.Role.ValueString()),
	}

	// Create the new user invite
	userInvite, err := r.IamClient.CreateUserInviteWithResponse(
		ctx,
		r.OrganizationId,
		createUserInviteRequest,
	)
	if err != nil {
		tflog.Error(ctx, "failed to update User Invite", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update and delete User Invite, got error: %s", err),
		)
		return
	}
	_, diagnostic = clients.NormalizeAPIError(ctx, userInvite.HTTPResponse, userInvite.Body)
	if diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	diags = data.ReadFromResponse(ctx, userInvite.JSON200, data.Email.ValueString(), data.Role.ValueString())
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated a User Invite resource: %v", data.InviteId.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserInviteResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data models.UserInvite

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	existingInviteId := data.InviteId.ValueString()

	// delete the old existing user invite
	deletedUserInvite, err := r.IamClient.DeleteUserInviteWithResponse(
		ctx,
		r.OrganizationId,
		existingInviteId,
	)
	if err != nil {
		tflog.Error(ctx, "failed to delete User Invite", map[string]interface{}{"error": err})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete User Invite, got error: %s", err),
		)
		return
	}
	statusCode, diagnostic := clients.NormalizeAPIError(ctx, deletedUserInvite.HTTPResponse, deletedUserInvite.Body)
	// It is recommended to ignore 404 Resource Not Found errors when deleting a resource
	if statusCode != http.StatusNotFound && diagnostic != nil {
		resp.Diagnostics.Append(diagnostic)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted a User Invite resource: %v", data.InviteId.ValueString()))
}
