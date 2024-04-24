data "astro_cluster_options" "example_cluster_options" {
  type = "HYBRID"
}

data "astro_cluster_options" "example_cluster_options_filter_by_provider" {
  type           = "HYBRID"
  cloud_provider = "AWS"
}
