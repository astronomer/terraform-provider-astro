resource "astro_custom_role" "example" {
  name        = "Deployment_Viewer"
  description = "Custom role for viewing deployments"
  scope_type  = "DEPLOYMENT"
  permissions = [
    "deployment.get",
    "deployment.delete"
  ]
}

# Custom role with restricted workspace access
resource "astro_custom_role" "restricted_example" {
  name        = "Limited_Deployment_Viewer"
  description = "Custom role restricted to specific workspaces"
  scope_type  = "DEPLOYMENT"
  permissions = [
    "deployment.get"
  ]
  restricted_workspace_ids = [
    "clxxxxxxxxxxxx",
    "clyyyyyyyyyyyyy"
  ]
}

