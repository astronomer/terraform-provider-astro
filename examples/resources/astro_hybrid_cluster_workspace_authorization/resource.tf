resource "astro_hybrid_cluster_workspace_authorization" "example" {
  cluster_id    = astro_cluster.example.id
  workspace_ids = [astro_workspace.example.id]
}