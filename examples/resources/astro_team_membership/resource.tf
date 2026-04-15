# Basic usage — add a single user to a team
resource "astro_team_membership" "single" {
  team_id = "clhpichn8002m01mqa4ocs7g6"
  user_id = "clv9xyzuserid0000000000000"
}

# Use for_each to manage a set of memberships from a variable
resource "astro_team_membership" "bulk" {
  for_each = toset(["clv9user1000000000000000", "clv9user2000000000000000"])

  team_id = "clhpichn8002m01mqa4ocs7g6"
  user_id = each.value
}

# Import an existing membership
import {
  id = "clhpichn8002m01mqa4ocs7g6/clv9xyzuserid0000000000000" # <team_id>/<user_id>
  to = astro_team_membership.single
}

# ─────────────────────────────────────────────────────────────────────────────
# Decentralized ownership example
#
# Variable declarations for the examples below
variable "platform_engineer_user_ids" {
  type        = set(string)
  description = "User IDs of platform engineers to add to the shared team"
}

variable "new_analytics_user_ids" {
  type        = set(string)
  description = "User IDs to onboard into the analytics workspace"
}

variable "platform_engineers_team_id" {
  type        = string
  description = "ID of the shared platform-engineers team (passed in from the platform module output)"
}

#
#
# A shared "platform-engineers" team is owned by the platform module.
# Each product team independently grants that team access to their own
# workspace — no coordination with the platform team required.
# ─────────────────────────────────────────────────────────────────────────────

# platform/main.tf — owned by the platform team
# Creates the shared team. Does NOT manage workspace or deployment roles.
resource "astro_team" "platform_engineers" {
  name              = "platform-engineers"
  organization_role = "ORGANIZATION_MEMBER"
  # No member_ids here — memberships are managed independently below
}

# platform/main.tf — platform team manages its own members
resource "astro_team_membership" "platform_member" {
  for_each = toset(var.platform_engineer_user_ids)

  team_id = astro_team.platform_engineers.id
  user_id = each.value
}

# analytics/main.tf — owned by the analytics team
# Grants the shared platform team access to the analytics workspace.
# This module has no dependency on the platform module's state.
resource "astro_team_membership" "analytics_onboarding" {
  for_each = toset(var.new_analytics_user_ids)

  team_id = var.platform_engineers_team_id # passed in as a variable
  user_id = each.value
}
