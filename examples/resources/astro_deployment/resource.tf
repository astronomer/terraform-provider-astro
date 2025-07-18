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