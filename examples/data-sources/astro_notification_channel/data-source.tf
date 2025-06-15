data "astro_notification_channel" "example_notification_channel" {
  id = "cm4nwrvyg024h01mk2dn58m5s"
}

# Output the API token using terraform apply
output "notification_channel" {
  value = data.astro_notification_channel.example_notification_channel
}