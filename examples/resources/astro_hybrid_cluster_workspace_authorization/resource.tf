resource "astro_workspace" "example" {
  name = "example"
}

resource "astro_hybrid_cluster" "example" {
  name         = "example"
  workspace_id = astro_workspace.example.id
}

resource "astro_hybrid_cluster_workspace_authorization" "example" {
  cluster_id    = astro_hybrid_cluster.example.id
  workspace_ids = [astro_workspace.example.id]
}