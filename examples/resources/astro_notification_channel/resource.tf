resource "astro_notification_channel" "email_notification_channel" {
  name        = "Email Notification Channel"
  type        = "EMAIL"
  entity_type = "DEPLOYMENT"
  entity_id   = "cm1zkps2a0cv301ph39benet6"
  definition = {
    recipients = ["test@gmail.com"]
  }
  is_shared = true
}

resource "astro_notification_channel" "slack_notification_channel" {
  name        = "Slack Notification Channel"
  type        = "SLACK"
  entity_type = "DEPLOYMENT"
  entity_id   = "cm1zkps2a0cv301ph39benet6"
  definition = {
    webhook_url = "SLACK_WEBHOOK_URL"
  }
  is_shared = true
}

resource "astro_notification_channel" "pagerduty_notification_channel" {
  name        = "PagerDuty Notification Channel"
  type        = "PAGERDUTY"
  entity_type = "DEPLOYMENT"
  entity_id   = "cm1zkps2a0cv301ph39benet6"
  definition = {
    integration_key = "PAGERDUTY_INTEGRATION_KEY"
  }
  is_shared = true
}

resource "astro_notification_channel" "opsgenie_notification_channel" {
  name        = "OpsGenie Notification Channel"
  type        = "OPSGENIE"
  entity_type = "DEPLOYMENT"
  entity_id   = "cm1zkps2a0cv301ph39benet6"
  definition = {
    api_key = "OPSGENIE_API_KEY"
  }
  is_shared = true
}

resource "astro_notification_channel" "dag_trigger_notification_channel" {
  name        = "DAG Trigger Notification Channel"
  type        = "DAG_TRIGGER"
  entity_type = "DEPLOYMENT"
  entity_id   = "cm1zkps2a0cv301ph39benet6"
  definition = {
    dag_id               = "example_dag_id"
    deployment_id        = "cm1zkps2a0cv301ph39benet6"
    deployment_api_token = "example_api_token"
  }
  is_shared = true
}

# Import an existing notification channel
import {
  id = "cm4ntm56001gk01mbhudv1elv"
  to = astro_notification_channel.email_notification_channel
}

resource "astro_notification_channel" "example_notification_channel" {
  name        = "Example Notification Channel"
  type        = "EMAIL"
  entity_type = "DEPLOYMENT"
  entity_id   = "cm1zkps2a0cv301ph39benet6"
  definition = {
    recipients = ["test@gmail.com"]
  }
  is_shared = true
}