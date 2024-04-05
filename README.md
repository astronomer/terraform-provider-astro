# Astronomer Terraform Provider

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the following `Makefile` command:

```shell
make dep
make build
```

4. The provider binary will be available in the `bin` directory

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

```shell
export ASTRI_API_TOKEN=<token>
terraform apply
terraform plan
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, see [Building The Provider](## Building The Provider).

To run terraform with the provider, create a `.terraformrc` file in your home directory with the following content to override the provider installation with the local build:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/astronomer/astronomer" = "~/astronomer/astronomer-terraform-provider/bin" # Path to the provider binary
  }
direct {}
}
```

## Example `main.tf` file
```terraform
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
  organization_id = "<cuid>"
  host            = "https://api.astronomer-dev.io"
  token           = var.token
}

data "astronomer_workspace" "example" {
  id = "<cuid>>"
}

output "data_workspace_example" {
  value = data.astronomer_workspace.example
}

resource "astronomer_workspace" "tf_workspace" {
  name                  = "my workspace"
  description           = "my first workspace"
  cicd_enforced_default = false
}

output "terraform_workspace" {
  value = astronomer_workspace.tf_workspace
}
```

## Testing
TODO: In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.



