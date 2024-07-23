resource "astro_api_token" "example_organization_token" {
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

resource "astro_api_token" "example_organization_token_with_multiple_roles" {
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

resource "astro_api_token" "example_workspace_token" {
  name        = "workspace api token"
  description = "workspace api token description"
  type        = "WORKSPACE"
  roles = [{
    "role" : "WORKSPACE_OWNER",
    "entity_id" : "clx42sxw501gl01o0gjenthnh",
    "entity_type" : "WORKSPACE"
  }]
}

resource "astro_api_token" "example_workspace_token_with_deployment_role" {
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

resource "astro_api_token" "example_deployment_token" {
  name        = "deployment api token"
  description = "deployment api token description"
  type        = "DEPLOYMENT"
  roles = [{
    "role" : "DEPLOYMENT_ADMIN",
    "entity_id" : "clyn6kxud003x01mtxmccegnh",
    "entity_type" : "DEPLOYMENT"
  }]
}

resource "astro_api_token" "example_deployment_token_with_custom_role" {
  name        = "deployment api token with custom role"
  description = "deployment api token description"
  type        = "DEPLOYMENT"
  roles = [{
    "role" : "CUSTOM_ROLE",
    "entity_id" : "clyn6kxud003x01mtxmccegnh",
    "entity_type" : "DEPLOYMENT"
  }]
}
