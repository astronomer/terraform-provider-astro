resource "astro_team" "example" {
  name              = "team"
  description       = "team-description"
  member_ids        = ["clhpichn8002m01mqa4ocs7g6"]
  organization_role = "ORGANIZATION_OWNER"
}

resource "astro_team" "example_with_no_optional_fields" {
  name = "team"
}

