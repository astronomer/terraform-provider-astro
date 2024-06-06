---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "astro_teams Data Source - astro"
subcategory: ""
description: |-
  Teams data source
---

# astro_teams (Data Source)

Teams data source

## Example Usage

```terraform
data "astro_teams" "example_teams" {}

data "astro_teams" "example_teams_filter_by_team_ids" {
  team_ids = ["clozc036j01to01jrlgvueo8t", "clozc036j01to01jrlgvueo81"]
}

data "astro_teams" "example_teams_filter_by_names" {
  names = ["my first team", "my second team"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `names` (Set of String)

### Read-Only

- `teams` (Attributes Set) (see [below for nested schema](#nestedatt--teams))

<a id="nestedatt--teams"></a>
### Nested Schema for `teams`

Required:

- `id` (String) Team identifier

Read-Only:

- `created_at` (String) Workspace creation timestamp
- `created_by` (Attributes) Workspace creator (see [below for nested schema](#nestedatt--teams--created_by))
- `deployment_roles` (Attributes Set) The roles to assign to the deployments (see [below for nested schema](#nestedatt--teams--deployment_roles))
- `description` (String) Team description
- `is_idp_managed` (Boolean) Whether the team is managed by an identity provider
- `name` (String) Team name
- `organization_role` (String) The role to assign to the organization
- `roles_count` (Number) Number of roles assigned to the team
- `updated_at` (String) Workspace last updated timestamp
- `updated_by` (Attributes) Workspace updater (see [below for nested schema](#nestedatt--teams--updated_by))
- `workspace_roles` (Attributes Set) The roles to assign to the workspaces (see [below for nested schema](#nestedatt--teams--workspace_roles))

<a id="nestedatt--teams--created_by"></a>
### Nested Schema for `teams.created_by`

Read-Only:

- `api_token_name` (String)
- `avatar_url` (String)
- `full_name` (String)
- `id` (String)
- `subject_type` (String)
- `username` (String)


<a id="nestedatt--teams--deployment_roles"></a>
### Nested Schema for `teams.deployment_roles`

Required:

- `deployment_id` (String) The ID of the deployment to assign the role to
- `role` (String) The role to assign to the deployment


<a id="nestedatt--teams--updated_by"></a>
### Nested Schema for `teams.updated_by`

Read-Only:

- `api_token_name` (String)
- `avatar_url` (String)
- `full_name` (String)
- `id` (String)
- `subject_type` (String)
- `username` (String)


<a id="nestedatt--teams--workspace_roles"></a>
### Nested Schema for `teams.workspace_roles`

Required:

- `role` (String) The role to assign to the workspace
- `workspace_id` (String) The ID of the workspace to assign the role to