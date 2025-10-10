package resources

import (
	"context"
	"fmt"
	"net/http"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// ClusterResourceRefreshFunc returns a retry.StateRefreshFunc that polls the platform API for the cluster status
// If the cluster is not found, it returns "DELETED" status
// If the cluster is found, it returns the cluster status
// If there is an error, it returns the error
// WaitForStateContext will keep polling until the target status is reached, the timeout is reached or an err is returned
func ClusterResourceRefreshFunc(ctx context.Context, platformClient *platform.ClientWithResponses, organizationId string, clusterId string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		cluster, err := platformClient.GetClusterWithResponse(ctx, organizationId, clusterId)
		if err != nil {
			tflog.Error(ctx, "failed to get cluster while polling for cluster 'CREATED' status", map[string]interface{}{"error": err})
			return nil, "", err
		}
		statusCode, diagnostic := clients.NormalizeAPIError(ctx, cluster.HTTPResponse, cluster.Body)
		if statusCode == http.StatusNotFound {
			return &platform.Cluster{}, "DELETED", nil
		}
		if diagnostic != nil {
			return nil, "", fmt.Errorf("error getting cluster %s", diagnostic.Detail())
		}
		if cluster != nil && cluster.JSON200 != nil {
			switch cluster.JSON200.Status {
			case platform.ClusterStatusCREATED:
				return cluster.JSON200, string(cluster.JSON200.Status), nil
			case platform.ClusterStatusUPDATEFAILED, platform.ClusterStatusCREATEFAILED:
				return cluster.JSON200, string(cluster.JSON200.Status), fmt.Errorf("cluster mutation failed for cluster '%v'", cluster.JSON200.Id)
			case platform.ClusterStatusCREATING, platform.ClusterStatusUPDATING, platform.ClusterStatusUPGRADEPENDING:
				return cluster.JSON200, string(cluster.JSON200.Status), nil
			case platform.ClusterStatusACCESSDENIED:
				return cluster.JSON200, string(cluster.JSON200.Status), fmt.Errorf("access denied for cluster '%v'", cluster.JSON200.Id)
			case "": // Handle empty status as transient initialization state
				tflog.Debug(ctx, "cluster status is empty, treating as pending", map[string]interface{}{"clusterId": cluster.JSON200.Id})
				return cluster.JSON200, string(platform.ClusterStatusCREATING), nil
			default:
				return cluster.JSON200, string(cluster.JSON200.Status), fmt.Errorf("unexpected cluster status '%v' for cluster '%v'", cluster.JSON200.Status, cluster.JSON200.Id)
			}
		}
		return nil, "", fmt.Errorf("error getting cluster %s", clusterId)
	}
}
