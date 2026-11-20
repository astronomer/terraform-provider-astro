# The astro_ai_spending_caps resource manages an organization's per-user and per-API-token Astro AI
# spending caps as a single resource. Amounts are monthly limits in USD cents.
#
# Each of default_cap_cents, user_cap_cents and api_token_cap_cents is managed only when it is set.
# Left out, it is not changed. Set, it is the complete list for what it covers: a default left out
# of default_cap_cents, or an override not in a map, is removed on apply, including ones set in the
# Astro UI. {} removes everything it covers. Destroying the resource changes nothing in Astro. Use at
# most one astro_ai_spending_caps resource per organization.
#
# The AI spending caps API is a Labs feature that Astro turns on per organization.

resource "astro_ai_spending_caps" "org" {
  # The default for every user and every API token without an override.
  default_cap_cents = {
    user      = 5000 # $50
    api_token = 2500 # $25
  }

  # An override replaces the default for that user or API token, higher or lower, and 0 turns AI off
  # for it. The keys are user IDs and API token IDs.
  user_cap_cents = {
    "clm8q7f6q000008lcgyougpsk" = 20000 # $200
    "clm8q8c1x000108lc1w2h5p9d" = 0     # no AI for this user
  }

  # api_token_cap_cents is left out, so API tokens' overrides are not changed.
}
