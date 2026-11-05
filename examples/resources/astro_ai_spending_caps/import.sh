# The AI spending caps are a singleton per organization. Import them using the organization ID. The
# provider reads both defaults, when they are set, and every user's and API token's own cap into the
# resource, so the configuration then has to list the defaults and caps you keep: an attribute
# missing from it is removed on the next apply.
terraform import astro_ai_spending_caps.org <organization_id>
