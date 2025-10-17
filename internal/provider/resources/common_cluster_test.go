package resources_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/resources"
	"github.com/stretchr/testify/assert"
)

// mockHTTPClient is a simple mock for testing HTTP responses
type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

// createMockClusterResponse creates a mock HTTP response with the given cluster
func createMockClusterResponse(cluster *platform.Cluster, statusCode int) *http.Response {
	body, _ := json.Marshal(cluster)
	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(string(body))),
		Header:     header,
	}
}

func TestUnit_ClusterResourceRefreshFunc_EmptyStatus(t *testing.T) {
	ctx := context.Background()
	clusterId := "test-cluster-id"
	organizationId := "test-org-id"

	// Create a mock cluster with empty status
	mockCluster := &platform.Cluster{
		Id:     clusterId,
		Name:   "test-cluster",
		Status: "", // Empty status
	}

	// Create mock HTTP client
	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return createMockClusterResponse(mockCluster, http.StatusOK), nil
		},
	}

	// Create platform client with mock
	platformClient, err := platform.NewClientWithResponses(
		"http://localhost",
		platform.WithHTTPClient(mockClient),
	)
	assert.NoError(t, err)

	// Call the refresh function
	refreshFunc := resources.ClusterResourceRefreshFunc(ctx, platformClient, organizationId, clusterId)
	result, state, err := refreshFunc()

	// Assertions for empty status handling
	assert.NoError(t, err, "Empty status should not return an error, it should be treated as pending")
	assert.NotNil(t, result, "Result should not be nil")
	assert.Equal(t, string(platform.ClusterStatusCREATING), state, "Empty status should be treated as CREATING")

	// Verify the returned cluster
	returnedCluster, ok := result.(*platform.Cluster)
	assert.True(t, ok, "Result should be a *platform.Cluster")
	assert.Equal(t, clusterId, returnedCluster.Id)
}

func TestUnit_ClusterResourceRefreshFunc_CreatingStatus(t *testing.T) {
	ctx := context.Background()
	clusterId := "test-cluster-id"
	organizationId := "test-org-id"

	mockCluster := &platform.Cluster{
		Id:     clusterId,
		Name:   "test-cluster",
		Status: platform.ClusterStatusCREATING,
	}

	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return createMockClusterResponse(mockCluster, http.StatusOK), nil
		},
	}

	platformClient, err := platform.NewClientWithResponses(
		"http://localhost",
		platform.WithHTTPClient(mockClient),
	)
	assert.NoError(t, err)

	refreshFunc := resources.ClusterResourceRefreshFunc(ctx, platformClient, organizationId, clusterId)
	result, state, err := refreshFunc()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, string(platform.ClusterStatusCREATING), state)
}

func TestUnit_ClusterResourceRefreshFunc_CreatedStatus(t *testing.T) {
	ctx := context.Background()
	clusterId := "test-cluster-id"
	organizationId := "test-org-id"

	mockCluster := &platform.Cluster{
		Id:     clusterId,
		Name:   "test-cluster",
		Status: platform.ClusterStatusCREATED,
	}

	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return createMockClusterResponse(mockCluster, http.StatusOK), nil
		},
	}

	platformClient, err := platform.NewClientWithResponses(
		"http://localhost",
		platform.WithHTTPClient(mockClient),
	)
	assert.NoError(t, err)

	refreshFunc := resources.ClusterResourceRefreshFunc(ctx, platformClient, organizationId, clusterId)
	result, state, err := refreshFunc()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, string(platform.ClusterStatusCREATED), state)
}

func TestUnit_ClusterResourceRefreshFunc_UpdatingStatus(t *testing.T) {
	ctx := context.Background()
	clusterId := "test-cluster-id"
	organizationId := "test-org-id"

	mockCluster := &platform.Cluster{
		Id:     clusterId,
		Name:   "test-cluster",
		Status: platform.ClusterStatusUPDATING,
	}

	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return createMockClusterResponse(mockCluster, http.StatusOK), nil
		},
	}

	platformClient, err := platform.NewClientWithResponses(
		"http://localhost",
		platform.WithHTTPClient(mockClient),
	)
	assert.NoError(t, err)

	refreshFunc := resources.ClusterResourceRefreshFunc(ctx, platformClient, organizationId, clusterId)
	result, state, err := refreshFunc()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, string(platform.ClusterStatusUPDATING), state)
}

func TestUnit_ClusterResourceRefreshFunc_UpgradePendingStatus(t *testing.T) {
	ctx := context.Background()
	clusterId := "test-cluster-id"
	organizationId := "test-org-id"

	mockCluster := &platform.Cluster{
		Id:     clusterId,
		Name:   "test-cluster",
		Status: platform.ClusterStatusUPGRADEPENDING,
	}

	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return createMockClusterResponse(mockCluster, http.StatusOK), nil
		},
	}

	platformClient, err := platform.NewClientWithResponses(
		"http://localhost",
		platform.WithHTTPClient(mockClient),
	)
	assert.NoError(t, err)

	refreshFunc := resources.ClusterResourceRefreshFunc(ctx, platformClient, organizationId, clusterId)
	result, state, err := refreshFunc()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, string(platform.ClusterStatusUPGRADEPENDING), state)
}

func TestUnit_ClusterResourceRefreshFunc_CreateFailedStatus(t *testing.T) {
	ctx := context.Background()
	clusterId := "test-cluster-id"
	organizationId := "test-org-id"

	mockCluster := &platform.Cluster{
		Id:     clusterId,
		Name:   "test-cluster",
		Status: platform.ClusterStatusCREATEFAILED,
	}

	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return createMockClusterResponse(mockCluster, http.StatusOK), nil
		},
	}

	platformClient, err := platform.NewClientWithResponses(
		"http://localhost",
		platform.WithHTTPClient(mockClient),
	)
	assert.NoError(t, err)

	refreshFunc := resources.ClusterResourceRefreshFunc(ctx, platformClient, organizationId, clusterId)
	result, state, err := refreshFunc()

	assert.Error(t, err, "CREATE_FAILED status should return an error")
	assert.Contains(t, err.Error(), "cluster mutation failed")
	assert.Equal(t, string(platform.ClusterStatusCREATEFAILED), state)
	assert.NotNil(t, result, "Result should still be returned even on error")
}

func TestUnit_ClusterResourceRefreshFunc_UpdateFailedStatus(t *testing.T) {
	ctx := context.Background()
	clusterId := "test-cluster-id"
	organizationId := "test-org-id"

	mockCluster := &platform.Cluster{
		Id:     clusterId,
		Name:   "test-cluster",
		Status: platform.ClusterStatusUPDATEFAILED,
	}

	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return createMockClusterResponse(mockCluster, http.StatusOK), nil
		},
	}

	platformClient, err := platform.NewClientWithResponses(
		"http://localhost",
		platform.WithHTTPClient(mockClient),
	)
	assert.NoError(t, err)

	refreshFunc := resources.ClusterResourceRefreshFunc(ctx, platformClient, organizationId, clusterId)
	result, state, err := refreshFunc()

	assert.Error(t, err, "UPDATE_FAILED status should return an error")
	assert.Contains(t, err.Error(), "cluster mutation failed")
	assert.Equal(t, string(platform.ClusterStatusUPDATEFAILED), state)
	assert.NotNil(t, result, "Result should still be returned even on error")
}

func TestUnit_ClusterResourceRefreshFunc_AccessDeniedStatus(t *testing.T) {
	ctx := context.Background()
	clusterId := "test-cluster-id"
	organizationId := "test-org-id"

	mockCluster := &platform.Cluster{
		Id:     clusterId,
		Name:   "test-cluster",
		Status: platform.ClusterStatusACCESSDENIED,
	}

	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return createMockClusterResponse(mockCluster, http.StatusOK), nil
		},
	}

	platformClient, err := platform.NewClientWithResponses(
		"http://localhost",
		platform.WithHTTPClient(mockClient),
	)
	assert.NoError(t, err)

	refreshFunc := resources.ClusterResourceRefreshFunc(ctx, platformClient, organizationId, clusterId)
	result, state, err := refreshFunc()

	assert.Error(t, err, "ACCESS_DENIED status should return an error")
	assert.Contains(t, err.Error(), "access denied")
	assert.Equal(t, string(platform.ClusterStatusACCESSDENIED), state)
	assert.NotNil(t, result, "Result should still be returned even on error")
}

func TestUnit_ClusterResourceRefreshFunc_UnknownStatus(t *testing.T) {
	ctx := context.Background()
	clusterId := "test-cluster-id"
	organizationId := "test-org-id"

	mockCluster := &platform.Cluster{
		Id:     clusterId,
		Name:   "test-cluster",
		Status: platform.ClusterStatus("UNKNOWN_STATUS"),
	}

	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return createMockClusterResponse(mockCluster, http.StatusOK), nil
		},
	}

	platformClient, err := platform.NewClientWithResponses(
		"http://localhost",
		platform.WithHTTPClient(mockClient),
	)
	assert.NoError(t, err)

	refreshFunc := resources.ClusterResourceRefreshFunc(ctx, platformClient, organizationId, clusterId)
	result, state, err := refreshFunc()

	assert.Error(t, err, "Unknown status should return an error")
	assert.Contains(t, err.Error(), "unexpected cluster status")
	assert.Equal(t, "UNKNOWN_STATUS", state)
	assert.NotNil(t, result, "Result should still be returned even on error")
}

func TestUnit_ClusterResourceRefreshFunc_NotFound(t *testing.T) {
	ctx := context.Background()
	clusterId := "test-cluster-id"
	organizationId := "test-org-id"

	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(`{"message": "Not found"}`)),
				Header:     make(http.Header),
			}, nil
		},
	}

	platformClient, err := platform.NewClientWithResponses(
		"http://localhost",
		platform.WithHTTPClient(mockClient),
	)
	assert.NoError(t, err)

	refreshFunc := resources.ClusterResourceRefreshFunc(ctx, platformClient, organizationId, clusterId)
	result, state, err := refreshFunc()

	assert.NoError(t, err, "404 should not return error, cluster is considered deleted")
	assert.Equal(t, "DELETED", state)
	assert.NotNil(t, result)
}

func TestUnit_ClusterResourceRefreshFunc_HTTPError(t *testing.T) {
	ctx := context.Background()
	clusterId := "test-cluster-id"
	organizationId := "test-org-id"

	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("network error")
		},
	}

	platformClient, err := platform.NewClientWithResponses(
		"http://localhost",
		platform.WithHTTPClient(mockClient),
	)
	assert.NoError(t, err)

	refreshFunc := resources.ClusterResourceRefreshFunc(ctx, platformClient, organizationId, clusterId)
	result, state, err := refreshFunc()

	assert.Error(t, err, "Network errors should be returned")
	assert.Contains(t, err.Error(), "network error")
	assert.Nil(t, result)
	assert.Equal(t, "", state)
}
