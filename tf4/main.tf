terraform {
  required_providers {
    astro = {
      source = "astronomer/astro"
    }
  }
}

provider "astro" {
  organization_id = "clsaoc8id051901jsmvivh82z"
}

resource "astro_workspace" "tf_example" {
  name                  = "tfexample"
  description           = "my terraform workspace workspace"
  cicd_enforced_default = true
}

resource "astro_environment_object" "tf_example" {
  object_key      = "tfenvobjectexampleobjkey"
  object_type     = "AIRFLOW_VARIABLE"
  scope           = "WORKSPACE"
  scope_entity_id = astro_workspace.tf_example.id
  airflow_variable = {
    value     = "tfenvobjectexamplevalue"
    is_secret = false
  }
  auto_link_deployments = true
}

resource "astro_environment_object" "tf_example_2" {
  object_key      = "tfenvobjectexampleobjkey2"
  object_type     = "AIRFLOW_VARIABLE"
  scope           = "WORKSPACE"
  scope_entity_id = astro_workspace.tf_example.id
  airflow_variable = {
    value     = "tfenvobjectexamplevalue"
    is_secret = true
  }
  auto_link_deployments = false
}
