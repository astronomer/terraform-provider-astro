data "astro_deployments" "example_deployments" {}

data "astro_deployments" "example_deployments_filter_by_names" {
  names = ["my deployment"]
}

data "astro_deployments" "example_deployments_filter_by_deployment_ids" {
  deployment_ids = ["clozc036j01to01jrlgvueo8t"]
}

data "astro_deployments" "example_deployments_filter_by_workspace_ids" {
  workspace_ids = ["clozc036j01to01jrlgvu798d"]
}