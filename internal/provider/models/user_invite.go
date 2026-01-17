package models

import (
	"context"
	"time"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UserInvite describes the user_invite resource
type UserInvite struct {
	Email     types.String `tfsdk:"email"`
	Role      types.String `tfsdk:"role"`
	ExpiresAt types.String `tfsdk:"expires_at"`
	InviteId  types.String `tfsdk:"invite_id"`
	Invitee   types.Object `tfsdk:"invitee"`
	Inviter   types.Object `tfsdk:"inviter"`
	UserId    types.String `tfsdk:"user_id"`
}

func (data *UserInvite) ReadFromResponse(ctx context.Context, userInvite *iam.Invite, email string, role string) diag.Diagnostics {
	var diags diag.Diagnostics
	data.Email = types.StringValue(email)
	data.Role = types.StringValue(role)
	data.ExpiresAt = types.StringValue(userInvite.ExpiresAt.Format(time.RFC3339Nano))
	data.InviteId = types.StringValue(userInvite.InviteId)
	data.Invitee, diags = SubjectProfileTypesObject(ctx, userInvite.Invitee)
	if diags.HasError() {
		return diags
	}
	data.Inviter, diags = SubjectProfileTypesObject(ctx, userInvite.Inviter)
	if diags.HasError() {
		return diags
	}
	data.UserId = types.StringPointerValue(userInvite.UserId)
	return nil
}
