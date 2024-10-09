---
page_title: "Use Terraform Import Script to migrate existing resources"
---

# Use Terraform Import Script to migrate existing resources
In this guide, we will automate the migration of existing resources into Terraform using the [Terraform Import Script](https://github.com/astronomer/terraform-provider-astro/releases/tag/import/v0.1.3). The Astro Terraform Import Script is a tool designed to help you import existing Astro resources into your Terraform configuration.

### Supported Resources
- Workspace 
- Deployment 
- Cluster
- Hybrid Cluster Workspace Authorization
- API Token 
- Team
- Team Roles
- User Roles

## Step 1: Download the Import Script
1. Download the `terraform-provider-astro-import-script` executable file from [releases](https://github.com/astronomer/terraform-provider-astro/releases) based on your OS and architecture.

## Step 2: Run the Import Script
Run the script with the following command:
    - On Unix-based systems: 
    ```
    chmod +x terraform-provider-astro-import-script_<version-number>_<os>_<arc>
    ./terraform-provider-astro-import-script_<version-number>_<os>_<arc> [options]
    ```
    - On Windows:
    ```
    .\terraform-provider-astro-import-script_<version-number>_<os>_<arc>.exe [options]
    ```
### Options
`-resources`: Comma-separated list of resources to import. Accepted values are workspace, deployment, cluster, api_token, team, team_roles, user_roles.
->Ensure you have the necessary permissions in your Astro organization to access the resources you're attempting to import.
`-token`: API token to authenticate with the Astro platform. If not provided, the script will attempt to use the ASTRO_API_TOKEN environment variable.
`-organizationId`: Organization ID to import resources from.
`-runTerraformInit`: Run terraform init after generating the import configuration. Used for initializing the Terraform state in our GitHub Actions.
`-help`: Display help information.

## Step 3: Review the Output
The script will generate two main files:
`import.tf`: Contains the Terraform import blocks for the specified resources.
`generated.tf`: Contains the Terraform resource configurations for the imported resources.
The generated Terraform configurations may require some manual adjustment to match your specific requirements or to resolve any conflicts.
