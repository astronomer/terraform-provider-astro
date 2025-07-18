---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "astro_deployment Resource - astro"
subcategory: ""
description: |-
  Deployment resource
---

# astro_deployment (Resource)

Deployment resource

## Example Usage

```terraform
resource "astro_deployment" "dedicated" {
  original_astro_runtime_version = "11.3.0"
  name                           = "my dedicated deployment"
  description                    = "an example deployment"
  type                           = "DEDICATED"
  cluster_id                     = "clv17vgft000801kkydsws63x"
  contact_emails                 = ["preview@astronomer.test"]
  default_task_pod_cpu           = "0.25"
  default_task_pod_memory        = "0.5Gi"
  executor                       = "KUBERNETES"
  is_cicd_enforced               = true
  is_dag_deploy_enabled          = true
  is_development_mode            = false
  is_high_availability           = true
  resource_quota_cpu             = "10"
  resource_quota_memory          = "20Gi"
  scheduler_size                 = "SMALL"
  workspace_id                   = "clnp86ly5000401ndaga21g81"
  desired_workload_identity      = "arn:aws:iam::123456789:role/AirflowS3Logs-clmk2qqia000008mhff3ndjr0"
  environment_variables = [{
    key       = "key1"
    value     = "value1"
    is_secret = false
  }]
}

resource "astro_deployment" "dedicated_astro_remote" {
  original_astro_runtime_version = "3.0-1"
  name                           = "my dedicated deployment"
  description                    = "an example deployment"
  type                           = "DEDICATED"
  cluster_id                     = "clv17vgft000801kkydsws63x"
  contact_emails                 = ["preview@astronomer.test"]
  executor                       = "ASTRO"
  is_cicd_enforced               = true
  is_dag_deploy_enabled          = false
  is_development_mode            = false
  is_high_availability           = true
  scheduler_size                 = "SMALL"
  workspace_id                   = "clnp86ly5000401ndaga21g81"
  desired_workload_identity      = "arn:aws:iam::123456789:role/AirflowS3Logs-clmk2qqia000008mhff3ndjr0"
  environment_variables = [{
    key       = "key1"
    value     = "value1"
    is_secret = false
  }]
  remote_execution = {
    enabled                   = true
    allowed_ip_address_ranges = ["8.8.8.8/32"]
    task_log_bucket           = "s3://my-task-log-bucket"
  }
}

resource "astro_deployment" "standard" {
  original_astro_runtime_version = "11.3.0"
  name                           = "my standard deployment"
  description                    = "an example deployment"
  type                           = "STANDARD"
  cloud_provider                 = "AWS"
  region                         = "us-east-1"
  contact_emails                 = []
  default_task_pod_cpu           = "0.25"
  default_task_pod_memory        = "0.5Gi"
  executor                       = "CELERY"
  is_cicd_enforced               = true
  is_dag_deploy_enabled          = true
  is_development_mode            = false
  is_high_availability           = false
  resource_quota_cpu             = "10"
  resource_quota_memory          = "20Gi"
  scheduler_size                 = "SMALL"
  workspace_id                   = "clnp86ly500a401ndaga20g81"
  environment_variables          = []
  worker_queues = [{
    name               = "default"
    is_default         = true
    astro_machine      = "A5"
    max_worker_count   = 10
    min_worker_count   = 0
    worker_concurrency = 1
  }]
}

resource "astro_deployment" "standard_astro" {
  original_astro_runtime_version = "3-0.1"
  name                           = "my standard deployment"
  description                    = "an example deployment"
  type                           = "STANDARD"
  cloud_provider                 = "AWS"
  region                         = "us-east-1"
  contact_emails                 = []
  default_task_pod_cpu           = "0.25"
  default_task_pod_memory        = "0.5Gi"
  executor                       = "ASTRO"
  is_cicd_enforced               = true
  is_dag_deploy_enabled          = true
  is_development_mode            = false
  is_high_availability           = false
  resource_quota_cpu             = "10"
  resource_quota_memory          = "20Gi"
  scheduler_size                 = "SMALL"
  workspace_id                   = "clnp86ly500a401ndaga20g81"
  environment_variables          = []
  worker_queues = [{
    name               = "default"
    is_default         = true
    astro_machine      = "A5"
    max_worker_count   = 10
    min_worker_count   = 0
    worker_concurrency = 1
  }]
}

resource "astro_deployment" "hybrid" {
  original_astro_runtime_version = "11.3.0"
  name                           = "my hybrid deployment"
  description                    = "an example deployment"
  type                           = "HYBRID"
  cluster_id                     = "clnp86ly5000401ndagu20g81"
  task_pod_node_pool_id          = "clnp86ly5000301ndzfxz895w"
  contact_emails                 = ["example@astronomer.io"]
  executor                       = "KUBERNETES"
  is_cicd_enforced               = true
  is_dag_deploy_enabled          = true
  scheduler_replicas             = 1
  scheduler_au                   = 5
  workspace_id                   = "clnp86ly5000401ndaga20g81"
  environment_variables = [{
    key       = "key1"
    value     = "value1"
    is_secret = false
  }]
}

resource "astro_deployment" "hybrid_celery" {
  original_astro_runtime_version = "11.3.0"
  name                           = "my hybrid celery deployment"
  description                    = "an example deployment with celery executor"
  type                           = "HYBRID"
  cluster_id                     = "clnp86ly5000401ndagu20g81"
  contact_emails                 = ["example@astronomer.io"]
  executor                       = "CELERY"
  is_cicd_enforced               = true
  is_dag_deploy_enabled          = true
  scheduler_replicas             = 1
  scheduler_au                   = 5
  workspace_id                   = "clnp86ly5000401ndaga20g81"
  environment_variables = [{
    key       = "key1"
    value     = "value1"
    is_secret = false
  }]
  worker_queues = [{
    name               = "default"
    is_default         = true
    node_pool_id       = "clnp86ly5000301ndzfxz895w"
    max_worker_count   = 10
    min_worker_count   = 0
    worker_concurrency = 1
  }]
}

// Import an existing deployment
import {
  id = "clv17vgft000801kkydsws63x" // ID of the existing deployment
  to = astro_deployment.imported_deployment
}
resource "astro_deployment" "imported_deployment" {
  name                    = "import me"
  description             = "an existing deployment"
  type                    = "DEDICATED"
  cluster_id              = "clv17vgft000801kkydsws63x"
  contact_emails          = ["preview@astronomer.test"]
  default_task_pod_cpu    = "0.25"
  default_task_pod_memory = "0.5Gi"
  executor                = "KUBERNETES"
  is_cicd_enforced        = true
  is_dag_deploy_enabled   = true
  is_development_mode     = false
  is_high_availability    = true
  resource_quota_cpu      = "10"
  resource_quota_memory   = "20Gi"
  scheduler_size          = "SMALL"
  workspace_id            = "clnp86ly5000401ndaga21g81"
  environment_variables   = []
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `contact_emails` (Set of String) Deployment contact emails
- `description` (String) Deployment description
- `environment_variables` (Attributes Set) Deployment environment variables (see [below for nested schema](#nestedatt--environment_variables))
- `executor` (String) Deployment executor. Valid values: CELERY, KUBERNETES, ASTRO.
- `is_cicd_enforced` (Boolean) Deployment CI/CD enforced
- `is_dag_deploy_enabled` (Boolean) Whether DAG deploy is enabled - Changing this value may disrupt your deployment. Read more at https://docs.astronomer.io/astro/deploy-dags#enable-or-disable-dag-only-deploys-on-a-deployment
- `name` (String) Deployment name
- `type` (String) Deployment type - if changing this value, the deployment will be recreated with the new type
- `workspace_id` (String) Deployment workspace identifier - if changing this value, the deployment will be recreated in the new workspace

### Optional

- `cloud_provider` (String) Deployment cloud provider - required for 'STANDARD' deployments. If changing this value, the deployment will be recreated in the new cloud provider
- `cluster_id` (String) Deployment cluster identifier - required for 'HYBRID' and 'DEDICATED' deployments. If changing this value, the deployment will be recreated in the new cluster
- `default_task_pod_cpu` (String) Deployment default task pod CPU - required for 'STANDARD' and 'DEDICATED' deployments
- `default_task_pod_memory` (String) Deployment default task pod memory - required for 'STANDARD' and 'DEDICATED' deployments
- `desired_workload_identity` (String) Deployment's desired workload identity. The Terraform provider will use this provided workload identity to create the Deployment. If it is not provided the workload identity will be assigned automatically.
- `is_development_mode` (Boolean) Deployment development mode - required for 'STANDARD' and 'DEDICATED' deployments. If changing from 'False' to 'True', the deployment will be recreated
- `is_high_availability` (Boolean) Deployment high availability - required for 'STANDARD' and 'DEDICATED' deployments
- `original_astro_runtime_version` (String) Deployment's original Astro Runtime version. The Terraform provider will use this provided Astro runtime version to create the Deployment. The Astro runtime version can be updated with your Astro project Dockerfile, but if this value is changed, the Deployment will be recreated with this new Astro runtime version.
- `region` (String) Deployment region - required for 'STANDARD' deployments. If changing this value, the deployment will be recreated in the new region
- `remote_execution` (Attributes) Deployment remote execution configuration - only for 'DEDICATED' deployments (see [below for nested schema](#nestedatt--remote_execution))
- `resource_quota_cpu` (String) Deployment resource quota CPU - required for 'STANDARD' and 'DEDICATED' deployments
- `resource_quota_memory` (String) Deployment resource quota memory - required for 'STANDARD' and 'DEDICATED' deployments
- `scaling_spec` (Attributes) Deployment scaling spec - only for 'STANDARD' and 'DEDICATED' deployments (see [below for nested schema](#nestedatt--scaling_spec))
- `scheduler_au` (Number) Deployment scheduler AU - required for 'HYBRID' deployments
- `scheduler_replicas` (Number) Deployment scheduler replicas - required for 'HYBRID' deployments
- `scheduler_size` (String) Deployment scheduler size - required for 'STANDARD' and 'DEDICATED' deployments
- `task_pod_node_pool_id` (String) Deployment task pod node pool identifier - required if executor is 'KUBERNETES' and type is 'HYBRID'
- `worker_queues` (Attributes Set) Deployment worker queues - required for deployments with 'CELERY' executor. For 'STANDARD' and 'DEDICATED' deployments, use astro_machine. For 'HYBRID' deployments, use node_pool_id. (see [below for nested schema](#nestedatt--worker_queues))

### Read-Only

- `airflow_version` (String) Deployment Airflow version
- `astro_runtime_version` (String) Deployment's current Astro Runtime version
- `created_at` (String) Deployment creation timestamp
- `created_by` (Attributes) Deployment creator (see [below for nested schema](#nestedatt--created_by))
- `dag_tarball_version` (String) Deployment DAG tarball version
- `desired_dag_tarball_version` (String) Deployment desired DAG tarball version
- `external_ips` (Set of String) Deployment external IPs
- `id` (String) Deployment identifier
- `image_repository` (String) Deployment image repository
- `image_tag` (String) Deployment image tag
- `image_version` (String) Deployment image version
- `namespace` (String) Deployment namespace
- `oidc_issuer_url` (String) Deployment OIDC issuer URL
- `scaling_status` (Attributes) Deployment scaling status (see [below for nested schema](#nestedatt--scaling_status))
- `scheduler_cpu` (String) Deployment scheduler CPU
- `scheduler_memory` (String) Deployment scheduler memory
- `status` (String) Deployment status
- `status_reason` (String) Deployment status reason
- `updated_at` (String) Deployment last updated timestamp
- `updated_by` (Attributes) Deployment updater (see [below for nested schema](#nestedatt--updated_by))
- `webserver_airflow_api_url` (String) Deployment webserver Airflow API URL
- `webserver_ingress_hostname` (String) Deployment webserver ingress hostname
- `webserver_url` (String) Deployment webserver URL
- `workload_identity` (String) Deployment workload identity. This value can be changed via the Astro API if applicable.

<a id="nestedatt--environment_variables"></a>
### Nested Schema for `environment_variables`

Required:

- `is_secret` (Boolean) Whether Environment variable is a secret
- `key` (String) Environment variable key

Optional:

- `value` (String, Sensitive) Environment variable value

Read-Only:

- `updated_at` (String) Environment variable last updated timestamp


<a id="nestedatt--remote_execution"></a>
### Nested Schema for `remote_execution`

Required:

- `enabled` (Boolean) Whether remote execution is enabled

Optional:

- `allowed_ip_address_ranges` (Set of String) The allowed IP address ranges for remote execution
- `task_log_bucket` (String) The bucket for task logs
- `task_log_url_pattern` (String) The URL pattern for task logs

Read-Only:

- `remote_api_url` (String) The URL for the remote API


<a id="nestedatt--scaling_spec"></a>
### Nested Schema for `scaling_spec`

Required:

- `hibernation_spec` (Attributes) Hibernation configuration for the deployment. The deployment will hibernate according to the schedules defined in this configuration. To remove the hibernation configuration, set scaling_spec to null. (see [below for nested schema](#nestedatt--scaling_spec--hibernation_spec))

<a id="nestedatt--scaling_spec--hibernation_spec"></a>
### Nested Schema for `scaling_spec.hibernation_spec`

Optional:

- `override` (Attributes) Hibernation override configuration. Set to null to remove the override. (see [below for nested schema](#nestedatt--scaling_spec--hibernation_spec--override))
- `schedules` (Attributes Set) List of hibernation schedules. Set to null to remove all schedules. (see [below for nested schema](#nestedatt--scaling_spec--hibernation_spec--schedules))

<a id="nestedatt--scaling_spec--hibernation_spec--override"></a>
### Nested Schema for `scaling_spec.hibernation_spec.override`

Required:

- `is_hibernating` (Boolean)

Optional:

- `override_until` (String)

Read-Only:

- `is_active` (Boolean)


<a id="nestedatt--scaling_spec--hibernation_spec--schedules"></a>
### Nested Schema for `scaling_spec.hibernation_spec.schedules`

Required:

- `hibernate_at_cron` (String)
- `is_enabled` (Boolean)
- `wake_at_cron` (String)

Optional:

- `description` (String)




<a id="nestedatt--worker_queues"></a>
### Nested Schema for `worker_queues`

Required:

- `is_default` (Boolean) Worker queue default
- `max_worker_count` (Number) Worker queue max worker count
- `min_worker_count` (Number) Worker queue min worker count
- `name` (String) Worker queue name
- `worker_concurrency` (Number) Worker queue worker concurrency

Optional:

- `astro_machine` (String) Worker queue Astro machine value - required for 'STANDARD' and 'DEDICATED' deployments
- `node_pool_id` (String) Worker queue Node pool identifier - required for 'HYBRID' deployments

Read-Only:

- `pod_cpu` (String) Worker queue pod CPU
- `pod_memory` (String) Worker queue pod memory


<a id="nestedatt--created_by"></a>
### Nested Schema for `created_by`

Read-Only:

- `api_token_name` (String)
- `avatar_url` (String)
- `full_name` (String)
- `id` (String)
- `subject_type` (String)
- `username` (String)


<a id="nestedatt--scaling_status"></a>
### Nested Schema for `scaling_status`

Read-Only:

- `hibernation_status` (Attributes) (see [below for nested schema](#nestedatt--scaling_status--hibernation_status))

<a id="nestedatt--scaling_status--hibernation_status"></a>
### Nested Schema for `scaling_status.hibernation_status`

Read-Only:

- `is_hibernating` (Boolean)
- `next_event_at` (String)
- `next_event_type` (String)
- `reason` (String)



<a id="nestedatt--updated_by"></a>
### Nested Schema for `updated_by`

Read-Only:

- `api_token_name` (String)
- `avatar_url` (String)
- `full_name` (String)
- `id` (String)
- `subject_type` (String)
- `username` (String)
