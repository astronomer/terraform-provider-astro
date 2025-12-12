/* Workspace Per Team

Team 1 Workspace
- Dev Deployment
- Stage Deployment
- Prod Deployment

Team 2
- Dev Deployment
- Stage Deployment
- Prod Deployment

Team 3
- Dev Deployment
- Stage Deployment
- Prod Deployment
*/

terraform {
  required_providers {
    astro = {
      source = "astronomer/astro"
    }
  }
}


provider "astro" {
  organization_id = "XXXXX"
}

resource "astro_workspace" "team_1_workspace" {
  name                  = "Team 1 Workspace"
  description           = "Team 1 Workspace"
  cicd_enforced_default = true
}

resource "astro_cluster" "team_1_cluster" {
  type             = "DEDICATED"
  name             = "Team 1 AWS Cluster"
  region           = "us-east-1"
  cloud_provider   = "AWS"
  vpc_subnet_range = "172.20.0.0/20"
  workspace_ids    = []
  timeouts = {
    create = "3h"
    update = "2h"
    delete = "1h"
  }
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
  workspace_id            = astro_workspace.team_1_workspace.id
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
  workspace_id            = astro_workspace.team_1_workspace.id
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
  workspace_id            = astro_workspace.team_1_workspace.id
  environment_variables = [{
    key       = "key1"
    value     = "value1"
    is_secret = false
  }]
}

resource "astro_team" "team_1_admins" {
  name              = "Team 1"
  description       = ""
  member_ids        = ["cl26baazt276912f06nnne1234"]
  organization_role = "ORGANIZATION_MEMBER"
  # Available Organization Roles:
  # - ORGANIZATION_OWNER
  # - ORGANIZATION_BILLING_ADMIN
  # - ORGANIZATION_OBSERVE_ADMIN
  # - ORGANIZATION_OBSERVE_MEMBER
  # - ORGANIZATION_MEMBER
  # https://www.astronomer.io/docs/astro/user-permissions#organization-roles
  workspace_roles   = [{
    workspace_id    = astro_workspace.team_1_workspace.id
    # Available Workspace Roles:
    # - WORKSPACE_OWNER
    # - WORKSPACE_OPERATOR
    # - WORKSPACE_AUTHOR
    # - WORKSPACE_MEMBER
    # - WORKSPACE_ACCESSOR
    # https://www.astronomer.io/docs/astro/user-permissions#workspace-roles
    role            = "WORKSPACE_OWNER"
  }]
}



resource "astro_team" "team_1_users" {
  name              = "Team 1"
  description       = ""
  member_ids        = ["cl26baazt276912f06nnne5678", "cl26baazt276912f06nnne9999"]
  organization_role = "ORGANIZATION_MEMBER"
  workspace_roles   = [{
    workspace_id    = astro_workspace.team_1_workspace.id
    role            = "WORKSPACE_MEMBER"
  }]
  deployment_roles  = [{
    deployment_id   = astro_deployment.team_1_dev_deployment.id
    # Available Deployment Roles:
    # - DEPLOYMENT_ADMIN (https://www.astronomer.io/docs/astro/user-permissions#deployment-roles)
    # - Custom roles (https://www.astronomer.io/docs/astro/deployment-role-reference)
    role            = "DEPLOYMENT_ADMIN"
  }]
}

resource "astro_workspace" "team_2_workspace" {
  name                  = "Team 2 Workspace"
  description           = "Team 2 Workspace"
  cicd_enforced_default = true
}

resource "astro_cluster" "team_2_cluster" {
  type             = "DEDICATED"
  name             = "Team 2 AWS Cluster"
  region           = "us-east-1"
  cloud_provider   = "AWS"
  vpc_subnet_range = "172.20.0.0/20"
  workspace_ids    = []
  timeouts = {
    create = "3h"
    update = "2h"
    delete = "1h"
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
  workspace_id            = astro_workspace.team_2_workspace.id
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
  workspace_id            = astro_workspace.team_2_workspace.id
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
  workspace_id            = astro_workspace.team_2_workspace.id
  environment_variables = [{
    key       = "key1"
    value     = "value1"
    is_secret = false
  }]
}

resource "astro_workspace" "team_3_workspace" {
  name                  = "Team 3 Workspace"
  description           = "Team 3 Workspace"
  cicd_enforced_default = true
}

resource "astro_cluster" "team_3_cluster" {
  type             = "DEDICATED"
  name             = "Team 3 AWS Cluster"
  region           = "us-east-1"
  cloud_provider   = "AWS"
  vpc_subnet_range = "172.20.0.0/20"
  workspace_ids    = []
  timeouts = {
    create = "3h"
    update = "2h"
    delete = "1h"
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
  workspace_id            = astro_workspace.team_3_workspace.id
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
  workspace_id            = astro_workspace.team_3_workspace.id
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
  workspace_id            = astro_workspace.team_3_workspace.id
  environment_variables = [{
    key       = "key1"
    value     = "value1"
    is_secret = false
  }]
}