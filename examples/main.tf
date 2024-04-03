terraform {
  required_providers {
    astronomer = {
      source = "registry.terraform.io/astronomer/astronomer"
    }
  }
}

provider "astronomer" {
  organization_id = "cljzz64cc001n01mln1pgkvpj"
  host            = "https://api.astronomer-dev.io"
  token           = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhcGlUb2tlbklkIjoiY2x0aHg3ZWN0MDAxbDAxaTJibnM4NTVzMiIsImF1ZCI6ImFzdHJvbm9tZXItZWUiLCJpYXQiOjE3MDk4NTc4MzEsImlzQXN0cm9ub21lckdlbmVyYXRlZCI6dHJ1ZSwiaXNzIjoiaHR0cHM6Ly9hcGkuYXN0cm9ub21lci1kZXYuaW8iLCJraWQiOiJjbG80aGN0OWcwMDA1MDFxdmNpdGc0aGpoIiwicGVybWlzc2lvbnMiOlsiYXBpVG9rZW5JZDpjbHRoeDdlY3QwMDFsMDFpMmJuczg1NXMyIiwib3JnYW5pemF0aW9uSWQ6Y2xqeno2NGNjMDAxbjAxbWxuMXBna3ZwaiIsIm9yZ1Nob3J0TmFtZTpjbGp6ejY0Y2MwMDFuMDFtbG4xcGdrdnBqIl0sInNjb3BlIjoiYXBpVG9rZW5JZDpjbHRoeDdlY3QwMDFsMDFpMmJuczg1NXMyIG9yZ2FuaXphdGlvbklkOmNsanp6NjRjYzAwMW4wMW1sbjFwZ2t2cGogb3JnU2hvcnROYW1lOmNsanp6NjRjYzAwMW4wMW1sbjFwZ2t2cGoiLCJzdWIiOiJjbDdxcWU0dGYyNjQ0NDJkMjhmdHRvZTdnOCIsInZlcnNpb24iOiJjbHRoeDdlY3QwMDFrMDFpMjhpYXF1Y2s1In0.Mu1Q65LJQtpWHZTheipiQoJm3Yw3kaL8RDqg_lU4zXQoYhoD7CHC3ADSKgIrCqOkZ5CUh9kd68rhXfHLpwgbtq84Gfd1ejBlVHkOVWqQ6RLBxZXpOp24yrKusImnDSlU0fVTpzzsEug9cCdmQZ_1P4mxW4nYt3MpjrS_oPasbpNl_YRIo8pQvTaKK0uAagPX4rjvvkx4-aRYVh2gDr-cVcBjvhcfJawl7AflFkNIbFVkpd8vIbGeKgXDb2UN-AwAg3NWHUp6zenOykwxmcZ0dq-lRtoZnC88fpOpNH11Bu4y3lXQUPtaWRJaE7XNixjdnOcLehYb2ydGZGs0UQQV-Q"
}

data "astronomer_workspace" "example" {
  id = "cltj71ygr000101qafj5hhihs"
}

output "data_workspace_updated_by" {
  value = data.astronomer_workspace.example.created_by
}

resource "astronomer_workspace" "tf_workspace" {
  name                  = "tf-workspace"
  description           = "This is a Terraform created workspace"
  cicd_enforced_default = true
}

// terraform import astronomer_workspace.imported_workspace cuid
import {
  to = astronomer_workspace.imported_workspace
  id = "cltj6pn3v000001owkjx4xhuv"
}
resource "astronomer_workspace" "imported_workspace" {
  name                  = "imported_workspace_2"
  description           = "hi fred"
  cicd_enforced_default = false
}

output "imported_workspace_updated_by" {
  value = astronomer_workspace.imported_workspace.updated_by
}

output "imported_workspace_created_by" {
  value = astronomer_workspace.imported_workspace.created_by
}