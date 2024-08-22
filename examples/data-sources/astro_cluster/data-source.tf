data "astro_cluster" "example_cluster" {
  id = "clozc036j01to01jrlgvueo8t"
}

# Output the cluster value using terraform apply
output "cluster" {
  value = data.astro_cluster.example_cluster
}