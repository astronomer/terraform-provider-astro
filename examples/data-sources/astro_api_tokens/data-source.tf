data "astro_api_tokens" "example_api_tokens" {}

data "astro_api_tokens" "organization_only_example" {
  include_only_organization_tokens = true
}

data "astro_api_tokens" "workspace_example" {
  workspace_id = "clx42sxw501gl01o0gjenthnh"
}

data "astro_api_tokens" "deployment_example" {
  deployment_id = "clx44jyu001m201m5dzsbexqr"
}

# Output the API tokens using terraform apply
output "api_tokens" {
  value = data.astro_api_tokens.example_api_tokens
}