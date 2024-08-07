---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "astro_user_roles Resource - astro"
subcategory: ""
description: |-
  User Roles resource
---

# astro_user_roles (Resource)

User Roles resource

## Example Usage

```terraform
resource "astro_user_roles" "organization_role_only" {
  user_id           = "clzaftcaz006001lhkey6qzzg"
  organization_role = "ORGANIZATION_OWNER"
}

resource "astro_user_roles" "workspace_roles" {
  user_id           = "clzaftcaz006001lhkey6qzzg"
  organization_role = "ORGANIZATION_MEMBER"
  workspace_roles = [
    {
      workspace_id = "clx42sxw501gl01o0gjenthnh"
      role         = "WORKSPACE_MEMBER"
    }
  ]
}

resource "astro_user_roles" "deployment_roles" {
  user_id           = "clzaftcaz006001lhkey6qzzg"
  organization_role = "ORGANIZATION_MEMBER"
  deployment_roles = [
    {
      deployment_id = "clyn6kxud003x01mtxmccegnh"
      role          = "DEPLOYMENT_ADMIN"
    }
  ]
}

resource "astro_user_roles" "all_roles" {
  user_id           = "clzaftcaz006001lhkey6qzzg"
  organization_role = "ORGANIZATION_MEMBER"
  workspace_roles = [
    {
      workspace_id = "clx42sxw501gl01o0gjenthnh"
      role         = "WORKSPACE_OWNER"
    },
    {
      workspace_id = "clzafte7z006001lhkey6qzzb"
      role         = "WORKSPACE_MEMBER"
    }
  ]
  deployment_roles = [
    {
      deployment_id = "clyn6kxud003x01mtxmccegnh"
      role          = "my custom role"
    }
  ]
}

# Import an existing user roles
import {
  id = "clzaftcaz006001lhkey6qzzg" # ID of the existing user
  to = astro_user_roles.imported_user_roles
}
resource "astro_user_roles" "imported_user_roles" {
  user_id           = "clzaftcaz006001lhkey6qzzg"
  organization_role = "ORGANIZATION_MEMBER"
  workspace_roles = [
    {
      workspace_id = "clx42sxw501gl01o0gjenthnh"
      role         = "WORKSPACE_OWNER"
    },
    {
      workspace_id = "clzafte7z006001lhkey6qzzb"
      role         = "WORKSPACE_MEMBER"
    }
  ]
  deployment_roles = [
    {
      deployment_id = "clyn6kxud003x01mtxmccegnh"
      role          = "my custom role"
    }
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `organization_role` (String) The role to assign to the organization
- `user_id` (String) The ID of the user to assign the roles to

### Optional

- `deployment_roles` (Attributes Set) The roles to assign to the deployments (see [below for nested schema](#nestedatt--deployment_roles))
- `workspace_roles` (Attributes Set) The roles to assign to the workspaces (see [below for nested schema](#nestedatt--workspace_roles))

<a id="nestedatt--deployment_roles"></a>
### Nested Schema for `deployment_roles`

Required:

- `deployment_id` (String) The ID of the deployment to assign the role to
- `role` (String) The role to assign to the deployment


<a id="nestedatt--workspace_roles"></a>
### Nested Schema for `workspace_roles`

Required:

- `role` (String) The role to assign to the workspace
- `workspace_id` (String) The ID of the workspace to assign the role to
