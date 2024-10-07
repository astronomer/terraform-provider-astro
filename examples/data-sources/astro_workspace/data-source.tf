data "astro_workspace" "example_workspace" {
  id = "clozc036j01to01jrlgvueo8t"
}

# Output the workspace value using terraform apply
output "workspace" {
  value = data.astro_workspace.example_workspace
}