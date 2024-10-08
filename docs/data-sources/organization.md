---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "astro_organization Data Source - astro"
subcategory: ""
description: |-
  Organization data source
---

# astro_organization (Data Source)

Organization data source

## Example Usage

```terraform
data "astro_organization" "example_organization" {}

# Output the organization value using terraform apply
output "organization" {
  value = data.astro_organization.example_organization
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `billing_email` (String) Organization billing email
- `created_at` (String) Organization creation timestamp
- `created_by` (Attributes) Organization creator (see [below for nested schema](#nestedatt--created_by))
- `id` (String) Organization identifier
- `is_scim_enabled` (Boolean) Whether SCIM is enabled for the organization
- `name` (String) Organization name
- `payment_method` (String) Organization payment method
- `product` (String) Organization product type
- `status` (String) Organization status
- `support_plan` (String) Organization support plan
- `trial_expires_at` (String) Organization trial expiration timestamp
- `updated_at` (String) Organization last updated timestamp
- `updated_by` (Attributes) Organization updater (see [below for nested schema](#nestedatt--updated_by))

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
