data "astro_teams" "example_teams" {}

data "astro_teams" "example_teams_filter_by_names" {
  names = ["my first team", "my second team"]
}

# Output the teams value using terraform apply
output "example_teams" {
  value = data.astro_teams.example_teams
}