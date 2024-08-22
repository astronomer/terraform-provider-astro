data "astro_organization" "example" {}

# Output the organization value using terraform apply
output "organization" {
  value = data.astro_organization.example
}