# The AI spending caps are a singleton per organization. Import them using the organization ID. The
# provider reads both defaults and every user's and API token's override. Import first when the
# organization already has caps that the configuration would change or remove: the create fails and
# asks for it. The next plan then shows what the configuration would change or remove.
terraform import astro_ai_spending_caps.org <organization_id>
