package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	// MaxAiSpendingCapCents is the highest cap the API takes, $100,000, in USD cents.
	MaxAiSpendingCapCents = 10_000_000
	// MaxAiSpendingCapSubjects is the most users and API tokens an organization can give a cap of
	// their own.
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
		"default_user_cap_cents": resourceSchema.Int64Attribute{
			MarkdownDescription: "The default monthly AI spending cap, in USD cents, for every user without a cap of " +
				"their own in `user_cap_cents`. At most `10000000` ($100,000). `0` turns AI off for those users. " +
				"If unset, the default is not managed by this resource. If it was set and is removed from the " +
				"configuration, the default is removed.",
			Optional:   true,
			Validators: capValue,
		},
		"default_api_token_cap_cents": resourceSchema.Int64Attribute{
			MarkdownDescription: "The default monthly AI spending cap, in USD cents, for every API token without a " +
				"cap of its own in `api_token_cap_cents`. Set independently of the user default. At most " +
				"`10000000` ($100,000). `0` turns AI off for those API tokens. If unset, the default is not managed " +
				"by this resource. If it was set and is removed from the configuration, the default is removed.",
			Optional:   true,
			Validators: capValue,
		},
		"user_cap_cents": resourceSchema.MapAttribute{
			MarkdownDescription: "The own monthly AI spending cap of each user, in USD cents, keyed by user ID. A " +
				"user's own cap replaces the user default for them, higher or lower; `0` turns AI off for the user. " +
				"At most `10000000` ($100,000) each. When set, this attribute is authoritative for the " +
				"organization's users: a user with a cap of their own that is not in the map has it removed on " +
				"apply, including one set in the Astro UI. Set it to `{}` to remove every user's own cap. If unset, " +
				"users' caps are not managed by this resource. If it was set and is removed from the configuration, " +
				"the caps it held are removed.",
			Optional:    true,
			ElementType: types.Int64Type,
			Validators:  capMap(),
		},
		"api_token_cap_cents": resourceSchema.MapAttribute{
			MarkdownDescription: "The own monthly AI spending cap of each API token, in USD cents, keyed by API token " +
				"ID. It works like `user_cap_cents`, for API tokens: authoritative when set, `0` turns AI off for the " +
				"API token, and not managed when unset.",
			Optional:    true,
			ElementType: types.Int64Type,
			Validators:  capMap(),
		},
	}
}
