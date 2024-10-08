---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "astro_users Data Source - astro"
subcategory: ""
description: |-
  Users data source
---

# astro_users (Data Source)

Users data source

## Example Usage

```terraform
data "astro_users" "example_users" {}

data "astro_users" "example_users_filter_by_workspace_id" {
  workspace_id = "clx42sxw501gl01o0gjenthnh"
}

data "astro_users" "example_users_filter_by_deployment_id" {
  deployment_id = "clx44jyu001m201m5dzsbexqr"
}

# Output the users value using terraform apply
output "example_users" {
  value = data.astro_users.example_users
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `deployment_id` (String)
- `workspace_id` (String)

### Read-Only

- `users` (Attributes Set) (see [below for nested schema](#nestedatt--users))

<a id="nestedatt--users"></a>
### Nested Schema for `users`

Required:

- `id` (String) User identifier

Read-Only:

- `avatar_url` (String) User avatar URL
- `created_at` (String) User creation timestamp
- `deployment_roles` (Attributes Set) The roles assigned to the deployments (see [below for nested schema](#nestedatt--users--deployment_roles))
- `full_name` (String) User full name
- `organization_role` (String) The role assigned to the organization
- `status` (String) User status
- `updated_at` (String) User last updated timestamp
- `username` (String) User username
- `workspace_roles` (Attributes Set) The roles assigned to the workspaces (see [below for nested schema](#nestedatt--users--workspace_roles))

<a id="nestedatt--users--deployment_roles"></a>
### Nested Schema for `users.deployment_roles`

Read-Only:

- `deployment_id` (String) The ID of the deployment the role is assigned to
- `role` (String) The role assigned to the deployment


<a id="nestedatt--users--workspace_roles"></a>
### Nested Schema for `users.workspace_roles`

Read-Only:

- `role` (String) The role assigned to the workspace
- `workspace_id` (String) The ID of the workspace the role is assigned to
