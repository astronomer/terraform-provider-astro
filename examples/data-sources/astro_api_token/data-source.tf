data "astro_api_token" "example" {
  id = "clxm4836f00ql01me3nigmcr6"
}

# Output the API token value using terraform apply
output "api_token" {
  value = data.astro_api_token.example
}