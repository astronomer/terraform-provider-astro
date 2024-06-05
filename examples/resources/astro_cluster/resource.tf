resource "astro_cluster" "aws_example" {
  type             = "DEDICATED"
  name             = "my first aws cluster"
  region           = "us-east-1"
  cloud_provider   = "AWS"
  db_instance_type = "db.m6g.large"
  vpc_subnet_range = "172.20.0.0/20"
  workspace_ids    = []
  timeouts = {
    create = "3h"
    update = "2h"
    delete = "1h"
  }
}

resource "astro_cluster" "azure_example" {
  type             = "DEDICATED"
  name             = "my first azure cluster"
  region           = "westus2"
  cloud_provider   = "AZURE"
  db_instance_type = "Standard_D2ds_v4"
  vpc_subnet_range = "172.20.0.0/19"
  workspace_ids    = ["clv4wcf6f003u01m3zp7gsvzg"]
}

resource "astro_cluster" "gcp_example" {
  type                  = "DEDICATED"
  name                  = "my first gcp cluster"
  region                = "us-central1"
  cloud_provider        = "GCP"
  db_instance_type      = "Small General Purpose"
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
  db_instance_type      = "Small General Purpose"
  pod_subnet_range      = "172.21.0.0/19"
  service_peering_range = "172.23.0.0/20"
  service_subnet_range  = "172.22.0.0/22"
  vpc_subnet_range      = "172.20.0.0/22"
  workspace_ids         = []
}