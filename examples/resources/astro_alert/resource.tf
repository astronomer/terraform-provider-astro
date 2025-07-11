resource "astro_alert" "dag_failure_alert" {
  name                     = "DAG Failure Alert"
  type                     = "DAG_FAILURE"
  severity                 = "CRITICAL"
  entity_type              = "DEPLOYMENT"
  entity_id                = "cm1zkps2a0cv301ph39benet6"
  notification_channel_ids = ["cm4nwrvyg024h01mk2dn58m5s"]
  rules = {
    "properties" = {
      "deployment_id" = "cm1zkps2a0cv301ph39benet6"
    }
    "pattern_matches" = [
      {
        "entity_type"   = "DAG_ID"
        "operator_type" = "IS"
        "values"        = ["*", "test"]
      }
    ]
  }
}

resource "astro_alert" "dag_success_alert" {
  name                     = "DAG Success Alert"
  type                     = "DAG_SUCCESS"
  severity                 = "CRITICAL"
  entity_type              = "DEPLOYMENT"
  entity_id                = "cm1zkps2a0cv301ph39benet6"
  notification_channel_ids = ["cm4nwrvyg024h01mk2dn58m5s"]
  rules = {
    "properties" = {
      "deployment_id" = "cm1zkps2a0cv301ph39benet6"
    }
    "pattern_matches" = [
      {
        "entity_type"   = "DAG_ID"
        "operator_type" = "INCLUDES"
        "values"        = ["test"]
      }
    ]
  }
}

resource "astro_alert" "dag_duration_alert" {
  name                     = "DAG Duration Alert"
  type                     = "DAG_DURATION"
  severity                 = "CRITICAL"
  entity_type              = "DEPLOYMENT"
  entity_id                = "cm1zkps2a0cv301ph39benet6"
  notification_channel_ids = ["cm4nwrvyg024h01mk2dn58m5s"]
  rules = {
    "properties" = {
      "deployment_id"        = "cm1zkps2a0cv301ph39benet6"
      "dag_duration_seconds" = 3600
    }
    "pattern_matches" = [
      {
        "entity_type"   = "DAG_ID"
        "operator_type" = "IS"
        "values"        = ["*"]
      },
      {
        "entity_type"   = "DAG_ID"
        "operator_type" = "EXCLUDES"
        "values"        = ["bad_dag"]
      }
    ]
  }
}

resource "astro_alert" "dag_timeliness_alert" {
  name                     = "DAG Timeliness Alert"
  type                     = "DAG_TIMELINESS"
  severity                 = "CRITICAL"
  entity_type              = "DEPLOYMENT"
  entity_id                = "cm1zkps2a0cv301ph39benet6"
  notification_channel_ids = ["cm4nwrvyg024h01mk2dn58m5s"]
  rules = {
    "properties" = {
      "deployment_id"            = "cm1zkps2a0cv301ph39benet6"
      "dag_deadline"             = "08:00",
      "days_of_week"             = ["MONDAY"],
      "look_back_period_seconds" = 3600,
    }
    "pattern_matches" = [
      {
        "entity_type"   = "DAG_ID"
        "operator_type" = "IS"
        "values"        = ["etl_dag"]
      },
    ]
  }
}

resource "astro_alert" "task_failure_alert" {
  name                     = "Task Failure Alert"
  type                     = "TASK_FAILURE"
  severity                 = "CRITICAL"
  entity_type              = "DEPLOYMENT"
  entity_id                = "cm1zkps2a0cv301ph39benet6"
  notification_channel_ids = ["cm4nwrvyg024h01mk2dn58m5s"]
  rules = {
    "properties" = {
      "deployment_id" = "cm1zkps2a0cv301ph39benet6"
    }
    "pattern_matches" = [
      {
        "entity_type"   = "DAG_ID"
        "operator_type" = "IS"
        "values"        = ["*"]
      },
      {
        "entity_type"   = "TASK_ID"
        "operator_type" = "INCLUDES"
        "values"        = ["test_task"]
      }
    ]
  }
}

resource "astro_alert" "task_duration_alert" {
  name                     = "Task Duration Alert"
  type                     = "TASK_DURATION"
  severity                 = "CRITICAL"
  entity_type              = "DEPLOYMENT"
  entity_id                = "cm1zkps2a0cv301ph39benet6"
  notification_channel_ids = ["cm4nwrvyg024h01mk2dn58m5s"]
  rules = {
    "properties" = {
      "deployment_id"         = "cm1zkps2a0cv301ph39benet6"
      "task_duration_seconds" = 3600
    }
    "pattern_matches" = [
      {
        "entity_type"   = "DAG_ID"
        "operator_type" = "INCLUDES"
        "values"        = ["etl_dag"]
      },
      {
        "entity_type"   = "TASK_ID"
        "operator_type" = "EXCLUDES"
        "values"        = ["bad_task"]
      }
    ]
  }
}

# Import an existing alert
import {
  id = "cm1zkps2a0cv301ph39benet6" // ID of the existing alert
  to = astro_alert.dag_failure_alert
}

resource "astro_alert" "dag_failure_alert_imported" {
  name                     = "Imported DAG Failure Alert"
  type                     = "DAG_FAILURE"
  severity                 = "CRITICAL"
  entity_type              = "DEPLOYMENT"
  entity_id                = "cm1zkps2a0cv301ph39benet6"
  notification_channel_ids = ["cm4nwrvyg024h01mk2dn58m5s"]
  rules = {
    "properties" = {
      "deployment_id" = "cm1zkps2a0cv301ph39benet6"
    }
    "pattern_matches" = [
      {
        "entity_type"   = "DAG_ID"
        "operator_type" = "IS"
        "values"        = ["*", "test"]
      }
    ]
  }
}