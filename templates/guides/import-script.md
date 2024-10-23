---
page_title: "Use Terraform Import Script to migrate existing resources"
---

# Use Import Script to migrate existing resources
The Astro Terraform Import Script is a tool designed to help you import existing Astro resources into your Terraform configuration.

In this guide, we will migrate an existing Workspace, API token and Team into Terraform using the Terraform Import Script.

## Import Script options
- `-resources`: Comma-separated list of resources to import. Accepted values are 
`workspace`, `deployment`, `cluster`, `api_token`, `team`, `team_roles`,and `user_roles`. If not provided, all resources are imported.

-> Ensure you have the necessary permissions in your Astro organization to access the resources you're attempting to import.

- `-token`: API token to authenticate with the Astro platform. If not provided, the script will attempt to use the `ASTRO_API_TOKEN` environment variable.
- `-organizationId`: Organization ID to import resources from.
- `-runTerraformInit`: Run terraform init after generating the import configuration. Used for initializing the Terraform state in our GitHub Actions.
- `-help`: Display help information.


## Prerequisites
- An [Astro](https://www.astronomer.io/product/) organization with a Workspace, Team, and API token
- An initialized Terraform working directory

## Step 1: Download the Import Script
1. Download the `terraform-provider-astro-import-script` executable file from [releases](https://github.com/astronomer/terraform-provider-astro/releases) based on your OS and architecture. For this guide, the script will be `terraform-provider-astro-import-script_v0.1.3_darwin_arm64`.

## Step 2: Run the Import Script

-> Make sure you run `terraform init` before using the Import Script, or use the `-runTerraformInit` option when running the Import Script.

1. Authenticate with Astro by creating an [API token](https://www.astronomer.io/docs/astro/organization-api-tokens#create-an-organization-api-token) with the **organization owner** role and configure it as an `ASTRO_API_TOKEN` environment variable:
```
export ASTRO_API_TOKEN=&lt;your-api-token&gt;
```

2. If you are using a Unix-based systems, add execute permission to the script file: 
```
chmod +x terraform-provider-astro-import-script_&lt;version-number&gt;_&lt;os&gt;_&lt;arc&gt;
```
3. Run the Import Script. Insert the script's version, your computer's operating system, and your computer's architecture for `<version-number>`, `<os>` and `<arc>`.

- On Unix-based systems:
```
./terraform-provider-astro-import-script_&lt;version-number&gt;_&lt;os&gt;_&lt;arc&gt; [options]
```
- On Windows:

```
.\terraform-provider-astro-import-script_&lt;version-number&gt;_&lt;os&gt;_&lt;arc&gt;.exe [options]
```

To import your existing Workspace, API token and Team, specify those resources with the `-resources` option. The other option you will need to specify is `-organizationId`:
```
./terraform-provider-astro-import-script_v0.1.3_darwin_arm64 -organizationId &lt;your-organization-id&gt; -resources api_token,team,workspace
```

You should see the following output:
```
Terraform Import Script Starting
Resources to import:  [api_token team workspace]
Using organization ID: &lt;your-organization-id&gt
Terraform version 1.9.7 is installed and meets the minimum required version.
Importing teams for organization &lt;your-organization-id&gt
Importing API tokens for organization &lt;your-organization-id&gt
Importing workspaces for organization &lt;your-organization-id&gt
Importing Workspaces: [&lt;workspace-id&gt]
Successfully handled resource workspace
Importing API Tokens: [&lt;api_token-id&gt]
Successfully handled resource api_token
Importing Teams: [&lt;team-id&gt]
Successfully handled resource team
Successfully wrote import configuration to import.tf
Successfully deleted generated.tf
terraform.tfstate does not exist, no need to delete
astro_team.team_&lt;team-id&gt: Preparing import... [id=&lt;team-id&gt]
astro_api_token.api_token_&lt;api_token-id&gt: Preparing import... [id=&lt;api_token-id&gt]
astro_workspace.workspace_&lt;workspace-id&gt: Preparing import... [id=&lt;workspace-id&gt]
astro_workspace.workspace_&lt;workspace-id&gt: Refreshing state... [id=&lt;workspace-id&gt]
astro_api_token.api_token_&lt;api_token-id&gt: Refreshing state... [id=&lt;api_token-id&gt]
astro_team.team_&lt;team-id&gt: Refreshing state... [id=&lt;team-id&gt]

Terraform will perform the following actions:

...

Plan: 3 to import, 0 to add, 0 to change, 0 to destroy.

Terraform has generated configuration and written it to generated.tf. Please review the configuration and edit it as
necessary before adding it to version control.

Note: You didn't use the -out option to save this plan, so Terraform can't guarantee to take exactly these actions if you
run "terraform apply" now.
Import process completed. Summary:
Resource workspace processed successfully
Resource api_token processed successfully
Resource team processed successfully
```
-> If you import Deployments, they will not count torwards the `Plan: 3 to import, 0 to add, 0 to change, 0 to destroy` line of the output even when the Deployments are successfully imported. This is a known issue and is in the process of being fixed.

## Step 3: Review output
The script will generate two main files:
- `import.tf`: Contains the Terraform import blocks for the specified resources.
- `generated.tf`: Contains the Terraform resource configurations for the imported resources.
The generated Terraform configurations might require some manual adjustment to match your specific requirements or to resolve any conflicts.

## Step 4: Extract and organize resources
The `generated.tf` file created by the Import Script contains all of the specified resources in one file. Astronomer recommends that you extract and modularize the resources so they are easily maintained and reusable. The following example shows a well structured Terraform project for managing Astro infrastructure:
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
