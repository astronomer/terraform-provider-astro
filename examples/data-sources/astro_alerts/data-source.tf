data "astro_alerts" "example_alerts" {}

data "astro_alerts" "alert_ids_example" {
  alert_ids = ["cm4ntm56001gk01mbhudv1elv"]
}

data "astro_alerts" "workspace_ids_example" {
  workspace_ids = ["clx42sxw501gl01o0gjenthnh"]
}

data "astro_alerts" "deployment_ids_example" {
  deployment_ids = ["clx44jyu001m201m5dzsbexqr"]
}

data "astro_alerts" "alert_types_example" {
  alert_types = ["DAG_FAILURE", "DAG_SUCCESS"]
}

data "astro_alerts" "entity_type_example" {
  entity_type = "DEPLOYMENT"
}

# Output the API tokens using terraform apply
output "alerts" {
  value = data.astro_alerts.example_alerts
}