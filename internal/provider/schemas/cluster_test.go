package schemas

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// is_private_network_egress_enabled must be Optional so AWS clusters can turn Private Network
// Egress on, and Computed so a cluster that never sets it still records what the API reports.
func TestClusterResourceSchemaAttributes_IsPrivateNetworkEgressEnabledIsOptionalAndComputed(t *testing.T) {
	attributes := ClusterResourceSchemaAttributes(context.Background())

	egress, ok := attributes["is_private_network_egress_enabled"]
	assert.True(t, ok, "is_private_network_egress_enabled should exist in the cluster resource schema")
	assert.True(t, egress.IsOptional(), "is_private_network_egress_enabled should be Optional")
	assert.True(t, egress.IsComputed(), "is_private_network_egress_enabled should be Computed so omitting it does not conflict with the API value")
}

func TestClusterDataSourceSchemaAttributes_IsPrivateNetworkEgressEnabledIsComputed(t *testing.T) {
	attributes := ClusterDataSourceSchemaAttributes()

	egress, ok := attributes["is_private_network_egress_enabled"]
	assert.True(t, ok, "is_private_network_egress_enabled should exist in the cluster data source schema")
	assert.True(t, egress.IsComputed(), "is_private_network_egress_enabled should be Computed")
}

// The astro_clusters data source builds each element from ClustersElementAttributeTypes, so an
// attribute missing there makes the whole data source fail at runtime.
func TestClustersElementAttributeTypesMatchClusterDataSourceSchema(t *testing.T) {
	elementAttributeTypes := ClustersElementAttributeTypes()

	_, ok := elementAttributeTypes["is_private_network_egress_enabled"]
	assert.True(t, ok, "is_private_network_egress_enabled should exist in ClustersElementAttributeTypes")

	for name := range ClusterDataSourceSchemaAttributes() {
		_, ok := elementAttributeTypes[name]
		assert.True(t, ok, "cluster data source attribute %q is missing from ClustersElementAttributeTypes", name)
	}
	assert.Len(t, elementAttributeTypes, len(ClusterDataSourceSchemaAttributes()))
}
