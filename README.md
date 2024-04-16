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
1. Create an [API Token](https://docs.astronomer.io/astro/automation-authentication#step-1-create-an-api-token) to use in the provider. We recommend creating an organization API token since it is the most flexible but the type of your API token will depend on your use case.
2. Create a `main.tf` file with the following content:
```terraform
terraform {
  required_providers {
    astronomer = {
      source = "registry.terraform.io/astronomer/astronomer"
    }
  }
}

provider "astronomer" {
  organization_id = "<cuid>"
}

# your terraform commands here
```
3. Run the following commands to apply the provider:
```shell
export ASTRO_API_TOKEN=<token>
terraform init
terraform plan
terraform apply
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, see [Building The Provider](## Building The Provider).

To add example docs, add the correspond `.tf` files to the `examples` directory. These should be added for every new data source and resource.

To run terraform with the provider, create a `.terraformrc` file in your home directory (`~`) with the following content to override the provider installation with the local build:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/astronomer/astronomer" = "~/astronomer-terraform-provider/bin" # Your path to the provider binary
  }
  direct {}
}
```

## Example `main.tf` file for development and testing data sources and resources
```terraform
terraform {
  required_providers {
    astronomer = {
      source = "registry.terraform.io/astronomer/astronomer"
    }
  }
}

provider "astronomer" {
  organization_id = "<cuid>"
  host            = "https://api.astronomer-dev.io"
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
Unit tests can be run with `make test`.

### Acceptance tests
Acceptance integration tests use a Terraform CLI binary to run real Terraform commands against the Astro API. The goal is to approximate using the provider with Terraform in production as closely as possible.

Using the terraform-plugin-testing framework, each `resource.Test` runs an acceptance test on a resource.
- `ProtoV6ProviderFactories`: map of the provider factories that the test suite will use to create the provider - just has the `astronomer` provider
- `PreCheck`: a function that runs before the test suite starts to check that all the required environment variables are set
- `Steps`: a list of `terraform apply` sequences that the test suite will run. Each step is a `resource.TestStep` that contains a `Config` and `Check` function.
  - `Config`: the Terraform configuration that the test will run (ie. the `.tf` file)
  - `Check`: function that will verify the state of the resources after the `terraform apply` command has run.

In order to run the full suite of Acceptance tests, run `make testacc`.
You will also need to set the following environment variables:
- `ASTRO_API_HOST`
- `HOSTED_ORGANIZATION_ID`
- `HOSTED_ORGANIZATION_API_TOKEN` - an organization owner API token for the above organization
- `HYBRID_ORGANIZATION_ID`
- `HYBRID_ORGANIZATION_API_TOKEN` - an organization owner API token for the above organization

The acceptance tests will run against the Astronomer API and create/read/update/delete real resources.



