resource "astro_workspace" "example" {
  name                  = "example"
  description           = "example"
  cicd_enforced_default = "false"
}

resource "astro_cluster" "example" {
  name             = "example"
  type             = "DEDICATED"
  region           = "westus2"
  cloud_provider   = "AZURE"
  vpc_subnet_range = "172.20.0.0/19"
  db_instance_type = "Standard_D2ds_v4"
  workspace_ids    = [astro_workspace.example.id]
}

resource "astro_hybrid_cluster_workspace_authorization" "example" {
  cluster_id    = astro_cluster.example.id
  workspace_ids = [astro_workspace.example.id]
}