data "astro_custom_role" "example" {
  id = "cmk64yvat027n01q7f9gn5ghg"
}

# Output the custom role value using terraform apply
output "custom_role" {
  value = data.astro_custom_role.example
}