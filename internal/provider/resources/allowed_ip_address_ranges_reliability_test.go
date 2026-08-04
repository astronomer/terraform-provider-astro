package resources

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/clients/labs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rangesHandler is a configurable fake Astro API: `list` is what the iam GET endpoint returns, and
// `postStatus` is the status the labs POST (bulk create) endpoint returns.
type rangesHandler struct {
	list       []map[string]any
	postStatus int
}

func (h rangesHandler) server(t *testing.T) (*iam.ClientWithResponses, *labs.ClientWithResponses) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if h.list == nil {
				// simulate a success status with a non-JSON body (proxy/gateway page)
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("<html>not json</html>"))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"allowedIpAddressRanges": h.list, "limit": 1000, "offset": 0, "totalCount": len(h.list),
			})
		case http.MethodPost:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(h.postStatus)
			_ = json.NewEncoder(w).Encode(map[string]any{"message": "conflict", "requestId": "test"})
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	t.Cleanup(srv.Close)
	iamClient, err := iam.NewIamClient(srv.URL, "token", "test")
	require.NoError(t, err)
	labsClient, err := labs.NewLabsClient(srv.URL, "token", "test")
	require.NoError(t, err)
	return iamClient, labsClient
}

// #5: a 409 on bulk create is idempotent-success only when the ranges are already present.
func TestUnit_bulkCreate_409IsIdempotentWhenPresent(t *testing.T) {
	ctx := context.Background()

	t.Run("409 + range already present -> success", func(t *testing.T) {
		iamClient, labsClient := rangesHandler{
			list:       []map[string]any{{"id": "id-1", "ipAddressRange": "10.0.0.0/8"}},
			postStatus: http.StatusConflict,
		}.server(t)
		r := &allowedIpAddressRangesResource{iamClient: iamClient, labsClient: labsClient, organizationId: "org"}
		diags := r.bulkCreate(ctx, []string{"10.0.0.0/8"})
		assert.False(t, diags.HasError(), "409 with the range already present should be treated as success: %v", diags)
	})

	t.Run("409 + range NOT present -> surfaces the conflict", func(t *testing.T) {
		iamClient, labsClient := rangesHandler{
			list:       []map[string]any{}, // empty: the requested CIDR is not present
			postStatus: http.StatusConflict,
		}.server(t)
		r := &allowedIpAddressRangesResource{iamClient: iamClient, labsClient: labsClient, organizationId: "org"}
		diags := r.bulkCreate(ctx, []string{"10.0.0.0/8"})
		assert.True(t, diags.HasError(), "409 with the range absent must not be swallowed")
	})
}

// #6: a success status with no parseable body must error, not read as an empty list.
func TestUnit_listAll_nonJSONSuccessErrorsNotEmpty(t *testing.T) {
	ctx := context.Background()
	iamClient, labsClient := rangesHandler{list: nil}.server(t) // nil -> non-JSON 200
	r := &allowedIpAddressRangesResource{iamClient: iamClient, labsClient: labsClient, organizationId: "org"}
	_, diags := r.listAll(ctx)
	assert.True(t, diags.HasError(), "a 200 with a non-JSON body must error rather than return an empty (state-wiping) list")
}

// #7: idsForCidrs warns about ranges it cannot resolve rather than silently under-deleting.
func TestUnit_idsForCidrs_warnsOnMissing(t *testing.T) {
	ctx := context.Background()
	iamClient, labsClient := rangesHandler{
		list: []map[string]any{{"id": "id-1", "ipAddressRange": "10.0.0.0/8"}},
	}.server(t)
	r := &allowedIpAddressRangesResource{iamClient: iamClient, labsClient: labsClient, organizationId: "org"}

	ids, diags := r.idsForCidrs(ctx, []string{"10.0.0.0/8", "192.168.0.0/16"})
	assert.Equal(t, []string{"id-1"}, ids)
	assert.False(t, diags.HasError())
	require.NotEmpty(t, diags.Warnings(), "an unresolved CIDR should produce a warning")
	assert.True(t, strings.Contains(diags.Warnings()[0].Detail(), "192.168.0.0/16"),
		"the warning should name the missing range")
}
