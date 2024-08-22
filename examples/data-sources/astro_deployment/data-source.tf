data "astro_deployment" "example" {
  id = "clozc036j01to01jrlgvueo8t"
}

# Output the deployment value using terraform apply
output "deployment" {
  value = data.astro_deployment.example
}