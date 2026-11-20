package models

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AiSpendingCapDefaults is the default_cap_cents object.
type AiSpendingCapDefaults struct {
	User     types.Int64 `tfsdk:"user"`
	ApiToken types.Int64 `tfsdk:"api_token"`
}

var AiSpendingCapDefaultsAttrTypes = map[string]attr.Type{"user": types.Int64Type, "api_token": types.Int64Type}

// AiSpendingCapsResource describes the astro_ai_spending_caps resource data model. It manages the
// organization's per-user and per-API-token AI spending caps: the two defaults, and the override of
// each user and API token. Every attribute but id is optional, and an attribute that is not set is
// not managed.
type AiSpendingCapsResource struct {
	Id               types.String `tfsdk:"id"`
	DefaultCapCents  types.Object `tfsdk:"default_cap_cents"`
	UserCapCents     types.Map    `tfsdk:"user_cap_cents"`
	ApiTokenCapCents types.Map    `tfsdk:"api_token_cap_cents"`
}
