data "astro_clusters" "example_clusters" {}

data "astro_clusters" "example_clusters_filter_by_names" {
  names = ["my cluster"]
}

data "astro_clusters" "example_clusters_filter_by_cloud_provider" {
  cloud_provider = "AWS"
}

# Output the clusters value using terraform apply
output "clusters" {
  value = data.astro_clusters.example_clusters
}