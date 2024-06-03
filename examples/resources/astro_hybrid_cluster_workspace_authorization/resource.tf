resource "astro_workspace" "workspace" {
  name                  = "my-workspace"
  description           = "my first workspace"
  cicd_enforced_default = true
}

resource "astro_hybrid_cluster_workspace_authorization" "example" {
  cluster_id    = "clk8h0fv1006801j8yysfybbt"                                  # cluster id
  workspace_ids = ["cl70oe7cu445571iynrkthtybl", astro_workspace.workspace.id] # workspace ids (can pass in existing workspace ids or use the id of a workspace created in the same Terraform configuration)
}