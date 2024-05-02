resource "astro_workspace" "dev_workspace" {
  name                  = "dev-workspace"
  description           = "Dev Workspace"
  cicd_enforced_default = true
}

resource "astro_deployment" "team_1_dev_deployment" {
  name                    = "team-1-dev-deployment"
  description             = "Team 1 Dev Deployment"
  type                    = "STANDARD"
  cloud_provider          = "AWS"
  region                  = "us-east-1"
  contact_emails          = []
  default_task_pod_cpu    = "0.25"
  default_task_pod_memory = "0.5Gi"
  executor                = "CELERY"
  is_cicd_enforced        = true
  is_dag_deploy_enabled   = true
  is_development_mode     = true
  is_high_availability    = false
  resource_quota_cpu      = "10"
  resource_quota_memory   = "20Gi"
  scheduler_size          = "SMALL"
  workspace_id            = astro_workspace.dev_workspace.id
  environment_variables   = []
  worker_queues = [{
    name               = "default"
    is_default         = true
    astro_machine      = "A5"
    max_worker_count   = 10
    min_worker_count   = 0
    worker_concurrency = 1
  }]
  scaling_spec = {
    hibernation_spec = {
      override = {
        is_hibernating = true
        override_until = "2025-03-01T13:00:00Z"
      }
    }
  }
}

resource "astro_deployment" "team_2_dev_deployment" {
  name                    = "team-2-dev-deployment"
  description             = "Team 2 Dev Deployment"
  type                    = "STANDARD"
  cloud_provider          = "AWS"
  region                  = "us-east-1"
  contact_emails          = []
  default_task_pod_cpu    = "0.25"
  default_task_pod_memory = "0.5Gi"
  executor                = "CELERY"
  is_cicd_enforced        = true
  is_dag_deploy_enabled   = true
  is_development_mode     = true
  is_high_availability    = false
  resource_quota_cpu      = "10"
  resource_quota_memory   = "20Gi"
  scheduler_size          = "SMALL"
  workspace_id            = astro_workspace.dev_workspace.id
  environment_variables   = []
  worker_queues = [{
    name               = "default"
    is_default         = true
    astro_machine      = "A5"
    max_worker_count   = 10
    min_worker_count   = 0
    worker_concurrency = 1
  }]
  scaling_spec = {
    hibernation_spec = {
      override = {
        is_hibernating = true
        override_until = "2025-03-01T13:00:00Z"
      }
    }
  }
}

resource "astro_deployment" "team_3_dev_deployment" {
  name                    = "team-3-dev-deployment"
  description             = "Team 3 Dev Deployment"
  type                    = "STANDARD"
  cloud_provider          = "AWS"
  region                  = "us-east-1"
  contact_emails          = []
  default_task_pod_cpu    = "0.25"
  default_task_pod_memory = "0.5Gi"
  executor                = "CELERY"
  is_cicd_enforced        = true
  is_dag_deploy_enabled   = true
  is_development_mode     = true
  is_high_availability    = false
  resource_quota_cpu      = "10"
  resource_quota_memory   = "20Gi"
  scheduler_size          = "SMALL"
  workspace_id            = astro_workspace.dev_workspace.id
  environment_variables   = []
  worker_queues = [{
    name               = "default"
    is_default         = true
    astro_machine      = "A5"
    max_worker_count   = 10
    min_worker_count   = 0
    worker_concurrency = 1
  }]
  scaling_spec = {
    hibernation_spec = {
      override = {
        is_hibernating = true
        override_until = "2025-03-01T13:00:00Z"
      }
    }
  }
}

resource "astro_workspace" "staging_workspace" {
  name                  = "staging-workspace"
  description           = "Staging Workspace"
  cicd_enforced_default = true
}

resource "astro_deployment" "team_1_staging_deployment" {
  name                    = "team-1-staging-deployment"
  description             = "Team 1 Staging Deployment"
  type                    = "STANDARD"
  cloud_provider          = "AWS"
  region                  = "us-east-1"
  contact_emails          = []
  default_task_pod_cpu    = "0.25"
  default_task_pod_memory = "0.5Gi"
  executor                = "CELERY"
  is_cicd_enforced        = true
  is_dag_deploy_enabled   = true
  is_development_mode     = true
  is_high_availability    = false
  resource_quota_cpu      = "10"
  resource_quota_memory   = "20Gi"
  scheduler_size          = "SMALL"
  workspace_id            = astro_workspace.staging_workspace.id
  environment_variables   = []
  worker_queues = [{
    name               = "default"
    is_default         = true
    astro_machine      = "A5"
    max_worker_count   = 10
    min_worker_count   = 0
    worker_concurrency = 1
  }]
  scaling_spec = {
    hibernation_spec = {
      override = {
        is_hibernating = true
        override_until = "2025-03-01T13:00:00Z"
      }
    }
  }
}

resource "astro_deployment" "team_2_staging_deployment" {
  name                    = "team-2-staging-deployment"
  description             = "Team 2 Staging Deployment"
  type                    = "STANDARD"
  cloud_provider          = "AWS"
  region                  = "us-east-1"
  contact_emails          = []
  default_task_pod_cpu    = "0.25"
  default_task_pod_memory = "0.5Gi"
  executor                = "CELERY"
  is_cicd_enforced        = true
  is_dag_deploy_enabled   = true
  is_development_mode     = true
  is_high_availability    = false
  resource_quota_cpu      = "10"
  resource_quota_memory   = "20Gi"
  scheduler_size          = "SMALL"
  workspace_id            = astro_workspace.staging_workspace.id
  environment_variables   = []
  worker_queues = [{
    name               = "default"
    is_default         = true
    astro_machine      = "A5"
    max_worker_count   = 10
    min_worker_count   = 0
    worker_concurrency = 1
  }]
  scaling_spec = {
    hibernation_spec = {
      override = {
        is_hibernating = true
        override_until = "2025-03-01T13:00:00Z"
      }
    }
  }
}

resource "astro_deployment" "team_3_staging_deployment" {
  name                    = "team-3-staging-deployment"
  description             = "Team 3 Staging Deployment"
  type                    = "STANDARD"
  cloud_provider          = "AWS"
  region                  = "us-east-1"
  contact_emails          = []
  default_task_pod_cpu    = "0.25"
  default_task_pod_memory = "0.5Gi"
  executor                = "CELERY"
  is_cicd_enforced        = true
  is_dag_deploy_enabled   = true
  is_development_mode     = true
  is_high_availability    = false
  resource_quota_cpu      = "10"
  resource_quota_memory   = "20Gi"
  scheduler_size          = "SMALL"
  workspace_id            = astro_workspace.staging_workspace.id
  environment_variables   = []
  worker_queues = [{
    name               = "default"
    is_default         = true
    astro_machine      = "A5"
    max_worker_count   = 10
    min_worker_count   = 0
    worker_concurrency = 1
  }]
  scaling_spec = {
    hibernation_spec = {
      override = {
        is_hibernating = true
        override_until = "2025-03-01T13:00:00Z"
      }
    }
  }
}

resource "astro_workspace" "prod_workspace" {
  name                  = "prod-workspace"
  description           = "Prod Workspace"
  cicd_enforced_default = true
}

resource "astro_cluster" "prod_aws_cluster" {
  type             = "DEDICATED"
  name             = "prod-aws-cluster"
  region           = "us-east-1"
  cloud_provider   = "AWS"
  db_instance_type = "db.m6g.large"
  vpc_subnet_range = "172.20.0.0/20"
  workspace_ids    = []
  timeouts = {
    create = "3h"
    update = "2h"
    delete = "1m"
  }
}

resource "astro_deployment" "team_1_prod_deployment" {
  name                    = "team-1-prod-deployment"
  description             = "Team 1 Prod Deployment"
  type                    = "DEDICATED"
  cluster_id              = astro_cluster.prod_aws_cluster.id
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
  workspace_id            = astro_workspace.prod_workspace.id
  environment_variables = [{
    key       = "key1"
    value     = "value1"
    is_secret = false
  }]
}

resource "astro_deployment" "team_2_prod_deployment" {
  name                    = "team-2-prod-deployment"
  description             = "Team 2 Prod Deployment"
  type                    = "DEDICATED"
  cluster_id              = astro_cluster.prod_aws_cluster.id
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
  workspace_id            = astro_workspace.prod_workspace.id
  environment_variables = [{
    key       = "key1"
    value     = "value1"
    is_secret = false
  }]
}

resource "astro_deployment" "team_3_prod_deployment" {
  name                    = "team-3-prod-deployment"
  description             = "Team 3 Prod Deployment"
  type                    = "DEDICATED"
  cluster_id              = astro_cluster.prod_aws_cluster.id
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
  workspace_id            = astro_workspace.prod_workspace.id
  environment_variables = [{
    key       = "key1"
    value     = "value1"
    is_secret = false
  }]
}