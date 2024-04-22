---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "astronomer_clusters Data Source - astronomer"
subcategory: ""
description: |-
  Clusters data source
---

# astronomer_clusters (Data Source)

Clusters data source

## Example Usage

```terraform
data "astronomer_clusters" "example_clusters" {}

data "astronomer_clusters" "example_clusters_filter_by_names" {
  names = ["my cluster"]
}

data "astronomer_clusters" "example_clusters_filter_by_cloud_provider" {
  cloud_provider = ["AWS"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `cloud_provider` (String)
- `names` (List of String)

### Read-Only

- `clusters` (Attributes List) (see [below for nested schema](#nestedatt--clusters))

<a id="nestedatt--clusters"></a>
### Nested Schema for `clusters`

Required:

- `id` (String) Cluster identifier

Read-Only:

- `cloud_provider` (String) Cluster cloud provider
- `created_at` (String) Cluster creation timestamp
- `db_instance_type` (String) Cluster database instance type
- `is_limited` (Boolean) Whether the cluster is limited
- `metadata` (Attributes) Cluster metadata (see [below for nested schema](#nestedatt--clusters--metadata))
- `name` (String) Cluster name
- `node_pools` (Attributes List) Cluster node pools (see [below for nested schema](#nestedatt--clusters--node_pools))
- `pod_subnet_range` (String) Cluster pod subnet range
- `provider_account` (String) Cluster provider account
- `region` (String) Cluster region
- `service_peering_range` (String) Cluster service peering range
- `service_subnet_range` (String) Cluster service subnet range
- `status` (String) Cluster status
- `tags` (Attributes List) Cluster tags (see [below for nested schema](#nestedatt--clusters--tags))
- `tenant_id` (String) Cluster tenant ID
- `type` (String) Cluster type
- `updated_at` (String) Cluster last updated timestamp
- `vpc_subnet_range` (String) Cluster VPC subnet range
- `workspace_ids` (List of String) Cluster workspace IDs

<a id="nestedatt--clusters--metadata"></a>
### Nested Schema for `clusters.metadata`

Read-Only:

- `external_ips` (List of String) Cluster external IPs
- `oidc_issuer_url` (String) Cluster OIDC issuer URL


<a id="nestedatt--clusters--node_pools"></a>
### Nested Schema for `clusters.node_pools`

Read-Only:

- `cloud_provider` (String) Node pool cloud provider
- `cluster_id` (String) Node pool cluster identifier
- `created_at` (String) Node pool creation timestamp
- `id` (String) Node pool identifier
- `is_default` (Boolean) Whether the node pool is the default node pool of the cluster
- `max_node_count` (Number) Node pool maximum node count
- `name` (String) Node pool name
- `node_instance_type` (String) Node pool node instance type
- `supported_astro_machines` (List of String) Node pool supported Astro machines
- `updated_at` (String) Node pool last updated timestamp


<a id="nestedatt--clusters--tags"></a>
### Nested Schema for `clusters.tags`

Read-Only:

- `key` (String) Cluster tag key
- `value` (String) Cluster tag value