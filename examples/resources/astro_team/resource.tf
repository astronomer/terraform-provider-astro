resource "astro_team" "team" {
  name              = "team"
  description       = "team description"
  member_ids        = ["clhpichn8002m01mqa4ocs7g6"]
  organization_role = "ORGANIZATION_OWNER"
  workspace_roles = [{
    workspace_id = "clx42sxw501gl01o0gjenthnh"
    role         = "WORKSPACE_OWNER"
  }]
  deployment_roles = [{
    deployment_id = "clyn6kxud003x01mtxmccegnh"
    role          = "DEPLOYMENT_ADMIN"
  }]
}

resource "astro_team" "team_with_no_optional_fields" {
  name              = "team"
  organization_role = "ORGANIZATION_OWNER"
}

# Import an existing team
import {
  id = "clx486hno068301il306nuhsm" # ID of the existing team
  to = astro_team.imported_team
}
resource "astro_team" "imported_team" {
  name              = "imported team"
  description       = "imported team description"
  member_ids        = ["clhpichn8002m01mqa4ocs7g6"]
  organization_role = "ORGANIZATION_OWNER"
  workspace_roles = [{
    workspace_id = "clx42sxw501gl01o0gjenthnh"
    role         = "WORKSPACE_OWNER"
  }]
  deployment_roles = [{
    deployment_id = "clyn6kxud003x01mtxmccegnh"
    role          = "DEPLOYMENT_ADMIN"
  }]
}