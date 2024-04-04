resource "workspace_resource" "example" {
  name                  = "my-workspace"
  description           = "my first workspace"
  cicd_enforced_default = true
}
