data "astro_teams" "example_teams" {}

data "astro_teams" "example_teams_filter_by_team_ids" {
  team_ids = ["clozc036j01to01jrlgvueo8t", "clozc036j01to01jrlgvueo81"]
}

data "astro_teams" "example_teams_filter_by_names" {
  names = ["my first team", "my second team"]
}
