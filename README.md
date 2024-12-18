<div align="center">
    <picture>
        <img src="https://github.com/user-attachments/assets/f89d2532-c360-4082-b899-be2593bb2483" width="200">
    </picture>
    <p>Official Astro Terraform Provider to configure and manage your Astro infrastructure through an API.</p>
    <a href="https://registry.terraform.io/providers/astronomer/astro/latest/docs"><img src="https://img.shields.io/static/v1?label=Docs&labelColor=0F0C27&message=terraform-provider-astro&color=4E408D&style=for-the-badge" /></a>
    <a href="https://astronomer.docs.buildwithfern.com/docs/api/overview"><img src="https://img.shields.io/static/v1?label=Docs&labelColor=0F0C27&message=API Ref&color=4E408D&style=for-the-badge" /></a>
</div>

## Using the provider
1. Create an [API Token](https://docs.astronomer.io/astro/automation-authentication#step-1-create-an-api-token) to use in the provider. Astronomer recommends creating an Organization API token since it is the most flexible but the type of your API token will depend on your use case.
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
   See [Astro Provider docs](https://registry.terraform.io/providers/astronomer/astro/latest/docs) for supported resources and data sources.

3. Run the following commands to apply the provider:
```shell
export ASTRO_API_TOKEN=<token>
terraform init # only needed the first time - initializes a working directory and downloads the necessary provider plugins and modules and setting up the backend for storing your infrastructure's state
terraform plan # creates a plan consisting of a set of changes that will make your resources match your configuration
terraform apply # performs a plan just like terraform plan does, but then actually carries out the planned changes to each resource using the relevant infrastructure provider's API
```

## Importing Existing Resources
The Astro Terraform Import Script is a tool designed to help you import existing Astro resources into your Terraform configuration. 
This script automates the process of generating Terraform import blocks and resource configurations for the following resources: workspaces, deployments, clusters, hybrid cluster workspace authorizations, API tokens, teams, team roles, and user roles.
See Astro's [import script guide](https://registry.terraform.io/providers/astronomer/astro/latest/docs/guides/import-script) for more information.

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
- `-token`: API token to authenticate with the Astro platform. This requires the Organization Owner role. If not provided, the script will attempt to use the `ASTRO_API_TOKEN` environment variable.
- `-organizationId`: (Required) Organization ID to import resources from.
- `-runTerraformInit`: Run `terraform init` after generating the import configuration. Used for initializing the Terraform state in our GitHub Actions.
- `-help`: Display help information.

### Examples

1. Import all resources:
   ```
   ./terraform-provider-astro-import-script_<version-number>_<os>_<arc> -organizationId <your_org_id> -token <your_api_token>
   ```

2. Import only workspaces and deployments:
   ```
   ./terraform-provider-astro-import-script_<version-number>_<os>_<arc> -resources workspace,deployment -token <your_api_token> -organizationId <your_org_id>
   ```

3. Import all supported resources and run Terraform init:
   ```
   ./terraform-provider-astro-import-script_<version-number>_<os>_<arc> -resources workspace,deployment,cluster,api_token,team,team_roles,user_roles -token <your_api_token> -organizationId <your_org_id> -runTerraformInit
   ```

4. Use a different API host (for example, dev environment):
   ```
   ./terraform-provider-astro-import-script_<version-number>_<os>_<arc> -resources workspace -token <your_api_token> -organizationId <your_org_id>
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
