package resources

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/astronomer/terraform-provider-astro/internal/clients/labs"
)

// bulkCreate's IDs are zipped against request keys by index, so a chunk that yields fewer IDs than
// it requested must abort rather than leave a hole - otherwise later keys silently take the wrong
// alert's ID and Terraform manages (and destroys) the wrong objects.
func TestUnit_bulkCreate_ShortOrMissingChunkAborts(t *testing.T) {
	// alertsFor builds a JSON body carrying n alerts with predictable ids.
	alertsFor := func(prefix string, n int) map[string]any {
		out := make([]map[string]any, 0, n)
		for i := 0; i < n; i++ {
			out = append(out, map[string]any{
				"id": prefix, "name": prefix, "entityId": "e", "entityType": "DEPLOYMENT",
				"organizationId": "o", "severity": "CRITICAL", "type": "DAG_FAILURE",
				"rules": map[string]any{}, "createdAt": "2024-01-01T00:00:00Z",
				"updatedAt": "2024-01-01T00:00:00Z",
				"createdBy": map[string]any{"id": "u"}, "updatedBy": map[string]any{"id": "u"},
			})
		}
		return map[string]any{"alerts": out}
	}

	tests := []struct {
		name      string
		respond   func(w http.ResponseWriter, call int)
		wantIds   int
		wantError string
	}{
		{
			name: "success returns one id per request",
			respond: func(w http.ResponseWriter, call int) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(alertsFor("ok", alertsBulkCreateLimit))
			},
			wantIds: alertsBulkCreateLimit,
		},
		{
			// A 204 leaves JSON200 nil. Previously the chunk was skipped and the loop continued.
			name: "no-content chunk aborts instead of being skipped",
			respond: func(w http.ResponseWriter, call int) {
				w.WriteHeader(http.StatusNoContent)
			},
			wantIds:   0,
			wantError: "no response body",
		},
		{
			name: "short chunk aborts",
			respond: func(w http.ResponseWriter, call int) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(alertsFor("short", alertsBulkCreateLimit-1))
			},
			wantIds:   0,
			wantError: "the API returned 29",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			call := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				call++
				tc.respond(w, call)
			}))
			t.Cleanup(srv.Close)
			client, err := labs.NewLabsClient(srv.URL, "token", "test")
			require.NoError(t, err)

			r := &alertsResource{labsClient: client, organizationId: "org"}
			reqs := make([]labs.CreateAlertRequest, alertsBulkCreateLimit)
			ids, diags := r.bulkCreate(context.Background(), reqs)

			assert.Len(t, ids, tc.wantIds)
			if tc.wantError == "" {
				assert.False(t, diags.HasError())
				return
			}
			require.True(t, diags.HasError())
			assert.Contains(t, diags[0].Detail(), tc.wantError)
		})
	}
}

// The prefix property the positional zip depends on: when a later chunk fails, the IDs from
// earlier chunks are still returned, and still line up with the first N requests.
func TestUnit_bulkCreate_KeepsValidPrefixOnLaterChunkFailure(t *testing.T) {
	call := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		if call == 1 {
			w.Header().Set("Content-Type", "application/json")
			alerts := make([]map[string]any, 0, alertsBulkCreateLimit)
			for i := 0; i < alertsBulkCreateLimit; i++ {
				alerts = append(alerts, map[string]any{
					"id": "first", "name": "n", "entityId": "e", "entityType": "DEPLOYMENT",
					"organizationId": "o", "severity": "CRITICAL", "type": "DAG_FAILURE",
					"rules": map[string]any{}, "createdAt": "2024-01-01T00:00:00Z",
					"updatedAt": "2024-01-01T00:00:00Z",
					"createdBy": map[string]any{"id": "u"}, "updatedBy": map[string]any{"id": "u"},
				})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"alerts": alerts})
			return
		}
		w.WriteHeader(http.StatusNoContent) // second chunk comes back bodiless
	}))
	t.Cleanup(srv.Close)
	client, err := labs.NewLabsClient(srv.URL, "token", "test")
	require.NoError(t, err)

	r := &alertsResource{labsClient: client, organizationId: "org"}
	ids, diags := r.bulkCreate(context.Background(), make([]labs.CreateAlertRequest, alertsBulkCreateLimit+5))

	assert.True(t, diags.HasError())
	// Exactly the first chunk, so keys[0:30] still map to their own alerts.
	assert.Len(t, ids, alertsBulkCreateLimit)
}
