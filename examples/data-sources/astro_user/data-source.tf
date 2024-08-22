data "astro_user" "example_user" {
  id = "clhpichn8002m01mqa4ocs7g6"
}

# Output the user value using terraform apply
output "user" {
  value = data.astro_user.example_user
}