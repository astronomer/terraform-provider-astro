terraform {
  required_providers {
    astronomer = {
      source = "registry.terraform.io/astronomer/astronomer"
    }
  }
}

variable "token" {
  type = string
}

provider "astronomer" {
  organization_id = "cljzz64cc001n01mln1pgkvpj"
  host            = "https://api.astronomer-dev.io"
  token           = var.token
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