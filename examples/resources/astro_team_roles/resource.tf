resource "astro_team_roles" "organization_role_only" {
  team_id           = "clnp86ly5000401ndaga21g81"
  organization_role = "ORGANIZATION_MEMBER"
}

resource "astro_team_roles" "workspace_roles" {
  team_id           = "clnp86ly5000401ndaga21g81"
  organization_role = "ORGANIZATION_MEMBER"
  workspace_roles = [
    {
      workspace_id = "clwp86ly5000401ndaga21g85"
      role         = "WORKSPACE_ACCESSOR"
    },
    {
      workspace_id = "clwp86ly5000401ndaga21g82"
      role         = "WORKSPACE_MEMBER"
    }
  ]
}

resource "astro_team_roles" "deployment_roles" {
  team_id           = "clnp86ly5000401ndaga21g81"
  organization_role = "ORGANIZATION_MEMBER"
  deployment_roles = [
    {
      deployment_id = "cldp86ly5000401ndaga21g86"
      role          = "DEPLOYMENT_ADMIN"
    }
  ]
}

resource "astro_team_roles" "all_roles" {
  team_id           = "clnp86ly5000401ndaga21g81"
  organization_role = "ORGANIZATION_MEMBER"
  workspace_roles = [
    {
      workspace_id = "clwp86ly5000401ndaga21g85"
      role         = "WORKSPACE_OWNER"
    },
    {
      workspace_id = "clwp86ly5000401ndaga21g82"
      role         = "WORKSPACE_MEMBER"
    }
  ]
  deployment_roles = [
    {
      deployment_id = "cldp86ly5000401ndaga21g86"
      role          = "my custom role"
    }
  ]
}

// Import existing team roles
import {
  id = "clnp86ly5000401ndaga21g81" // ID of the existing team
  to = astro_team_roles.imported_team_roles
}
resource "astro_team_roles" "imported_team_roles" {
  team_id           = "clnp86ly5000401ndaga21g81"
  organization_role = "ORGANIZATION_MEMBER"
  workspace_roles = [
    {
      workspace_id = "clwp86ly5000401ndaga21g85"
      role         = "WORKSPACE_OWNER"
    }
  ]
}