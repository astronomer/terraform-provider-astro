package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func TeamMembershipResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Unique identifier for this membership (format: `<team_id>/<user_id>`)",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"team_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the team",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				validators.IsCuid(),
			},
		},
		"user_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the user to add to the team",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				validators.IsCuid(),
			},
		},
	}
}
