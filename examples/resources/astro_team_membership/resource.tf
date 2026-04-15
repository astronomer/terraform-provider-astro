resource "astro_team_membership" "example" {
  team_id = "clv9xyzteamid0000000000000"
  user_id = "clv9xyzuserid0000000000000"
}

# Use for_each to manage multiple memberships independently
resource "astro_team_membership" "engineers" {
  for_each = toset(var.engineer_user_ids)

  team_id = astro_team.platform.id
  user_id = each.value
}
