data "astro_deployment_options" "example" {}

data "astro_deployment_options" "example_with_deployment_id_query_param" {
  deployment_id = "clozc036j01to01jrlgvueo8t"
}

data "astro_deployment_options" "example_with_deployment_type_query_param" {
  deployment_type = "DEDICATED"
}

data "astro_deployment_options" "example_with_executor_query_param" {
  executor = "CELERY"
}

data "astro_deployment_options" "example_with_cloud_provider_query_param" {
  cloud_provider = "AWS"
}
