---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "astro_cluster_options Data Source - astro"
subcategory: ""
description: |-
  ClusterOptions data source
---

# astro_cluster_options (Data Source)

ClusterOptions data source

## Example Usage

```terraform
data "astro_cluster_options" "example_cluster_options" {
  type = "HYBRID"
}

data "astro_cluster_options" "example_cluster_options_filter_by_provider" {
  type           = "HYBRID"
  cloud_provider = "AWS"
}

# Output the cluster options value using terraform apply
output "cluster_options" {
  value = data.astro_cluster_options.example_cluster_options
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `type` (String)

### Optional

- `cloud_provider` (String)

### Read-Only

- `cluster_options` (Attributes Set) (see [below for nested schema](#nestedatt--cluster_options))

<a id="nestedatt--cluster_options"></a>
### Nested Schema for `cluster_options`

Read-Only:

- `database_instances` (Attributes Set) ClusterOption database instances (see [below for nested schema](#nestedatt--cluster_options--database_instances))
- `default_database_instance` (Attributes) ClusterOption default database instance (see [below for nested schema](#nestedatt--cluster_options--default_database_instance))
- `default_node_instance` (Attributes) ClusterOption default node instance (see [below for nested schema](#nestedatt--cluster_options--default_node_instance))
- `default_pod_subnet_range` (String) ClusterOption default pod subnet range
- `default_region` (Attributes) ClusterOption default region (see [below for nested schema](#nestedatt--cluster_options--default_region))
- `default_service_peering_range` (String) ClusterOption default service peering range
- `default_service_subnet_range` (String) ClusterOption default service subnet range
- `default_vpc_subnet_range` (String) ClusterOption default vps subnet range
- `node_count_default` (Number) ClusterOption node count default
- `node_count_max` (Number) ClusterOption node count max
- `node_count_min` (Number) ClusterOption node count min
- `node_instances` (Attributes Set) ClusterOption node instances (see [below for nested schema](#nestedatt--cluster_options--node_instances))
- `provider` (String) ClusterOption provider
- `regions` (Attributes Set) ClusterOption regions (see [below for nested schema](#nestedatt--cluster_options--regions))

<a id="nestedatt--cluster_options--database_instances"></a>
### Nested Schema for `cluster_options.database_instances`

Read-Only:

- `cpu` (Number) Provider instance cpu
- `memory` (String) Provider instance memory
- `name` (String) Provider instance name


<a id="nestedatt--cluster_options--default_database_instance"></a>
### Nested Schema for `cluster_options.default_database_instance`

Read-Only:

- `cpu` (Number) Provider instance cpu
- `memory` (String) Provider instance memory
- `name` (String) Provider instance name


<a id="nestedatt--cluster_options--default_node_instance"></a>
### Nested Schema for `cluster_options.default_node_instance`

Read-Only:

- `cpu` (Number) Provider instance cpu
- `memory` (String) Provider instance memory
- `name` (String) Provider instance name


<a id="nestedatt--cluster_options--default_region"></a>
### Nested Schema for `cluster_options.default_region`

Read-Only:

- `banned_instances` (Set of String) Region banned instances
- `limited` (Boolean) Region is limited bool
- `name` (String) Region is limited bool


<a id="nestedatt--cluster_options--node_instances"></a>
### Nested Schema for `cluster_options.node_instances`

Read-Only:

- `cpu` (Number) Provider instance cpu
- `memory` (String) Provider instance memory
- `name` (String) Provider instance name


<a id="nestedatt--cluster_options--regions"></a>
### Nested Schema for `cluster_options.regions`

Read-Only:

- `banned_instances` (Set of String) Region banned instances
- `limited` (Boolean) Region is limited bool
- `name` (String) Region is limited bool
