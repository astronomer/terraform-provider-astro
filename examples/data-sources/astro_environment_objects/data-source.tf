data "astro_environment_objects" "example_environment_objects" {}

data "astro_environment_objects" "workspace_example" {
  workspace_id = "clx42sxw501gl01o0gjenthnh"
}

data "astro_environment_objects" "deployment_example" {
  deployment_id  = "clx44jyu001m201m5dzsbexqr"
  resolve_linked = true
}

data "astro_environment_objects" "connections_example" {
  workspace_id = "clx42sxw501gl01o0gjenthnh"
  object_type  = "CONNECTION"
}

data "astro_environment_objects" "variables_example" {
  workspace_id = "clx42sxw501gl01o0gjenthnh"
  object_type  = "AIRFLOW_VARIABLE"
}

data "astro_environment_objects" "metrics_exports_example" {
  workspace_id = "clx42sxw501gl01o0gjenthnh"
  object_type  = "METRICS_EXPORT"
}

data "astro_environment_objects" "object_key_example" {
  workspace_id = "clx42sxw501gl01o0gjenthnh"
  object_key   = "warehouse_postgres"
}

data "astro_environment_objects" "with_secrets_example" {
  workspace_id = "clx42sxw501gl01o0gjenthnh"
  show_secrets = true
}

# Output the environment objects using terraform apply
output "environment_objects" {
  value = data.astro_environment_objects.example_environment_objects
}
