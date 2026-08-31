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
// channel's ID and Terraform manages (and destroys) the wrong objects.
func TestUnit_notificationChannelsBulkCreate_ShortOrMissingChunkAborts(t *testing.T) {
	channelsFor := func(prefix string, n int) map[string]any {
		out := make([]map[string]any, 0, n)
		for i := 0; i < n; i++ {
			out = append(out, map[string]any{
				"id": prefix, "name": prefix, "entityId": "e", "entityType": "DEPLOYMENT",
				"organizationId": "o", "type": "SLACK", "isShared": false,
				"definition": map[string]any{}, "createdAt": "2024-01-01T00:00:00Z",
				"updatedAt": "2024-01-01T00:00:00Z",
			})
		}
		return map[string]any{"notificationChannels": out}
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
				_ = json.NewEncoder(w).Encode(channelsFor("ok", notificationChannelsBulkCreateLimit))
			},
			wantIds: notificationChannelsBulkCreateLimit,
		},
		{
			// A 204 leaves JSON200 nil; the chunk must abort, not be skipped.
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
				_ = json.NewEncoder(w).Encode(channelsFor("short", notificationChannelsBulkCreateLimit-1))
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

			r := &notificationChannelsResource{labsClient: client, organizationId: "org"}
			reqs := make([]labs.CreateNotificationChannelRequest, notificationChannelsBulkCreateLimit)
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

// When a later chunk fails, the IDs from earlier chunks are still returned and still line up with
// the first N requests, so the positional key->id zip stays correct.
func TestUnit_notificationChannelsBulkCreate_KeepsValidPrefixOnLaterChunkFailure(t *testing.T) {
	call := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		if call == 1 {
			w.Header().Set("Content-Type", "application/json")
			channels := make([]map[string]any, 0, notificationChannelsBulkCreateLimit)
			for i := 0; i < notificationChannelsBulkCreateLimit; i++ {
				channels = append(channels, map[string]any{
					"id": "first", "name": "n", "entityId": "e", "entityType": "DEPLOYMENT",
					"organizationId": "o", "type": "SLACK", "isShared": false,
					"definition": map[string]any{}, "createdAt": "2024-01-01T00:00:00Z",
					"updatedAt": "2024-01-01T00:00:00Z",
				})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"notificationChannels": channels})
			return
		}
		w.WriteHeader(http.StatusNoContent) // second chunk comes back bodiless
	}))
	t.Cleanup(srv.Close)
	client, err := labs.NewLabsClient(srv.URL, "token", "test")
	require.NoError(t, err)

	r := &notificationChannelsResource{labsClient: client, organizationId: "org"}
	ids, diags := r.bulkCreate(context.Background(), make([]labs.CreateNotificationChannelRequest, notificationChannelsBulkCreateLimit+5))

	assert.True(t, diags.HasError())
	// Exactly the first chunk, so keys[0:30] still map to their own channels.
	assert.Len(t, ids, notificationChannelsBulkCreateLimit)
}

// bulkDelete tolerates a 404 (already deleted) but surfaces other errors, and chunks by the limit.
func TestUnit_notificationChannelsBulkDelete(t *testing.T) {
	t.Run("404 is tolerated", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		t.Cleanup(srv.Close)
		client, err := labs.NewLabsClient(srv.URL, "token", "test")
		require.NoError(t, err)

		r := &notificationChannelsResource{labsClient: client, organizationId: "org"}
		diags := r.bulkDelete(context.Background(), []string{"a", "b"})
		assert.False(t, diags.HasError())
	})

	t.Run("server error surfaces", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		t.Cleanup(srv.Close)
		client, err := labs.NewLabsClient(srv.URL, "token", "test")
		require.NoError(t, err)

		r := &notificationChannelsResource{labsClient: client, organizationId: "org"}
		diags := r.bulkDelete(context.Background(), []string{"a"})
		assert.True(t, diags.HasError())
	})
}
