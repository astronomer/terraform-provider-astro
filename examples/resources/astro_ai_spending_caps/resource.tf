# The astro_ai_spending_caps resource manages an organization's per-user and per-API-token Astro AI
# spending caps as a single resource. Amounts are monthly limits in USD cents.
#
# Every attribute is optional, and one that is not set is not managed. A map that is set is
# authoritative: a user or API token with a cap of its own that is not in it has the cap removed on
# apply, including one set in the Astro UI. Use at most one astro_ai_spending_caps resource.
#
# The AI spending caps API is a Labs feature that Astro turns on per organization.

resource "astro_ai_spending_caps" "org" {
  # The default for every user and every API token without a cap of their own.
  default_user_cap_cents      = 5000 # $50
  default_api_token_cap_cents = 2500 # $25

  # An own cap replaces the default for that user or API token, higher or lower, and 0 turns AI off
  # for it. The keys are user IDs and API token IDs.
  user_cap_cents = {
    "clm8q7f6q000008lcgyougpsk" = 20000 # $200
    "clm8q8c1x000108lc1w2h5p9d" = 0     # no AI for this user
  }

  api_token_cap_cents = {
    "clm8pxjjw000008l23jm08hyu" = 10000 # $100
  }
}
