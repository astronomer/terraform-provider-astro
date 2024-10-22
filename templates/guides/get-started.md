---
page_title: "Get started with Astro Terraform Provider"
---

# Get started with Astro Terraform Provider
In this guide, we will automate the onboarding of a new team by creating and managing an Astro Workspace and Deployment. By the end of this tutorial, you will have a fully automated setup that is reproducible and easily scalable to more teams.

## Step 1: Create Your Terraform Working Directory
1. Create a folder, `my-data-platform` for your Terraform project.
2. Save the following code in a file named `terraform.tf`:
```
terraform {
  required_providers {
    astro = {
      source  = "astronomer/astro"
      version = "1.0.0"
    }
  }
}

provider "astro" {
  organization_id = &lt;your-organization-id&gt;
}
```
3. Insert your organization's ID for `<your-organization-id>`. The working directory will contain all your Terraform code, and all Terraform commands will be run from this directory.

## Step 2: Initialize the Terraform Working Directory
1. Run `terraform init`. You will see Terraform downloading and installing the Astro Terraform provider to your local computer:
```
$ terraform init

Initializing the backend...

Initializing provider plugins...
- Finding astronomer/astro versions matching "1.0.0"...
- Installing astronomer/astro v1.0.0...
- Installed astronomer/astro v1.0.0 (signed by a HashiCorp partner, key ID F5206453FDEA33CF)

...

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.
```
2. The versions and hashes of providers are stored in a generated file `.terraform.lock.hcl`. Store this file in version control.

## Step 3: Authenticate with Astro
1. [Create an API token](https://www.astronomer.io/docs/astro/automation-authentication#step-1-create-an-api-token) in Astro. Since you are creating a Workspace, you need an [Organization API token](https://www.astronomer.io/docs/astro/organization-api-tokens) with [Organization Owner permissions](https://www.astronomer.io/docs/astro/user-permissions#organization-roles).
2. Configure the API token as an environment variable `ASTRO_API_TOKEN` to run Terraform commands:
`export ASTRO_API_TOKEN=<your-api-token>`

Alternatively, users can set their API token value in the provider block:
```
provider "astro" {
  organization_id = "&lt;your-organization-id&gt;"
  token = "&lt;your-api-token&gt;"
}
```

## Step 4: Define Resources in Terraform
In a file `main.tf`, define two resources, an `astro_workspace` and an `astro_deployment`. These resources will represent an Astro workspace and Astro Deployment, defined in Terraform code:
```
# Create a new workspace
resource "astro_workspace" "my_first_tf_workspace" {
  name                  = "My first TF workspace"
  description           = "My first Terraform-created workspace"
  cicd_enforced_default = false
}

# Create a new Deployment
resource "astro_deployment" "my_first_tf_deployment" {
  name                    = "My first TF deployment"
  description             = "My first Terraform-created deployment"
  type                    = "STANDARD"
  cloud_provider          = "AWS"
  region                  = "us-east-1"
  contact_emails          = []
  default_task_pod_cpu    = "0.25"
  default_task_pod_memory = "0.5Gi"
  executor                = "CELERY"
  is_cicd_enforced        = true
  is_dag_deploy_enabled   = true
  is_development_mode     = false
  is_high_availability    = false
  resource_quota_cpu      = "10"
  resource_quota_memory   = "20Gi"
  scheduler_size          = "SMALL"
  workspace_id            = astro_workspace.my_first_tf_workspace.id
  environment_variables   = []
  worker_queues = [{
    name               = "default"
    is_default         = true
    astro_machine      = "A5"
    max_worker_count   = 10
    min_worker_count   = 0
    worker_concurrency = 1
  }]
}
```
-> One of the key characteristics (and benefits) of using Terraform is that it's *declarative*. For example, `workspace_id = astro_workspace.my_first_tf_workspace.id` tells Terraform to configure the Workspace ID in the Deployment. This means the Workspace must be created first, producing an ID which is a generated value and unknown at the time of writing. You don't have to instruct Terraform to create resources in a certain order, you only have to instruct what to create. The resources above can be defined in any order. Terraform takes the relationships between resources into account when deciding the order of creating resources.

## Step 5: (Optional) Define Outputs
In a file `outputs.tf`, define values you'd like to log after creating the infrastructure. We'll output the Workspace and Deployment IDs:
```
output "terraform_workspace" {
  description = "ID of the TF created workspace"
  value       = astro_workspace.my_first_tf_workspace.id
}

output "terraform_deployment" {
  description = "ID of the TF created deployment"
  value       = astro_deployment.my_first_tf_deployment.id
}
```

## Step 6: Preview the Terraform changes
You should now have 3 files: `terraform.tf`, `main.tf` and `outputs.tf` (optional).
1. Run [`terraform plan`](https://developer.hashicorp.com/terraform/cli/commands/plan) to let Terraform create an execution plan and preview the infrastructure changes that Terraform will make. You should see the following text:
```
$ terraform plan

Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # astro_deployment.my_first_tf_deployment will be created
  + resource "astro_deployment" "my_first_tf_deployment" {
      ...
    }

  # astro_workspace.my_first_tf_workspace will be created
  + resource "astro_workspace" "my_first_tf_workspace" {
      ...
    }

Plan: 2 to add, 0 to change, 0 to destroy.

Changes to Outputs:
  + terraform_deployment = (known after apply)
  + terraform_workspace  = (known after apply)
  ```
2. Verify the generated plan contains the text `Plan: 2 to add, 0 to change, 0 to destroy.` This validates that the plan is to create two resources, which are the Workspace and Deployment as defined in `main.tf`.

## Step 7: Apply the Terraform Plan
Run `terraform apply` and select `yes` to execute the plan. This creates the Astro resources and will print their ids, as you defined in `outputs.tf`: 
```
$ terraform apply

Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

...

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes

astro_workspace.my_first_tf_workspace: Creating...
astro_workspace.my_first_tf_workspace: Creation complete after 0s [id=&lt;workspace-id&gt]
astro_deployment.my_first_tf_deployment: Creating...
astro_deployment.my_first_tf_deployment: Creation complete after 1s [id=&lt;deployment-id&gt]

Apply complete! Resources: 2 added, 0 changed, 0 destroyed.

Outputs:

terraform_deployment = "&lt;deployment-id&gt"
terraform_workspace = "&lt;workspace-id&gt"
```
The resources were created and will now be visible in Astro.

## Step 8: Clean Up Terraform-Created Resources
Run `terraform destroy` and select `yes`:
```
$ terraform destroy
astro_workspace.my_first_tf_workspace: Refreshing state...
  [id=&lt;workspace-id&gt]
astro_deployment.my_first_tf_deployment: Refreshing state...
  [id=&lt;deployment-id&gt]

Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following symbols:
  - destroy

Terraform will perform the following actions:

  ...

Plan: 0 to add, 0 to change, 2 to destroy.

Changes to Outputs:
  - terraform_deployment = "&lt;deployment-id&gt" -> null
  - terraform_workspace  = "&lt;workspace-id&gt" -> null

Do you really want to destroy all resources?
  Terraform will destroy all your managed infrastructure, as shown above.
  There is no undo. Only 'yes' will be accepted to confirm.

  Enter a value: yes

astro_deployment.my_first_tf_deployment: Destroying...
  [id=&lt;deployment-id&gt]
astro_deployment.my_first_tf_deployment: Destruction complete after 1s
astro_workspace.my_first_tf_workspace: Destroying...
  [id=&lt;workspace-id&gt]
astro_workspace.my_first_tf_workspace: Destruction complete after 0s

Destroy complete! Resources: 2 destroyed.
```
The output shows two destroyed resources which are the Workspace and Deployment that you first created.