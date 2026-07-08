# astro_users_list is identical to astro_users but returns the `users`
# collection as an ordered list instead of a set, for significantly better
# `terraform plan` performance on large organizations.

data "astro_users_list" "example_users" {}

data "astro_users_list" "example_users_filter_by_workspace_id" {
  workspace_id = "clx42sxw501gl01o0gjenthnh"
}

data "astro_users_list" "example_users_filter_by_deployment_id" {
  deployment_id = "clx44jyu001m201m5dzsbexqr"
}

# Output the users value using terraform apply
output "example_users_list" {
  value = data.astro_users_list.example_users
}
