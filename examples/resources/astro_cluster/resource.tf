resource "astro_cluster" "aws_example" {
  type               = "DEDICATED"
  name               = "my first aws cluster"
  region             = "us-east-1"
  cloud_provider     = "AWS"
  vpc_subnet_range   = "172.20.0.0/20"
  secondary_vpc_cidr = "100.64.0.0/19" # Optional: secondary CIDR for pod networking
  workspace_ids      = []
  timeouts = {    # Optional timeouts for create, update, and delete
    create = "3h" # Timeout after 3 hours if the cluster is not created
    update = "2h" # Timeout after 2 hours if the cluster is not updated
    delete = "1h" # Timeout after 1 hour if the cluster is not deleted
  }
}

resource "astro_cluster" "azure_example" {
  type             = "DEDICATED"
  name             = "my first azure cluster"
  region           = "westus2"
  cloud_provider   = "AZURE"
  vpc_subnet_range = "172.20.0.0/19"
  workspace_ids    = ["clv4wcf6f003u01m3zp7gsvzg"]
}

resource "astro_cluster" "aws_dr_example" {
  type                = "DEDICATED"
  name                = "my dr-enabled cluster"
  region              = "us-east-1"
  cloud_provider      = "AWS"
  vpc_subnet_range    = "172.20.0.0/20"
  workspace_ids       = []
  is_dr_enabled       = true
  dr_region           = "us-west-2"
  dr_vpc_subnet_range = "172.21.0.0/20"
}

resource "astro_cluster" "gcp_example" {
  type                  = "DEDICATED"
  name                  = "my first gcp cluster"
  region                = "us-central1"
  cloud_provider        = "GCP"
  pod_subnet_range      = "172.21.0.0/19"
  service_peering_range = "172.23.0.0/20"
  service_subnet_range  = "172.22.0.0/22"
  vpc_subnet_range      = "172.20.0.0/22"
  workspace_ids         = []
}

// Import an existing cluster
import {
  id = "clozc036j01to01jrlgvuf98d" // ID of the existing cluster
  to = astro_cluster.imported_cluster
}
resource "astro_cluster" "imported_cluster" {
  type                  = "DEDICATED"
  name                  = "an existing cluster to import"
  region                = "us-central1"
  cloud_provider        = "GCP"
  pod_subnet_range      = "172.21.0.0/19"
  service_peering_range = "172.23.0.0/20"
  service_subnet_range  = "172.22.0.0/22"
  vpc_subnet_range      = "172.20.0.0/22"
  workspace_ids         = []
}