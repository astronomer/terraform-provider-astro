# AIRFLOW_VARIABLE — workspace-scoped, non-secret
resource "astro_environment_object" "var_workspace_plain" {
  object_key      = "etl_default_region"
  object_type     = "AIRFLOW_VARIABLE"
  scope           = "WORKSPACE"
  scope_entity_id = "clx42sxw501gl01o0gjenthnh"

  value     = "us-east-1"
  is_secret = false
}

# AIRFLOW_VARIABLE — workspace-scoped, secret. Toggling `is_secret` forces replacement.
resource "astro_environment_object" "var_workspace_secret" {
  object_key      = "external_api_key"
  object_type     = "AIRFLOW_VARIABLE"
  scope           = "WORKSPACE"
  scope_entity_id = "clx42sxw501gl01o0gjenthnh"

  value     = "sk-abc123-replace-me"
  is_secret = true
}

# AIRFLOW_VARIABLE — auto-linked to every Deployment in the workspace
resource "astro_environment_object" "var_workspace_autolinked" {
  object_key            = "feature_flag_v2_pipeline"
  object_type           = "AIRFLOW_VARIABLE"
  scope                 = "WORKSPACE"
  scope_entity_id       = "clx42sxw501gl01o0gjenthnh"
  auto_link_deployments = true

  value     = "enabled"
  is_secret = false
}

# AIRFLOW_VARIABLE — workspace-scoped with a per-Deployment override
resource "astro_environment_object" "var_workspace_with_override" {
  object_key      = "warehouse_database"
  object_type     = "AIRFLOW_VARIABLE"
  scope           = "WORKSPACE"
  scope_entity_id = "clx42sxw501gl01o0gjenthnh"

  value     = "analytics_prod"
  is_secret = false

  links = [
    {
      scope           = "DEPLOYMENT"
      scope_entity_id = "clx44jyu001m201m5dzsbexqr"
      overrides = {
        value = "analytics_staging"
      }
    },
  ]
}

# AIRFLOW_VARIABLE — deployment-scoped (no links/auto_link_deployments allowed)
resource "astro_environment_object" "var_deployment_only" {
  object_key      = "deployment_local_flag"
  object_type     = "AIRFLOW_VARIABLE"
  scope           = "DEPLOYMENT"
  scope_entity_id = "clx44jyu001m201m5dzsbexqr"

  value     = "true"
  is_secret = false
}

# CONNECTION — Postgres with extra JSON (preserved byte-for-byte across refresh)
resource "astro_environment_object" "conn_workspace_postgres" {
  object_key      = "warehouse_postgres"
  object_type     = "CONNECTION"
  scope           = "WORKSPACE"
  scope_entity_id = "clx42sxw501gl01o0gjenthnh"

  type     = "postgres"
  host     = "warehouse.example.com"
  port     = 5432
  login    = "airflow"
  password = "REPLACE_ME"
  schema   = "analytics"
  extra    = jsonencode({ sslmode = "require", timeout = 30 })
}

# CONNECTION — Snowflake with a typed auth provider (auth_type_id)
resource "astro_environment_object" "conn_workspace_snowflake" {
  object_key      = "snowflake_warehouse"
  object_type     = "CONNECTION"
  scope           = "WORKSPACE"
  scope_entity_id = "clx42sxw501gl01o0gjenthnh"

  type         = "snowflake"
  auth_type_id = "snowflake-password"
  host         = "abc12345.us-east-1.snowflakecomputing.com"
  login        = "AIRFLOW_USER"
  password     = "REPLACE_ME"
  schema       = "ANALYTICS"
  extra        = jsonencode({ account = "abc12345", warehouse = "AIRFLOW_WH", role = "AIRFLOW_ROLE" })
}

# CONNECTION — minimal HTTP (only `type` is required)
resource "astro_environment_object" "conn_workspace_http" {
  object_key      = "internal_metrics_api"
  object_type     = "CONNECTION"
  scope           = "WORKSPACE"
  scope_entity_id = "clx42sxw501gl01o0gjenthnh"

  type = "http"
  host = "https://metrics.internal.example.com"
}

# CONNECTION — workspace-scoped with a per-Deployment override and an exclusion
resource "astro_environment_object" "conn_workspace_with_overrides" {
  object_key      = "warehouse_with_per_env_overrides"
  object_type     = "CONNECTION"
  scope           = "WORKSPACE"
  scope_entity_id = "clx42sxw501gl01o0gjenthnh"

  type     = "postgres"
  host     = "warehouse.example.com"
  port     = 5432
  login    = "airflow"
  password = "REPLACE_ME"
  schema   = "analytics"

  links = [
    {
      scope           = "DEPLOYMENT"
      scope_entity_id = "clx44jyu001m201m5dzsbexqr"
      overrides = {
        host   = "warehouse-staging.example.com"
        schema = "analytics_staging"
        extra  = jsonencode({ sslmode = "prefer" })
      }
    },
  ]

  exclude_links = [
    { scope = "DEPLOYMENT", scope_entity_id = "clx44sandbox001m5dzsbexqr" },
  ]
}

# CONNECTION — deployment-scoped
resource "astro_environment_object" "conn_deployment_postgres" {
  object_key      = "dev_postgres"
  object_type     = "CONNECTION"
  scope           = "DEPLOYMENT"
  scope_entity_id = "clx44jyu001m201m5dzsbexqr"

  type     = "postgres"
  host     = "dev-warehouse.example.com"
  port     = 5432
  login    = "dev"
  password = "REPLACE_ME"
  schema   = "dev_analytics"
}

# METRICS_EXPORT — Prometheus with bearer-token auth, custom headers and labels
resource "astro_environment_object" "metrics_workspace_bearer" {
  object_key      = "prometheus_remote_write"
  object_type     = "METRICS_EXPORT"
  scope           = "WORKSPACE"
  scope_entity_id = "clx42sxw501gl01o0gjenthnh"

  endpoint      = "https://prometheus.example.com/api/v1/write"
  exporter_type = "PROMETHEUS"
  auth_type     = "AUTH_TOKEN"
  basic_token   = "REPLACE_ME"
  labels        = { environment = "prod", team = "data" }
  headers       = { "X-Scope-OrgID" = "astro-tenant-1" }
}

# METRICS_EXPORT — Prometheus with basic auth (note: `password` is the HTTP Basic-auth password here)
resource "astro_environment_object" "metrics_workspace_basic_auth" {
  object_key      = "prometheus_remote_write_basic"
  object_type     = "METRICS_EXPORT"
  scope           = "WORKSPACE"
  scope_entity_id = "clx42sxw501gl01o0gjenthnh"

  endpoint      = "https://prometheus.example.com/api/v1/write"
  exporter_type = "PROMETHEUS"
  auth_type     = "BASIC"
  username      = "metrics"
  password      = "REPLACE_ME"
}

# METRICS_EXPORT — workspace-scoped with a per-Deployment override
resource "astro_environment_object" "metrics_workspace_with_override" {
  object_key      = "prometheus_per_env"
  object_type     = "METRICS_EXPORT"
  scope           = "WORKSPACE"
  scope_entity_id = "clx42sxw501gl01o0gjenthnh"

  endpoint      = "https://prometheus.example.com/api/v1/write"
  exporter_type = "PROMETHEUS"
  auth_type     = "AUTH_TOKEN"
  basic_token   = "REPLACE_ME"

  links = [
    {
      scope           = "DEPLOYMENT"
      scope_entity_id = "clx44jyu001m201m5dzsbexqr"
      overrides = {
        endpoint = "https://prometheus-staging.example.com/api/v1/write"
        labels   = { environment = "staging" }
      }
    },
  ]
}

# METRICS_EXPORT — deployment-scoped
resource "astro_environment_object" "metrics_deployment_only" {
  object_key      = "prometheus_dev"
  object_type     = "METRICS_EXPORT"
  scope           = "DEPLOYMENT"
  scope_entity_id = "clx44jyu001m201m5dzsbexqr"

  endpoint      = "https://prometheus-dev.example.com/api/v1/write"
  exporter_type = "PROMETHEUS"
}

# Import an existing environment object
import {
  id = "cm4ntm56001gk01mbhudv1elv"
  to = astro_environment_object.conn_workspace_postgres
}
