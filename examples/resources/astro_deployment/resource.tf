resource "astro_deployment" "dedicated" {
  name                    = "my dedicated deployment"
  description             = "an example deployment"
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
  environment_variables = [{
    key       = "key1"
    value     = "value1"
    is_secret = false
  }]
}

resource "astro_deployment" "standard" {
  name                    = "my standard deployment"
  description             = "an example deployment"
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
  workspace_id            = "clnp86ly500a401ndaga20g81"
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

resource "astro_deployment" "hybrid" {
  name                  = "my hybrid deployment"
  description           = "an example deployment"
  type                  = "HYBRID"
  cluster_id            = "clnp86ly5000401ndagu20g81"
  task_pod_node_pool_id = "clnp86ly5000301ndzfxz895w"
  contact_emails        = ["example@astronomer.io"]
  executor              = "KUBERNETES"
  is_cicd_enforced      = true
  is_dag_deploy_enabled = true
  scheduler_replicas    = 1
  scheduler_au          = 5
  workspace_id          = "clnp86ly5000401ndaga20g81"
  environment_variables = [{
    key       = "key1"
    value     = "value1"
    is_secret = false
  }]
}