data "astro_notification_channels" "example_notification_channels" {}

data "astro_notification_channels" "notification_channel_ids_example" {
  notification_channel_ids = ["cm7dz31ye00cz01n1en2xilut"]
}

data "astro_notification_channels" "workspace_ids_example" {
  workspace_ids = ["clx42sxw501gl01o0gjenthnh"]
}

data "astro_notification_channels" "deployment_ids_example" {
  deployment_ids = ["clx44jyu001m201m5dzsbexqr"]
}

data "astro_notification_channels" "notification_channel_types_example" {
  notification_channel_types = ["SLACK", "EMAIL"]
}

data "astro_notification_channels" "entity_type_example" {
  entity_type = "DEPLOYMENT"
}

# Output the API tokens using terraform apply
output "notification_channels" {
  value = data.astro_notification_channels.example_notification_channels
}