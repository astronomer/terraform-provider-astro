data "astro_environment_object" "example_environment_object" {
  id = "cm4ntm56001gk01mbhudv1elv"
}

# Output the environment object using terraform apply
output "environment_object" {
  value = data.astro_environment_object.example_environment_object
}
