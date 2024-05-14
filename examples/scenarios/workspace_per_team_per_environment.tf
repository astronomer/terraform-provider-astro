/* Workspace Per Team and Per Environment

Team 1 workspace non-prod
- Team 1 dev deployment
- Team 1 staging deployment
Team 1 workspace prod
- Team 1 prod deployment

Team 2 workspace non-prod
- Team 2 dev deployment
- Team 2 staging deployment
Team 2 workspace prod
- Team 2 prod deployment

Team 3 workspace non-prod
- Team 3 dev deployment
- Team 3 staging deployment
Team 3 workspace prod
- Team 3 prod deployment
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

resource "astro_workspace" "team_1_workspace_non_prod" {
  name                  = "Team 1 Workspace Non Prod"
  description           = "Team 1 Workspace Non Prod"
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
  workspace_id            = astro_workspace.team_1_workspace_non_prod.id
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

resource "astro_deployment" "team_1_stage_deployment" {
  name                    = "Team 1 Stage Deployment"
  description             = "Team 1 Stage Deployment"
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
  workspace_id            = astro_workspace.team_1_workspace_non_prod.id
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

resource "astro_workspace" "team_1_workspace_prod" {
  name                  = "Team 1 Workspace Prod"
  description           = "Team 1 Workspace Prod"
  cicd_enforced_default = true
}

resource "astro_cluster" "team_1_cluster" {
  type             = "DEDICATED"
  name             = "Team 1 AWS Cluster"
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
  cluster_id              = astro_cluster.team_1_cluster.id
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
  workspace_id            = astro_workspace.team_1_workspace_prod.id
  environment_variables = [{
    key       = "key1"
    value     = "value1"
    is_secret = false
  }]
}

resource "astro_workspace" "team_2_workspace_non_prod" {
  name                  = "Team 2 Workspace Non Prod"
  description           = "Team 2 Workspace Non Prod"
  cicd_enforced_default = true
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
  workspace_id            = astro_workspace.team_2_workspace_non_prod.id
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

resource "astro_deployment" "team_2_stage_deployment" {
  name                    = "Team 2 Stage Deployment"
  description             = "Team 2 Stage Deployment"
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
  workspace_id            = astro_workspace.team_2_workspace_non_prod.id
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

resource "astro_workspace" "team_2_workspace_prod" {
  name                  = "Team 2 Workspace Prod"
  description           = "Team 2 Workspace Prod"
  cicd_enforced_default = true
}

resource "astro_cluster" "team_2_cluster" {
  type             = "DEDICATED"
  name             = "Team 2 AWS Cluster"
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

resource "astro_deployment" "team_2_prod_deployment" {
  name                    = "Team 2 Prod Deployment"
  description             = "Team 2 Prod Deployment"
  type                    = "DEDICATED"
  cluster_id              = astro_cluster.team_2_cluster.id
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
  workspace_id            = astro_workspace.team_2_workspace_prod.id
  environment_variables = [{
    key       = "key1"
    value     = "value1"
    is_secret = false
  }]
}

resource "astro_workspace" "team_3_workspace_non_prod" {
  name                  = "Team 3 Workspace Non Prod"
  description           = "Team 3 Workspace Non Prod"
  cicd_enforced_default = true
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
  workspace_id            = astro_workspace.team_3_workspace_non_prod.id
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

resource "astro_deployment" "team_3_stage_deployment" {
  name                    = "Team 3 Stage Deployment"
  description             = "Team 3 Stage Deployment"
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
  workspace_id            = astro_workspace.team_3_workspace_non_prod.id
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

resource "astro_workspace" "team_3_workspace_prod" {
  name                  = "Team 3 Workspace Prod"
  description           = "Team 3 Workspace Prod"
  cicd_enforced_default = true
}

resource "astro_cluster" "team_3_cluster" {
  type             = "DEDICATED"
  name             = "Team 3 AWS Cluster"
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

resource "astro_deployment" "team_3_prod_deployment" {
  name                    = "Team 3 Prod Deployment"
  description             = "Team 3 Prod Deployment"
  type                    = "DEDICATED"
  cluster_id              = astro_cluster.team_3_cluster.id
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
  workspace_id            = astro_workspace.team_3_workspace_prod.id
  environment_variables = [{
    key       = "key1"
    value     = "value1"
    is_secret = false
  }]
}