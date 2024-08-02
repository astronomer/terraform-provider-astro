package schemas

import (
	"regexp"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func UserInviteResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"email": resourceSchema.StringAttribute{
			MarkdownDescription: "The email address of the user being invited",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile(validators.EmailString), "must be a valid email address"),
			},
		},
		"role": resourceSchema.StringAttribute{
			MarkdownDescription: "The Organization role to assign to the user",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(string(iam.ORGANIZATIONOWNER),
					string(iam.ORGANIZATIONMEMBER),
					string(iam.ORGANIZATIONBILLINGADMIN),
				),
			},
		},
		"expires_at": resourceSchema.StringAttribute{
			MarkdownDescription: "The expiration date of the invite",
			Computed:            true,
		},
		"invite_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the invite",
			Computed:            true,
		},
		"invitee": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The profile of the invitee",
			Computed:            true,
			Attributes:          ResourceSubjectProfileSchemaAttributes(),
		},
		"inviter": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The profile of the inviter",
			Computed:            true,
			Attributes:          ResourceSubjectProfileSchemaAttributes(),
		},
		"user_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the user",
			Computed:            true,
		},
	}
}
