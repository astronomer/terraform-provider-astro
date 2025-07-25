---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "astro_deployment Data Source - astro"
subcategory: ""
description: |-
  Deployment data source
---

# astro_deployment (Data Source)

Deployment data source

## Example Usage

```terraform
data "astro_deployment" "example_deployment" {
  id = "clozc036j01to01jrlgvueo8t"
}

# Output the deployment value using terraform apply
output "deployment" {
  value = data.astro_deployment.example_deployment
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) Deployment identifier

### Read-Only

- `airflow_version` (String) Deployment Airflow version
- `astro_runtime_version` (String) Deployment Astro Runtime version
- `cloud_provider` (String) Deployment cloud provider
- `cluster_id` (String) Deployment cluster identifier
- `contact_emails` (Set of String) Deployment contact emails
- `created_at` (String) Deployment creation timestamp
- `created_by` (Attributes) Deployment creator (see [below for nested schema](#nestedatt--created_by))
- `dag_tarball_version` (String) Deployment DAG tarball version
- `default_task_pod_cpu` (String) Deployment default task pod CPU
- `default_task_pod_memory` (String) Deployment default task pod memory
- `description` (String) Deployment description
- `desired_dag_tarball_version` (String) Deployment desired DAG tarball version
- `environment_variables` (Attributes Set) Deployment environment variables (see [below for nested schema](#nestedatt--environment_variables))
- `executor` (String) Deployment executor
- `external_ips` (Set of String) Deployment external IPs
- `image_repository` (String) Deployment image repository
- `image_tag` (String) Deployment image tag
- `image_version` (String) Deployment image version
- `is_cicd_enforced` (Boolean) Whether the Deployment enforces CI/CD deploys
- `is_dag_deploy_enabled` (Boolean) Whether DAG deploy is enabled
- `is_development_mode` (Boolean) Whether Deployment is in development mode
- `is_high_availability` (Boolean) Whether Deployment has high availability
- `name` (String) Deployment name
- `namespace` (String) Deployment namespace
- `oidc_issuer_url` (String) Deployment OIDC issuer URL
- `region` (String) Deployment region
- `remote_execution` (Attributes) Deployment remote execution configuration (see [below for nested schema](#nestedatt--remote_execution))
- `resource_quota_cpu` (String) Deployment resource quota CPU
- `resource_quota_memory` (String) Deployment resource quota memory
- `scaling_spec` (Attributes) Deployment scaling spec (see [below for nested schema](#nestedatt--scaling_spec))
- `scaling_status` (Attributes) Deployment scaling status (see [below for nested schema](#nestedatt--scaling_status))
- `scheduler_au` (Number) Deployment scheduler AU
- `scheduler_cpu` (String) Deployment scheduler CPU
- `scheduler_memory` (String) Deployment scheduler memory
- `scheduler_replicas` (Number) Deployment scheduler replicas
- `scheduler_size` (String) Deployment scheduler size
- `status` (String) Deployment status
- `status_reason` (String) Deployment status reason
- `task_pod_node_pool_id` (String) Deployment task pod node pool identifier
- `type` (String) Deployment type
- `updated_at` (String) Deployment last updated timestamp
- `updated_by` (Attributes) Deployment updater (see [below for nested schema](#nestedatt--updated_by))
- `webserver_airflow_api_url` (String) Deployment webserver Airflow API URL
- `webserver_ingress_hostname` (String) Deployment webserver ingress hostname
- `webserver_url` (String) Deployment webserver URL
- `worker_queues` (Attributes Set) Deployment worker queues (see [below for nested schema](#nestedatt--worker_queues))
- `workload_identity` (String) Deployment workload identity
- `workspace_id` (String) Deployment workspace identifier

<a id="nestedatt--created_by"></a>
### Nested Schema for `created_by`

Read-Only:

- `api_token_name` (String)
- `avatar_url` (String)
- `full_name` (String)
- `id` (String)
- `subject_type` (String)
- `username` (String)


<a id="nestedatt--environment_variables"></a>
### Nested Schema for `environment_variables`

Read-Only:

- `is_secret` (Boolean) Whether Environment variable is a secret
- `key` (String) Environment variable key
- `updated_at` (String) Environment variable last updated timestamp
- `value` (String) Environment variable value


<a id="nestedatt--remote_execution"></a>
### Nested Schema for `remote_execution`

Read-Only:

- `allowed_ip_address_ranges` (Set of String) The allowed IP address ranges for remote execution
- `enabled` (Boolean) Whether remote execution is enabled
- `remote_api_url` (String) The URL for the remote API
- `task_log_bucket` (String) The bucket for task logs
- `task_log_url_pattern` (String) The URL pattern for task logs


<a id="nestedatt--scaling_spec"></a>
### Nested Schema for `scaling_spec`

Read-Only:

- `hibernation_spec` (Attributes) (see [below for nested schema](#nestedatt--scaling_spec--hibernation_spec))

<a id="nestedatt--scaling_spec--hibernation_spec"></a>
### Nested Schema for `scaling_spec.hibernation_spec`

Read-Only:

- `override` (Attributes) (see [below for nested schema](#nestedatt--scaling_spec--hibernation_spec--override))
- `schedules` (Attributes Set) (see [below for nested schema](#nestedatt--scaling_spec--hibernation_spec--schedules))

<a id="nestedatt--scaling_spec--hibernation_spec--override"></a>
### Nested Schema for `scaling_spec.hibernation_spec.override`

Read-Only:

- `is_active` (Boolean) Whether the override is active
- `is_hibernating` (Boolean) Whether the override is hibernating
- `override_until` (String) Time until the override is active


<a id="nestedatt--scaling_spec--hibernation_spec--schedules"></a>
### Nested Schema for `scaling_spec.hibernation_spec.schedules`

Read-Only:

- `description` (String) Description of the schedule
- `hibernate_at_cron` (String) Cron expression for hibernation
- `is_enabled` (Boolean) Whether the schedule is enabled
- `wake_at_cron` (String) Cron expression for waking




<a id="nestedatt--scaling_status"></a>
### Nested Schema for `scaling_status`

Read-Only:

- `hibernation_status` (Attributes) (see [below for nested schema](#nestedatt--scaling_status--hibernation_status))

<a id="nestedatt--scaling_status--hibernation_status"></a>
### Nested Schema for `scaling_status.hibernation_status`

Read-Only:

- `is_hibernating` (Boolean) Whether the deployment is hibernating
- `next_event_at` (String) Time of the next event
- `next_event_type` (String) Type of the next event
- `reason` (String) Reason for the current state



<a id="nestedatt--updated_by"></a>
### Nested Schema for `updated_by`

Read-Only:

- `api_token_name` (String)
- `avatar_url` (String)
- `full_name` (String)
- `id` (String)
- `subject_type` (String)
- `username` (String)


<a id="nestedatt--worker_queues"></a>
### Nested Schema for `worker_queues`

Read-Only:

- `astro_machine` (String) Worker queue Astro machine value
- `id` (String) Worker queue identifier
- `is_default` (Boolean) Whether Worker queue is default
- `max_worker_count` (Number) Worker queue max worker count
- `min_worker_count` (Number) Worker queue min worker count
- `name` (String) Worker queue name
- `node_pool_id` (String) Worker queue node pool identifier
- `pod_cpu` (String) Worker queue pod CPU
- `pod_memory` (String) Worker queue pod memory
- `worker_concurrency` (Number) Worker queue worker concurrency
