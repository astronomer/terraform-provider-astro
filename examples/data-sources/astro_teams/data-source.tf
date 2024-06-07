data "astro_teams" "example_teams" {}

data "astro_teams" "example_teams_filter_by_names" {
  names = ["my first team", "my second team"]
}
