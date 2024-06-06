resource "astro_hybrid_cluster_workspace_authorization" "example" {
  cluster_id    = "clk8h0fv1006801j8yysfybbt"
  workspace_ids = ["cl70oe7cu445571iynrkthtybl", "cl70oe7cu445571iynrkthacsd"]
}

// Import existing hybrid cluster workspace authorization
import {
  id = "clk8h0fv1006801j8yysfybbt" // ID of the existing hybrid cluster
  to = astro_hybrid_cluster_workspace_authorization.imported_cluster_workspace_authorization
}
resource "astro_hybrid_cluster_workspace_authorization" "imported_cluster_workspace_authorization" {
  cluster_id    = "clk8h0fv1006801j8yysfybbt"
  workspace_ids = ["cl70oe7cu445571iynrkthtybl"]
}