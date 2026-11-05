package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// AiSpendingCapsResource describes the astro_ai_spending_caps resource data model. It manages the
// organization's per-user and per-API-token AI spending caps: the two defaults, and the override of
// each user and API token. Every attribute but id is optional, and an attribute that is not set is
// not managed.
type AiSpendingCapsResource struct {
	Id                      types.String `tfsdk:"id"`
	DefaultUserCapCents     types.Int64  `tfsdk:"default_user_cap_cents"`
	DefaultApiTokenCapCents types.Int64  `tfsdk:"default_api_token_cap_cents"`
	UserCapCents            types.Map    `tfsdk:"user_cap_cents"`
	ApiTokenCapCents        types.Map    `tfsdk:"api_token_cap_cents"`
}
