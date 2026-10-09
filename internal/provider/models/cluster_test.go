package models_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	platform_v1 "github.com/astronomer/terraform-provider-astro/internal/clients/platform_v1"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/samber/lo"
)

func clusterResponse(cloudProvider platform_v1.ClusterCloudProvider, egress *bool) *platform_v1.Cluster {
	return &platform_v1.Cluster{
		Id:                            "clt1y2z3a4b5c6d7e8f9g0h1i",
		Name:                          "test-cluster",
		CloudProvider:                 cloudProvider,
		Region:                        "us-east-1",
		VpcSubnetRange:                "172.20.0.0/22",
		Status:                        platform_v1.ClusterStatusCREATED,
		Type:                          platform_v1.ClusterTypeDEDICATED,
		OrganizationId:                "clt1y2z3a4b5c6d7e8f9g0h1j",
		IsPrivateNetworkEgressEnabled: egress,
	}
}

// Private Network Egress reported by the API is surfaced as-is.
func TestUnit_ClusterReadFromResponse_PrivateNetworkEgressEnabled(t *testing.T) {
	var data models.ClusterResource
	require.False(t, data.ReadFromResponse(
		context.Background(),
		clusterResponse(platform_v1.ClusterCloudProviderAWS, lo.ToPtr(true)),
	).HasError())

	assert.False(t, data.IsPrivateNetworkEgressEnabled.IsNull())
	assert.True(t, data.IsPrivateNetworkEgressEnabled.ValueBool())
}

// An AWS cluster with the flag omitted has Private Network Egress disabled. Reporting that as
// null would fail an apply that configured `false` as an inconsistent result.
func TestUnit_ClusterReadFromResponse_PrivateNetworkEgressOmittedOnAwsIsFalse(t *testing.T) {
	var data models.ClusterResource
	require.False(t, data.ReadFromResponse(
		context.Background(),
		clusterResponse(platform_v1.ClusterCloudProviderAWS, nil),
	).HasError())

	assert.False(t, data.IsPrivateNetworkEgressEnabled.IsNull())
	assert.False(t, data.IsPrivateNetworkEgressEnabled.ValueBool())
}

// Private Network Egress is an AWS-only capability, so it stays null elsewhere.
func TestUnit_ClusterReadFromResponse_PrivateNetworkEgressIsNullOnNonAws(t *testing.T) {
	for _, cloudProvider := range []platform_v1.ClusterCloudProvider{
		platform_v1.ClusterCloudProviderGCP,
		platform_v1.ClusterCloudProviderAZURE,
	} {
		var data models.ClusterResource
		require.False(t, data.ReadFromResponse(
			context.Background(),
			clusterResponse(cloudProvider, nil),
		).HasError())

		assert.True(t, data.IsPrivateNetworkEgressEnabled.IsNull(), "expected null for %v", cloudProvider)
	}
}

// The data source read path must agree with the resource read path.
func TestUnit_ClusterDataSourceReadFromResponse_PrivateNetworkEgress(t *testing.T) {
	var data models.ClusterDataSource
	require.False(t, data.ReadFromResponse(
		context.Background(),
		clusterResponse(platform_v1.ClusterCloudProviderAWS, lo.ToPtr(true)),
	).HasError())
	assert.True(t, data.IsPrivateNetworkEgressEnabled.ValueBool())

	var gcpData models.ClusterDataSource
	require.False(t, gcpData.ReadFromResponse(
		context.Background(),
		clusterResponse(platform_v1.ClusterCloudProviderGCP, nil),
	).HasError())
	assert.True(t, gcpData.IsPrivateNetworkEgressEnabled.IsNull())
}
