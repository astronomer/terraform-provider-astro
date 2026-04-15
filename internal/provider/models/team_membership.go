package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// TeamMembership describes the team_membership resource
type TeamMembership struct {
	ID     types.String `tfsdk:"id"`
	TeamId types.String `tfsdk:"team_id"`
	UserId types.String `tfsdk:"user_id"`
}
