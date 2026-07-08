# The astro_alerts resource manages many alerts as a single resource. It batches the underlying
# API calls and automatically chunks requests that exceed the per-request limits (30 for
# create/update, 20 for delete), so you can declare any number of alerts in one block.
#
# Alerts are keyed by a stable, user-defined string. Changing a key is treated as deleting the old
# alert and creating a new one.

resource "astro_alerts" "team_alerts" {
  alerts = {
    "etl_dag_failure" = {
      name                     = "ETL DAG Failure"
      type                     = "DAG_FAILURE"
      severity                 = "CRITICAL"
      entity_type              = "DEPLOYMENT"
      entity_id                = "cm1zkps2a0cv301ph39benet6"
      notification_channel_ids = ["cm4nwrvyg024h01mk2dn58m5s"]
      rules = {
        properties = {
          deployment_id = "cm1zkps2a0cv301ph39benet6"
        }
        pattern_matches = [
          {
            entity_type   = "DAG_ID"
            operator_type = "IS"
            values        = ["etl_dag"]
          }
        ]
      }
    }

    "reporting_dag_duration" = {
      name                     = "Reporting DAG Duration"
      type                     = "DAG_DURATION"
      severity                 = "WARNING"
      entity_type              = "DEPLOYMENT"
      entity_id                = "cm1zkps2a0cv301ph39benet6"
      notification_channel_ids = ["cm4nwrvyg024h01mk2dn58m5s"]
      rules = {
        properties = {
          deployment_id        = "cm1zkps2a0cv301ph39benet6"
          dag_duration_seconds = 3600
        }
        pattern_matches = [
          {
            entity_type   = "DAG_ID"
            operator_type = "IS"
            values        = ["reporting_dag"]
          }
        ]
      }
    }
  }
}

# Generate many alerts programmatically — the resource chunks them across requests for you.
resource "astro_alerts" "per_dag_failure_alerts" {
  alerts = {
    for dag in toset(["a", "b", "c"]) :
    "${dag}_failure" => {
      name                     = "${dag} DAG Failure"
      type                     = "DAG_FAILURE"
      severity                 = "CRITICAL"
      entity_type              = "DEPLOYMENT"
      entity_id                = "cm1zkps2a0cv301ph39benet6"
      notification_channel_ids = ["cm4nwrvyg024h01mk2dn58m5s"]
      rules = {
        properties = {
          deployment_id = "cm1zkps2a0cv301ph39benet6"
        }
        pattern_matches = [
          {
            entity_type   = "DAG_ID"
            operator_type = "IS"
            values        = [dag]
          }
        ]
      }
    }
  }
}
