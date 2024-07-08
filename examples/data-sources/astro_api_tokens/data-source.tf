data "astro_api_tokens" "example" {}

data "astro_api_tokens" "org_only_example" {
  include_only_organization_tokens = true
}

data "astro_api_tokens" "workspace_example" {
  workspace_id = "clx42sxw501gl01o0gjenthnh"
}

data "astro_api_tokens" "deployment_example" {
  deployment_id = "clx44jyu001m201m5dzsbexqr"
}