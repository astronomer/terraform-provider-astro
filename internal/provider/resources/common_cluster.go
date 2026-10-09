package resources

import (
	"context"
	"fmt"
	"net/http"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	platform_v1 "github.com/astronomer/terraform-provider-astro/internal/clients/platform_v1"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// maxConsecutivePollTransportFailures is how many back-to-back transport-level failures the
// cluster poller tolerates before giving up.
//
// Cluster mutations are polled for up to three hours, so the poll loop is exposed to every
// transient network fault in that window: a reset connection, a dropped keep-alive, a brief
// DNS or proxy blip. WaitForStateContext aborts on the first error a refresh func returns, so
// without this a single blip fails an otherwise healthy multi-hour apply.
//
// The counter resets after any successful poll, so this bounds a run of consecutive failures,
// not the total over the life of the operation. A genuinely unreachable API still fails, just
// after this many attempts instead of one.
const maxConsecutivePollTransportFailures = 5

// ClusterResourceRefreshFunc returns a retry.StateRefreshFunc that polls the platform API for the cluster status
// If the cluster is not found, it returns "DELETED" status
// If the cluster is found, it returns the cluster status
// If there is an error, it returns the error
// WaitForStateContext will keep polling until the target status is reached, the timeout is reached or an err is returned
//
// Transport-level failures are the exception: they are retried up to
// maxConsecutivePollTransportFailures times rather than failing the operation outright. Such a
// failure is reported to WaitForStateContext as a nil result, which it treats as "not found
// yet" and keeps polling. That consumes one of its NotFoundChecks (20 by default, and also
// reset by any successful poll), so the retry budget here must stay below that to be the
// binding limit.
func ClusterResourceRefreshFunc(ctx context.Context, platformV1Client *platform_v1.ClientWithResponses, organizationId string, clusterId string) retry.StateRefreshFunc {
	// Scoped per returned func, so each cluster operation gets its own budget.
	consecutiveTransportFailures := 0

	return func() (any, string, error) {
		cluster, err := platformV1Client.GetClusterWithResponse(ctx, organizationId, clusterId)
		if err != nil {
			// A cancelled or timed-out context is terminal - retrying cannot succeed, and the
			// caller needs the error rather than a poll loop that spins until its own timeout.
			if ctxErr := ctx.Err(); ctxErr != nil {
				tflog.Error(ctx, "stopped polling for cluster status: context is done", map[string]interface{}{"error": err})
				return nil, "", err
			}

			consecutiveTransportFailures++
			if consecutiveTransportFailures >= maxConsecutivePollTransportFailures {
				tflog.Error(ctx, "failed to get cluster while polling for cluster status", map[string]interface{}{
					"error":                        err,
					"consecutiveTransportFailures": consecutiveTransportFailures,
				})
				return nil, "", fmt.Errorf(
					"unable to reach the API while polling cluster '%v' after %d consecutive attempts, got error: %w",
					clusterId, consecutiveTransportFailures, err,
				)
			}

			tflog.Warn(ctx, "failed to get cluster while polling for cluster status, retrying", map[string]interface{}{
				"error":                        err,
				"consecutiveTransportFailures": consecutiveTransportFailures,
				"maxConsecutiveFailures":       maxConsecutivePollTransportFailures,
			})
			// nil result, nil error: WaitForStateContext keeps polling.
			return nil, "", nil
		}
		consecutiveTransportFailures = 0

		statusCode, diagnostic := clients.NormalizeAPIError(ctx, cluster.HTTPResponse, cluster.Body)
		if statusCode == http.StatusNotFound {
			return &platform_v1.Cluster{}, "DELETED", nil
		}
		if diagnostic != nil {
			return nil, "", fmt.Errorf("error getting cluster %s", diagnostic.Detail())
		}
		if cluster != nil && cluster.JSON200 != nil {
			switch cluster.JSON200.Status {
			case platform_v1.ClusterStatusCREATED:
				return cluster.JSON200, string(cluster.JSON200.Status), nil
			case platform_v1.ClusterStatusUPDATEFAILED, platform_v1.ClusterStatusCREATEFAILED:
				return cluster.JSON200, string(cluster.JSON200.Status), fmt.Errorf("cluster mutation failed for cluster '%v'", cluster.JSON200.Id)
			case platform_v1.ClusterStatusFAILOVERFAILED:
				return cluster.JSON200, string(cluster.JSON200.Status), fmt.Errorf("cluster failover failed for cluster '%v'", cluster.JSON200.Id)
			case platform_v1.ClusterStatusCREATING, platform_v1.ClusterStatusUPDATING, platform_v1.ClusterStatusUPGRADEPENDING, platform_v1.ClusterStatusFAILINGOVER:
				return cluster.JSON200, string(cluster.JSON200.Status), nil
			case platform_v1.ClusterStatusACCESSDENIED:
				return cluster.JSON200, string(cluster.JSON200.Status), fmt.Errorf("access denied for cluster '%v'", cluster.JSON200.Id)
			default:
				return cluster.JSON200, string(cluster.JSON200.Status), fmt.Errorf("unexpected cluster status '%v' for cluster '%v'", cluster.JSON200.Status, cluster.JSON200.Id)
			}
		}
		return nil, "", fmt.Errorf("error getting cluster %s", clusterId)
	}
}
