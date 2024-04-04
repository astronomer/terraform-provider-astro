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

output "data_workspace_example" {
  value = data.astronomer_workspace.example
}

resource "astronomer_workspace" "tf_workspace" {
  name                  = "tf-workspace-1234"
  description           = "This is a Terraform created workspace"
  cicd_enforced_default = false
}

output "terraform_workspace" {
  value = astronomer_workspace.tf_workspace
}

// terraform import astronomer_workspace.imported_workspace cuid
import {
  to = astronomer_workspace.imported_workspace
  id = "clukhp501000401jdyc42imci"
}
resource "astronomer_workspace" "imported_workspace" {
  name                  = "imported_workspace"
  description           = "hi fred"
  cicd_enforced_default = false
}

output "imported_workspace" {
  value = astronomer_workspace.imported_workspace
}
