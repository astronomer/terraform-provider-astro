data "astronomer_deployment_options" "example" {}

data "astronomer_deployment_options" "example_with_deployment_id_query_param" {
  deployment_id = "clozc036j01to01jrlgvueo8t"
}

data "astronomer_deployment_options" "example_with_deployment_type_query_param" {
  deployment_type = "DEDICATED"
}

data "astronomer_deployment_options" "example_with_executor_query_param" {
  executor = "CELERY"
}

data "astronomer_deployment_options" "example_with_cloud_provider_query_param" {
  cloud_provider = "AWS"
}
