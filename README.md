# Terraform Provider Astro

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
    astro = {
      source = "registry.terraform.io/astronomer/astro"
    }
  }
}

provider "astro" {
  organization_id = "<cuid>"
}

# your terraform commands here
```
3. Run the following commands to apply the provider:
```shell
export ASTRO_API_TOKEN=<token>
terraform init # only needed the first time - initializes a working directory and downloads the necessary provider plugins and modules and setting up the backend for storing your infrastructure's state
terraform plan # creates a plan consisting of a set of changes that will make your resources match your configuration
terraform apply # performs a plan just like terraform plan does, but then actually carries out the planned changes to each resource using the relevant infrastructure provider's API
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, see [Building The Provider](## Building The Provider).

To add example docs, add the correspond `.tf` files to the `examples` directory. These should be added for every new data source and resource.

To run terraform with the provider, create a `.terraformrc` file in your home directory (`~`) with the following content to override the provider installation with the local build:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/astronomer/astro" = "~/terraform-provider-astro/bin" # Your path to the provider binary
  }
  direct {}
}
```

## Example `main.tf` file for development and testing data sources and resources
```terraform
terraform {
  required_providers {
    astro = {
      source = "registry.terraform.io/astronomer/astro"
    }
  }
}

# provider configuration
provider "astro" {
  organization_id = "<cuid>"
  host            = "https://api.astronomer-dev.io"
}

# get information on an existing workspace
data "astro_workspace" "example" {
  id = "<cuid>>"
}

# output the workspace data to the terminal
output "data_workspace_example" {
  value = data.astro_workspace.example
}

# create a new workspace
resource "astro_workspace" "tf_workspace" {
  name                  = "my workspace"
  description           = "my first workspace"
  cicd_enforced_default = false
}

# output the newly created workspace resource to the terminal
output "terraform_workspace" {
  value = astro_workspace.tf_workspace
}

# create a new cluster resource
resource "astro_cluster" "tf_cluster" {
    type = "DEDICATED"
    name = "my first cluster"
    region = "us-east-1"
    cloud_provider = "AWS"
    db_instance_type = "db.m6g.large"
    vpc_subnet_range = "172.20.0.0/20"
    workspace_ids = [astro_workspace.tf_workspace.id, data.astro_workspace.example.id]
    timeouts = {
        create = "3h"
        update = "2h"
        delete = "20m"
    }
}

# create a new dedicated deployment resource in that cluster
resource "astro_deployment" "tf_deployment" {
  name        = "my first dedicated deployment"
  description = ""
  cluster_id  = astro_cluster.tf_cluster.id
  type = "DEDICATED"
  contact_emails = ["example@astronomer.io"]
  default_task_pod_cpu = "0.25"
  default_task_pod_memory = "0.5Gi"
  executor = "KUBERNETES"
  is_cicd_enforced = true
  is_dag_deploy_enabled = true
  is_development_mode = false
  is_high_availability = true
  resource_quota_cpu = "10"
  resource_quota_memory = "20Gi"
  scheduler_size = "SMALL"
  workspace_id = astro_workspace.tf_workspace.id
  environment_variables = [{
      key = "key1"
      value = "value1"
      is_secret = false
  }]
}

# create a new standard deployment resource
resource "astro_standard_deployment" "tf_standard_deployment" {
  name        = "my first standard deployment"
  description = ""
  type = "STANDARD"
  cloud_provider = "AWS"
  region = "us-east-1"
  contact_emails = []
  default_task_pod_cpu = "0.25"
  default_task_pod_memory = "0.5Gi"
  executor = "CELERY"
  is_cicd_enforced = true
  is_dag_deploy_enabled = true
  is_development_mode = false
  is_high_availability = false
  resource_quota_cpu = "10"
  resource_quota_memory = "20Gi"
  scheduler_size = "SMALL"
  workspace_id = astro_workspace.tf_workspace.id
  environment_variables = []
  worker_queues = [{
      name = "default"
      is_default = true
      astro_machine = "A5"
      max_worker_count = 10
      min_worker_count = 0
      worker_concurrency = 1
  }]
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



