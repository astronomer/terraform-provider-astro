package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	// MaxAiSpendingCapCents is the highest cap the API takes, $100,000, in USD cents.
	MaxAiSpendingCapCents = 10_000_000
	// MaxAiSpendingCapSubjects is the most users and API tokens an organization can give an
	// override.
	MaxAiSpendingCapSubjects = 1000
)

// AiSpendingCapsResourceSchemaAttributes returns the attributes for the astro_ai_spending_caps
// resource.
func AiSpendingCapsResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	capValue := []validator.Int64{int64validator.Between(0, MaxAiSpendingCapCents)}
	capMap := func() []validator.Map {
		return []validator.Map{
			mapvalidator.SizeAtMost(MaxAiSpendingCapSubjects),
			mapvalidator.KeysAre(validators.IsCuid()),
			mapvalidator.ValueInt64sAre(int64validator.Between(0, MaxAiSpendingCapCents)),
		}
	}

	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the organization whose AI spending caps this resource manages. " +
				"This is also the ID used to import the resource.",
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		// The object is Optional+Computed, so leaving it out keeps its state value and changes nothing.
		// Its fields are Optional only, so a field left out of a set object is null, which removes it.
		"default_cap_cents": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The default monthly AI spending caps, in USD cents, for users and API tokens without an " +
				"override. When set, it is authoritative for both defaults: a field left out removes that default (users " +
				"or API tokens without an override then have no default limit), and `{}` removes both. When left out, the " +
				"defaults are not changed.",
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
			Attributes: map[string]resourceSchema.Attribute{
				"user": resourceSchema.Int64Attribute{
					MarkdownDescription: "The default for every user without an override. At most `10000000` ($100,000). " +
						"`0` turns AI off for those users.",
					Optional:   true,
					Validators: capValue,
				},
				"api_token": resourceSchema.Int64Attribute{
					MarkdownDescription: "The default for every API token without an override. At most `10000000` " +
						"($100,000). `0` turns AI off for those API tokens.",
					Optional:   true,
					Validators: capValue,
				},
			},
		},
		"user_cap_cents": resourceSchema.MapAttribute{
			MarkdownDescription: "The override of each user, a monthly AI spending cap in USD cents, keyed by user ID. A " +
				"user's override replaces the user default for them, higher or lower; `0` turns AI off for the user. " +
				"At most `10000000` ($100,000) each. When set, it is authoritative for the organization's users: a " +
				"user with an override that is not in the map has it removed on apply, including one set in the Astro " +
				"UI, and `{}` removes every user's override. When left out, users' overrides are not changed.",
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.Map{mapplanmodifier.UseStateForUnknown()},
			ElementType:   types.Int64Type,
			Validators:    capMap(),
		},
		"api_token_cap_cents": resourceSchema.MapAttribute{
			MarkdownDescription: "The override of each API token, a monthly AI spending cap in USD cents, keyed by API token " +
				"ID. It works like `user_cap_cents`, for API tokens: authoritative when set, `0` turns AI off for the " +
				"API token, `{}` removes every API token's override, and not changed when left out.",
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.Map{mapplanmodifier.UseStateForUnknown()},
			ElementType:   types.Int64Type,
			Validators:    capMap(),
		},
	}
}
