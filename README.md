# Terraform Provider Astro

<div align="center">
    <picture>
        <img src="https://github.com/user-attachments/assets/22586f12-3871-4bb6-8ead-40bec82ec3ce" width="200">
    </picture>
    <p>Official Astro Terraform Provider to automate, scale, and manage your Astro infrastructure through an API.</p>
    <a href="https://registry.terraform.io/providers/astronomer/astro/latest/docs"><img src="https://img.shields.io/static/v1?label=Docs&labelColor=0F0C27&message=terraform-provider-astro&color=4E408D&style=for-the-badge" /></a>
    <a href="https://astronomer.docs.buildwithfern.com/docs/api/overview"><img src="https://img.shields.io/static/v1?label=Docs&labelColor=0F0C27&message=API Ref&color=4E408D&style=for-the-badge" /></a>
</div>


## Using the provider
1. Create an [API Token](https://docs.astronomer.io/astro/automation-authentication#step-1-create-an-api-token) to use in the provider. Astronomer recommends creating an organization API token since it is the most flexible but the type of your API token will depend on your use case.
2. Create a `main.tf` file with the following content:
```terraform
terraform {
  required_providers {
    astro = {
      source = "astronomer/astro"
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

## Example `main.tf` file for testing data sources and resources
```terraform
terraform {
  required_providers {
    astro = {
      source = "astronomer/astro"
    }
  }
}

# provider configuration
provider "astro" {
  organization_id = "<cuid>"
}

# get information on an existing workspace
data "astro_workspace" "example" {
  id = "<cuid>"
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
    vpc_subnet_range = "172.20.0.0/20"
    workspace_ids = [astro_workspace.tf_workspace.id, data.astro_workspace.example.id]
    timeouts = {
        create = "3h"
        update = "2h"
        delete = "20m"
    }
}

# create a new dedicated deployment resource in that cluster
resource "astro_deployment" "tf_dedicated_deployment" {
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
resource "astro_deployment" "tf_standard_deployment" {
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

## Importing Existing Resources
The Astro Terraform Import Script is a tool designed to help you import existing Astro resources into your Terraform configuration. 
Currently, this script automates the process of generating Terraform import blocks and resource configurations for the following resources: workspaces, deployments, clusters, hybrid cluster workspace authorizations, API tokens, teams, team roles, and user roles.

To use the import script, download the `terraform-provider-astro-import-script` executable file from [releases](https://github.com/astronomer/terraform-provider-astro/releases) based on your OS and architecture and run it with the following command:

On Unix-based systems:

```
chmod +x terraform-provider-astro-import-script_<version-number>_<os>_<arc>

./terraform-provider-astro-import-script_<version-number>_<os>_<arc> [options]
```

On Windows:

```
.\terraform-provider-astro-import-script_<version-number>_<os>_<arc>.exe [options]
```

### Options

- `-resources`: Comma-separated list of resources to import. Accepted values are workspace, deployment, cluster, api_token, team, team_roles, user_roles.
- `-token`: API token to authenticate with the Astro platform. If not provided, the script will attempt to use the `ASTRO_API_TOKEN` environment variable.
- `-organizationId`: Organization ID to import resources from.
- `-runTerraformInit`: Run `terraform init` after generating the import configuration. Used for initializing the Terraform state in our GitHub Actions.
- `-help`: Display help information.

### Examples

1. Import workspaces and deployments:
   ```
   ./terraform-provider-astro-import-script_<version-number>_<os>_<arc> -resources workspace,deployment -token your_api_token -organizationId your_org_id
   ```

2. Import all supported resources and run Terraform init:
   ```
   ./terraform-provider-astro-import-script_<version-number>_<os>_<arc> -resources workspace,deployment,cluster,api_token,team,team_roles,user_roles -token your_api_token -organizationId your_org_id -runTerraformInit
   ```

3. Use a different API host (e.g., dev environment):
   ```
   ./terraform-provider-astro-import-script_<version-number>_<os>_<arc> -resources workspace -token your_api_token -organizationId your_org_id
   ```

### Output

The script will generate two main files:

1. `import.tf`: Contains the Terraform import blocks for the specified resources.
2. `generated.tf`: Contains the Terraform resource configurations for the imported resources.

### Notes

- Ensure you have the necessary permissions in your Astro organization to access the resources you're attempting to import.
- The generated Terraform configurations may require some manual adjustment to match your specific requirements or to resolve any conflicts.
- Always review the generated files before applying them to your Terraform state.

## FAQ and Troubleshooting

### Frequently Asked Questions

1. **What resources can I manage with this Terraform provider?** 
   - Workspaces, deployments, clusters, hybrid cluster workspace authorizations, API tokens, teams, team roles, and user roles.

2. **How do I authenticate with the Astro API?**
   - Use an API token set as the `ASTRO_API_TOKEN` environment variable or add it to the provider configuration.

3. **Can I import existing Astro resources into Terraform?**
   - Yes, use the Astro Terraform Import Script to generate import blocks and resource configurations.

4. **What Terraform versions are required?**
   - Terraform >= 1.7.


### Troubleshooting

1. **Issue: 401 Unauthorized error when running `terraform plan` or `terraform apply`**

   Solution: Your API token may have expired. Update your `ASTRO_API_TOKEN` environment variable with a fresh token:
   ```
   export ASTRO_API_TOKEN=<your-new-token>
   ```
   
2. **Issue: Import script fails to find resources**

   Solution:
    - Ensure you have the correct permissions in your Astro organization.
    - Verify that your API token is valid and has the necessary scopes and permissions.
    - Double-check the organization ID provided to the script.

3. **Issue: "Error: Invalid provider configuration" when initializing Terraform**

   Solution: Ensure your `.terraformrc` file is correctly set up, especially if you're using a local build of the provider for development.

If you encounter any issues not listed here, please check the [GitHub Issues](https://github.com/astronomer/terraform-provider-astro/issues) page or open a new issue with details about your problem.
