data "astronomer_cluster_options" "example_cluster_options" {
  type = "HYBRID"
}

data "astronomer_cluster_options" "example_cluster_options_filter_by_provider" {
  type           = "HYBRID"
  cloud_provider = "AWS"
}
