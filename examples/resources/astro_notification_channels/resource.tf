# The astro_notification_channels resource manages many notification channels as a single resource.
# It batches the underlying API calls and automatically chunks requests that exceed the per-request
# limit (30), so you can declare any number of channels in one block. This avoids the rate limiting
# that large applies hit when creating channels one at a time.
#
# Channels are keyed by a stable, user-defined string. Changing a key is treated as deleting the old
# channel and creating a new one.

resource "astro_notification_channels" "team_channels" {
  notification_channels = {
    "oncall_email" = {
      name        = "On-call Email"
      type        = "EMAIL"
      entity_type = "DEPLOYMENT"
      entity_id   = "cm1zkps2a0cv301ph39benet6"
      definition = {
        recipients = ["oncall@example.com"]
      }
      is_shared = true
    }

    "alerts_slack" = {
      name        = "Alerts Slack"
      type        = "SLACK"
      entity_type = "DEPLOYMENT"
      entity_id   = "cm1zkps2a0cv301ph39benet6"
      definition = {
        webhook_url = "SLACK_WEBHOOK_URL"
      }
    }

    "pagerduty" = {
      name        = "PagerDuty"
      type        = "PAGERDUTY"
      entity_type = "DEPLOYMENT"
      entity_id   = "cm1zkps2a0cv301ph39benet6"
      definition = {
        integration_key = "PAGERDUTY_INTEGRATION_KEY"
      }
    }
  }
}

# Generate many channels programmatically — the resource chunks them across requests for you.
resource "astro_notification_channels" "per_team_email_channels" {
  notification_channels = {
    for team in toset(["data", "platform", "ml"]) :
    "${team}_email" => {
      name        = "${team} Email"
      type        = "EMAIL"
      entity_type = "DEPLOYMENT"
      entity_id   = "cm1zkps2a0cv301ph39benet6"
      definition = {
        recipients = ["${team}@example.com"]
      }
    }
  }
}
