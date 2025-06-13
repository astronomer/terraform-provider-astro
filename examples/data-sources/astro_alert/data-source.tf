data "astro_alert" "example_alert" {
  id = "cm4ntm56001gk01mbhudv1elv"
}

# Output the API token using terraform apply
output "alert" {
  value = data.astro_alert.example_alert
}