---
page_title: "Use Terraform Import Script to migrate existing resources"
---

# Use Import Script to migrate existing resources
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

-> Make sure you run `terraform init` before using the import script, or use the `-runTerraformInit` option when running the import script.

1. If you are using a Unix-based systems, add execute permission to the script file: 
```
chmod +x terraform-provider-astro-import-script_&lt;version-number&gt;_&lt;os&gt;_&lt;arc&gt;
```
2. Run the import script. Insert the script's version, your computer's operating system, and your computer's architecture for `<version-number>`, `<os>` and `<arc>`.

- On Unix-based systems:
```
./terraform-provider-astro-import-script_&lt;version-number&gt;_&lt;os&gt;_&lt;arc&gt; [options]
```
- On Windows:

```
.\terraform-provider-astro-import-script_&lt;version-number&gt;_&lt;os&gt;_&lt;arc&gt;.exe [options]
```

### Options
- `-resources`: Comma-separated list of resources to import. Accepted values are workspace, deployment, cluster, api_token, team, team_roles, user_roles. If not provided, all resources are imported.

-> Ensure you have the necessary permissions in your Astro organization to access the resources you're attempting to import.

- `-token`: API token to authenticate with the Astro platform. If not provided, the script will attempt to use the ASTRO_API_TOKEN environment variable.
- `-organizationId`: Organization ID to import resources from.
- `-runTerraformInit`: Run terraform init after generating the import configuration. Used for initializing the Terraform state in our GitHub Actions.
- `-help`: Display help information.

You following is an example output from running the import script with `-resources deployment` and `-organizationId cf23qgwp001ag01qf0o8er413`:
```
Terraform Import Script Starting
Resources to import:  [deployment]
Using organization ID: cm23qgwp001ap01qm0o3er493
Terraform version 1.9.7 is installed and meets the minimum required version.
Importing deployments for organization cm23qgwp001ap01qm0o3er493
Importing Deployments: [cf23qgwp001ag01qf0o8er413]
Successfully handled resource deployment
Successfully wrote import configuration to import.tf
generated.tf does not exist, no need to delete
terraform.tfstate does not exist, no need to delete
astro_deployment.deployment_cf23qgwp001ag01qf0o8er413: Preparing import... [id=cf23qgwp001ag01qf0o8er413]
astro_deployment.deployment_cf23qgwp001ag01qf0o8er413: Refreshing state... [id=cf23qgwp001ag01qf0o8er413]

Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following symbols:
  + create
  ~ update in-place

Terraform will perform the following actions:

...

Plan: 1 to import, 6 to add, 1 to change, 0 to destroy.

generated.tf does not exist. Creating new file with deployment information.
Generated import for astro_deployment.deployment_cf23qgwp001ag01qf0o8er413
Successfully updated generated.tf with deployment information.
Import process completed successfully. The 'generated.tf' file now includes all resources, including deployments.
Import process completed. Summary:
Resource deployment processed successfully
```

## Step 3: Review output
The script will generate two main files:
- `import.tf`: Contains the Terraform import blocks for the specified resources.
- `generated.tf`: Contains the Terraform resource configurations for the imported resources.
The generated Terraform configurations may require some manual adjustment to match your specific requirements or to resolve any conflicts.

## Step 4: Extract and organize resources
The `generated.tf` file that is created by the import script will contain all of the specified resources in one file. It is recommended that you extract and modularize the resources so they are easily maintained and reusable:
```
terraform-astro-project/
├── environments/
│   ├── dev/
│   │   ├── main.tf              # Root module for development
│   │   ├── variables.tf         # Dev-specific variables
│   │   ├── outputs.tf           # Dev-specific outputs
│   │   └── dev.tfvars           # Variable values for development
│   ├── prod/
│   │   ├── main.tf              # Root module for production
│   │   ├── variables.tf         # Prod-specific variables
│   │   ├── outputs.tf           # Prod-specific outputs
│   │   └── prod.tfvars          # Variable values for production
├── modules/
│   ├── astro/
│   │   ├── main.tf              # Entry point for the module
│   │   ├── workspace.tf         # Defines Astro workspaces
│   │   ├── deployment.tf        # Defines Astro Deployments
│   │   ├── users.tf             # Defines user roles and access
│   │   ├── variables.tf         # Variables used in the astro module
│   │   └── outputs.tf           # Outputs from the astro module
└── cloud_provider.tf     
```
