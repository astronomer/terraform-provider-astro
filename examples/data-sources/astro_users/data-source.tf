data "astro_users" "example_users" {}

data "astro_users" "example_users_filter_by_workspace_id" {
  workspace_id = "clx42sxw501gl01o0gjenthnh"
}

data "astro_users" "example_users_filter_by_deployment_id" {
  deployment_id = "clx44jyu001m201m5dzsbexqr"
}
