data "astro_workspaces" "example_workspaces" {}

data "astro_workspaces" "example_workspaces_filter_by_workspace_ids" {
  workspace_ids = ["clozc036j01to01jrlgvueo8t", "clozc036j01to01jrlgvueo81"]
}

data "astro_workspaces" "example_workspaces_filter_by_names" {
  names = ["my first workspace", "my second workspace"]
}

# Output the workspaces value using terraform apply
output "example_workspaces" {
  value = data.astro_workspaces.example_workspaces
}