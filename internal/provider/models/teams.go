package models

import (
	"context"

	platform_v1 "github.com/astronomer/terraform-provider-astro/internal/clients/platform_v1"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Teams describes the data source data model.
type Teams struct {
	Teams types.Set `tfsdk:"teams"`
	Names types.Set `tfsdk:"names"` // query parameter
}

func (data *Teams) ReadFromResponse(ctx context.Context, teams []platform_v1.Team, teamsWithMembers map[string][]platform_v1.TeamMember) diag.Diagnostics {
	values := make([]attr.Value, len(teams))
	for i, team := range teams {
		var singleTeamData TeamDataSource
		var teamMembersPtr *[]platform_v1.TeamMember
		if members, exists := teamsWithMembers[team.Id]; exists {
			teamMembersPtr = &members
		}
		diags := singleTeamData.ReadFromResponse(ctx, &team, teamMembersPtr)
		if diags.HasError() {
			return diags
		}

		objectValue, diags := types.ObjectValueFrom(ctx, schemas.TeamsElementAttributeTypes(), singleTeamData)
		if diags.HasError() {
			return diags
		}
		values[i] = objectValue
	}
	var diags diag.Diagnostics
	data.Teams, diags = types.SetValue(types.ObjectType{AttrTypes: schemas.TeamsElementAttributeTypes()}, values)
	if diags.HasError() {
		return diags
	}

	return nil
}
