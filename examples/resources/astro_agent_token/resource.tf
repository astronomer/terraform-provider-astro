resource "astro_agent_token" "without_expiry" {
  deployment_id = "clx42kkcm01fo01o06agtmshg"
  name          = "my-agent-token-without-expiry"
  description   = "my agent token description"
}

resource "astro_agent_token" "with_expiry" {
  deployment_id         = "clx42kkcm01fo01o06agtmshg"
  name                  = "my-agent-token-with-expiry"
  description           = "my agent token description"
  expiry_period_in_days = 30
}

# Import an existing agent token
import {
  id = "<deployment_id>/<token_id>"
  to = astro_agent_token.imported_agent_token
}
resource "astro_agent_token" "imported_agent_token" {
  deployment_id = "clx42kkcm01fo01o06agtmshg"
  name          = "imported agent token"
}
