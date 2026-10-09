package resources

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	platform_v1 "github.com/astronomer/terraform-provider-astro/internal/clients/platform_v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	testOrganizationId = "clx46acvv060e01ilddqlbsmc"
	testClusterId      = "cmv169sro0iw001jbxl7xbj88"
)

// resetConnection hijacks and closes the connection without writing a response, so the client
// sees a transport-level failure rather than an HTTP error status. This reproduces the
// "read: connection reset by peer" failures seen against the live API during long cluster
// polls, which is the case the retry budget exists for.
func resetConnection(t *testing.T, w http.ResponseWriter) {
	t.Helper()
	hijacker, ok := w.(http.Hijacker)
	require.True(t, ok, "test server response writer must support hijacking")
	conn, _, err := hijacker.Hijack()
	require.NoError(t, err)
	require.NoError(t, conn.Close())
}

func clusterBody(t *testing.T, status platform_v1.ClusterStatus) []byte {
	t.Helper()
	body, err := json.Marshal(platform_v1.Cluster{
		Id:             testClusterId,
		Name:           "test-cluster",
		CloudProvider:  platform_v1.ClusterCloudProviderAWS,
		Region:         "us-east-1",
		VpcSubnetRange: "172.20.0.0/22",
		Status:         status,
		Type:           platform_v1.ClusterTypeDEDICATED,
		OrganizationId: testOrganizationId,
		CreatedAt:      time.Unix(0, 0).UTC(),
		UpdatedAt:      time.Unix(0, 0).UTC(),
	})
	require.NoError(t, err)
	return body
}

// pollServer serves cluster responses, resetting the connection for the first
// failuresBeforeSuccess requests. It reports how many requests it received.
func pollServer(t *testing.T, failuresBeforeSuccess int32, status platform_v1.ClusterStatus) (*platform_v1.ClientWithResponses, *int32) {
	t.Helper()
	var requests int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&requests, 1) <= failuresBeforeSuccess {
			resetConnection(t, w)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(clusterBody(t, status))
	}))
	t.Cleanup(server.Close)

	client, err := platform_v1.NewClientWithResponses(server.URL)
	require.NoError(t, err)
	return client, &requests
}

// A single dropped connection must not fail the poll. Before this was handled, one reset
// during a multi-hour cluster create aborted the whole apply.
func TestUnit_ClusterResourceRefreshFunc_RetriesTransientTransportFailure(t *testing.T) {
	client, requests := pollServer(t, 1, platform_v1.ClusterStatusCREATED)
	refresh := ClusterResourceRefreshFunc(context.Background(), client, testOrganizationId, testClusterId)

	// First poll: connection reset. Reported as "not found yet" so polling continues.
	result, state, err := refresh()
	require.NoError(t, err, "a transient transport failure must not abort the poll")
	assert.Nil(t, result)
	assert.Empty(t, state)

	// Second poll: the API answers.
	result, state, err = refresh()
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, string(platform_v1.ClusterStatusCREATED), state)
	assert.Equal(t, int32(2), atomic.LoadInt32(requests))
}

// Failures below the budget are absorbed; the poll still reaches the target state.
func TestUnit_ClusterResourceRefreshFunc_RetriesUpToBudget(t *testing.T) {
	client, requests := pollServer(t, maxConsecutivePollTransportFailures-1, platform_v1.ClusterStatusCREATED)
	refresh := ClusterResourceRefreshFunc(context.Background(), client, testOrganizationId, testClusterId)

	for i := 0; i < maxConsecutivePollTransportFailures-1; i++ {
		_, _, err := refresh()
		require.NoError(t, err, "failure %d should still be within the retry budget", i+1)
	}

	result, state, err := refresh()
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, string(platform_v1.ClusterStatusCREATED), state)
	assert.Equal(t, int32(maxConsecutivePollTransportFailures), atomic.LoadInt32(requests))
}

// An API that stays unreachable must still fail, rather than spinning until the operation's
// own multi-hour timeout. The error has to name the underlying cause.
func TestUnit_ClusterResourceRefreshFunc_GivesUpAfterBudgetExhausted(t *testing.T) {
	// Never succeeds.
	client, _ := pollServer(t, int32(maxConsecutivePollTransportFailures)+10, platform_v1.ClusterStatusCREATED)
	refresh := ClusterResourceRefreshFunc(context.Background(), client, testOrganizationId, testClusterId)

	for i := 0; i < maxConsecutivePollTransportFailures-1; i++ {
		_, _, err := refresh()
		require.NoError(t, err)
	}

	_, _, err := refresh()
	require.Error(t, err, "the poll must fail once the retry budget is exhausted")
	assert.Contains(t, err.Error(), "consecutive attempts")
	assert.Contains(t, err.Error(), testClusterId)
}

// The budget bounds a *run* of failures, not the total. A long poll that hits isolated blips
// far apart must never accumulate its way into a spurious failure.
func TestUnit_ClusterResourceRefreshFunc_ResetsBudgetAfterSuccessfulPoll(t *testing.T) {
	var requests int32
	// Every other request fails, so the budget is never consecutively exhausted.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&requests, 1)%2 == 1 {
			resetConnection(t, w)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(clusterBody(t, platform_v1.ClusterStatusCREATING))
	}))
	t.Cleanup(server.Close)

	client, err := platform_v1.NewClientWithResponses(server.URL)
	require.NoError(t, err)
	refresh := ClusterResourceRefreshFunc(context.Background(), client, testOrganizationId, testClusterId)

	// Far more alternating failures than the budget would allow if it never reset.
	for i := 0; i < maxConsecutivePollTransportFailures*4; i++ {
		_, _, err := refresh()
		require.NoError(t, err, "alternating failures must never exhaust the budget (iteration %d)", i)
	}
}

// A cancelled context is terminal: retrying cannot succeed, and the caller needs the error.
func TestUnit_ClusterResourceRefreshFunc_DoesNotRetryWhenContextIsDone(t *testing.T) {
	client, requests := pollServer(t, 100, platform_v1.ClusterStatusCREATED)

	ctx, cancel := context.WithCancel(context.Background())
	refresh := ClusterResourceRefreshFunc(ctx, client, testOrganizationId, testClusterId)
	cancel()

	_, _, err := refresh()
	require.Error(t, err, "a cancelled context must surface immediately")
	assert.Equal(t, int32(0), atomic.LoadInt32(requests), "no request should be attempted after cancellation")
}

// Real API errors must keep failing fast - the retry budget is only for transport failures.
func TestUnit_ClusterResourceRefreshFunc_DoesNotRetryApiErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom","requestId":"abc"}`))
	}))
	t.Cleanup(server.Close)

	client, err := platform_v1.NewClientWithResponses(server.URL)
	require.NoError(t, err)
	refresh := ClusterResourceRefreshFunc(context.Background(), client, testOrganizationId, testClusterId)

	_, _, err = refresh()
	require.Error(t, err, "an API error must abort the poll immediately")
	assert.Contains(t, err.Error(), "error getting cluster")
}

// A deleted cluster still reports DELETED rather than being mistaken for a transport failure.
func TestUnit_ClusterResourceRefreshFunc_ReportsDeleted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found","requestId":"abc"}`))
	}))
	t.Cleanup(server.Close)

	client, err := platform_v1.NewClientWithResponses(server.URL)
	require.NoError(t, err)
	refresh := ClusterResourceRefreshFunc(context.Background(), client, testOrganizationId, testClusterId)

	result, state, err := refresh()
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "DELETED", state)
}

// The retry budget must stay below the SDK's NotFoundChecks default: transport failures are
// reported to WaitForStateContext as nil results, so if the budget were the larger of the two
// the SDK would fail the operation first, with a misleading "couldn't find resource" error.
func TestUnit_MaxConsecutivePollTransportFailuresIsBelowNotFoundChecks(t *testing.T) {
	const sdkDefaultNotFoundChecks = 20 // retry.StateChangeConf applies this when NotFoundChecks is 0

	var conf retry.StateChangeConf
	assert.Zero(t, conf.NotFoundChecks, "the cluster pollers leave NotFoundChecks unset, so the SDK default applies")
	assert.Less(t, maxConsecutivePollTransportFailures, sdkDefaultNotFoundChecks)
}
