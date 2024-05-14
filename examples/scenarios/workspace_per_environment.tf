/* Workspace Per Environment

Dev workspace
- Team 1 Dev Deployment
- Team 2 Dev Deployment
- Team 3 Dev Deployment

Staging Workspace
- Team 1 Staging Deployment
- Team 2 Staging Deployment
- Team 3 Staging Deployment

Prod Workspace
- Team 1 Prod Deployment
- Team 2 Prod Deployment
- Team 3 Prod Deployment
*/

terraform {
  required_providers {
    astro = {
      source = "registry.terraform.io/astronomer/astro"
    }
  }
}

provider "astro" {
  organization_id = "XXXXX"
}

resource "astro_workspace" "dev_workspace" {
  name                  = "Dev Workspace"
  description           = "Dev Workspace"
  cicd_enforced_default = true
}

resource "astro_deployment" "team_1_dev_deployment" {
  name                    = "Team 1 Dev Deployment"
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
      schedules = [{
        is_enabled        = true
        hibernate_at_cron = "20 * * * *"
        wake_at_cron      = "10 * * * *"
      }]
    }
  }
}

resource "astro_deployment" "team_2_dev_deployment" {
  name                    = "Team 2 Dev Deployment"
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
      schedules = [{
        is_enabled        = true
        hibernate_at_cron = "20 * * * *"
        wake_at_cron      = "10 * * * *"
      }]
    }
  }
}

resource "astro_deployment" "team_3_dev_deployment" {
  name                    = "Team 3 Dev Deployment"
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
      schedules = [{
        is_enabled        = true
        hibernate_at_cron = "20 * * * *"
        wake_at_cron      = "10 * * * *"
      }]
    }
  }
}

resource "astro_workspace" "staging_workspace" {
  name                  = "Staging Workspace"
  description           = "Staging Workspace"
  cicd_enforced_default = true
}

resource "astro_deployment" "team_1_staging_deployment" {
  name                    = "Team 1 Staging Deployment"
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
      schedules = [{
        is_enabled        = true
        hibernate_at_cron = "20 * * * *"
        wake_at_cron      = "10 * * * *"
      }]
    }
  }
}

resource "astro_deployment" "team_2_staging_deployment" {
  name                    = "Team 2 Staging Deployment"
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
      schedules = [{
        is_enabled        = true
        hibernate_at_cron = "20 * * * *"
        wake_at_cron      = "10 * * * *"
      }]
    }
  }
}

resource "astro_deployment" "team_3_staging_deployment" {
  name                    = "Team 3 Staging Deployment"
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
      schedules = [{
        is_enabled        = true
        hibernate_at_cron = "20 * * * *"
        wake_at_cron      = "10 * * * *"
      }]
    }
  }
}

resource "astro_workspace" "prod_workspace" {
  name                  = "Prod Workspace"
  description           = "Prod Workspace"
  cicd_enforced_default = true
}

resource "astro_cluster" "prod_aws_cluster" {
  type             = "DEDICATED"
  name             = "Prod AWS Cluster"
  region           = "us-east-1"
  cloud_provider   = "AWS"
  db_instance_type = "db.m6g.large"
  vpc_subnet_range = "172.20.0.0/20"
  workspace_ids    = []
  timeouts = {
    create = "3h"
    update = "2h"
    delete = "1h"
  }
}

resource "astro_deployment" "team_1_prod_deployment" {
  name                    = "Team 1 Prod Deployment"
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
  name                    = "Team 2 Prod Deployment"
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
  name                    = "Team 3 Prod Deployment"
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