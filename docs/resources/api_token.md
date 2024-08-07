---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "astro_api_token Resource - astro"
subcategory: ""
description: |-
  API Token resource
---

# astro_api_token (Resource)

API Token resource

## Example Usage

```terraform
resource "astro_api_token" "organization_token" {
  name        = "organization api token"
  description = "organization api token description"
  type        = "ORGANIZATION"
  roles = [{
    "role" : "ORGANIZATION_OWNER",
    "entity_id" : "clx42kkcm01fo01o06agtmshg",
    "entity_type" : "ORGANIZATION"
  }]
  expiry_period_in_days = 30
}

resource "astro_api_token" "organization_token_with_multiple_roles" {
  name        = "organization api token with multiple roles"
  description = "organization api token description"
  type        = "ORGANIZATION"
  roles = [{
    "role" : "ORGANIZATION_OWNER",
    "entity_id" : "clx42kkcm01fo01o06agtmshg",
    "entity_type" : "ORGANIZATION"
    },
    {
      "role" : "WORKSPACE_OWNER",
      "entity_id" : "clx42sxw501gl01o0gjenthnh",
      "entity_type" : "WORKSPACE"
    },
    {
      "role" : "DEPLOYMENT_ADMIN",
      "entity_id" : "clyn6kxud003x01mtxmccegnh",
      "entity_type" : "DEPLOYMENT"
  }]
}

resource "astro_api_token" "workspace_token" {
  name        = "workspace api token"
  description = "workspace api token description"
  type        = "WORKSPACE"
  roles = [{
    "role" : "WORKSPACE_OWNER",
    "entity_id" : "clx42sxw501gl01o0gjenthnh",
    "entity_type" : "WORKSPACE"
  }]
}

resource "astro_api_token" "workspace_token_with_deployment_role" {
  name        = "workspace api token"
  description = "workspace api token description"
  type        = "WORKSPACE"
  roles = [{
    "role" : "WORKSPACE_OWNER",
    "entity_id" : "clx42sxw501gl01o0gjenthnh",
    "entity_type" : "WORKSPACE"
    },
    {
      "role" : "DEPLOYMENT_ADMIN",
      "entity_id" : "clyn6kxud003x01mtxmccegnh",
      "entity_type" : "DEPLOYMENT"
  }]
}

resource "astro_api_token" "deployment_token" {
  name        = "deployment api token"
  description = "deployment api token description"
  type        = "DEPLOYMENT"
  roles = [{
    "role" : "DEPLOYMENT_ADMIN",
    "entity_id" : "clyn6kxud003x01mtxmccegnh",
    "entity_type" : "DEPLOYMENT"
  }]
}

resource "astro_api_token" "deployment_token_with_custom_role" {
  name        = "deployment api token with custom role"
  description = "deployment api token description"
  type        = "DEPLOYMENT"
  roles = [{
    "role" : "CUSTOM_ROLE",
    "entity_id" : "clyn6kxud003x01mtxmccegnh",
    "entity_type" : "DEPLOYMENT"
  }]
}

# Import an existing api token
import {
  id = "clxm46ged05b301neuucdqwox" // ID of the existing api token
  to = astro_api_token.imported_api_token
}
resource "astro_api_token" "imported_api_token" {
  name        = "imported api token"
  description = "imported api token description"
  type        = "ORGANIZATION"
  roles = [{
    "role" : "ORGANIZATION_OWNER",
    "entity_id" : "clx42kkcm01fo01o06agtmshg",
    "entity_type" : "ORGANIZATION"
  }]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) API Token name
- `roles` (Attributes Set) The roles assigned to the API Token (see [below for nested schema](#nestedatt--roles))
- `type` (String) API Token type - if changing this value, the API Token will be recreated with the new type

### Optional

- `description` (String) API Token description
- `expiry_period_in_days` (Number) API Token expiry period in days

### Read-Only

- `created_at` (String) API Token creation timestamp
- `created_by` (Attributes) API Token creator (see [below for nested schema](#nestedatt--created_by))
- `end_at` (String) time when the API token will expire in UTC
- `id` (String) API Token identifier
- `last_used_at` (String) API Token last used timestamp
- `short_token` (String) API Token short token
- `start_at` (String) time when the API token will become valid in UTC
- `token` (String, Sensitive) API Token value. Warning: This value will be saved in plaintext in the terraform state file.
- `updated_at` (String) API Token last updated timestamp
- `updated_by` (Attributes) API Token updater (see [below for nested schema](#nestedatt--updated_by))

<a id="nestedatt--roles"></a>
### Nested Schema for `roles`

Required:

- `entity_id` (String) The ID of the entity to assign the role to
- `entity_type` (String) The type of entity to assign the role to
- `role` (String) The role to assign to the entity


<a id="nestedatt--created_by"></a>
### Nested Schema for `created_by`

Read-Only:

- `api_token_name` (String)
- `avatar_url` (String)
- `full_name` (String)
- `id` (String)
- `subject_type` (String)
- `username` (String)


<a id="nestedatt--updated_by"></a>
### Nested Schema for `updated_by`

Read-Only:

- `api_token_name` (String)
- `avatar_url` (String)
- `full_name` (String)
- `id` (String)
- `subject_type` (String)
- `username` (String)
