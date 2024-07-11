resource "astro_api_token" "example" {
  name        = "api token"
  description = "api token description"
  type        = "ORGANIZATION"
  roles = [{
    "role" : "ORGANIZATION_OWNER",
    "entity_id" : "clx42kkcm01fo01o06agtmshg",
    "entity_type" : "ORGANIZATION"
  }]
  expiry_period_in_days = 30
}