data "astronomer_deployments" "example_deployments" {}

data "astronomer_deployments" "example_deployments_filter_by_deployment_ids" {
  deployment_ids = ["clozc036j01to01jrlgvueo8t", "clozc036j01to01jrlgvueo81"]
}

data "astronomer_deployments" "example_deployments_filter_by_workspace_ids" {
  workspace_ids = ["clozc036j01to01jrlgvueo8t", "clozc036j01to01jrlgvueo81"]
}

data "astronomer_deployments" "example_deployments_filter_by_names" {
  names = ["my first deployment", "my second deployment"]
}