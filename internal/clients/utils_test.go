package clients_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
)

func TestUnit_CoreRequestEditor(t *testing.T) {
	ctx := context.Background()
	req := http.Request{URL: &url.URL{Path: "/path"}, Header: make(http.Header)}

	err := clients.CoreRequestEditor(ctx, &req, "http://localhost", "token", "v1")
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost/path", req.URL.String())
	assert.Equal(t, "Bearer token", req.Header.Get("authorization"))
	assert.Equal(t, "terraform-provider-astro", req.Header.Get("x-astro-client-identifier"))
	assert.NotEmpty(t, req.Header.Get("x-astro-client-version"))
	assert.NotEmpty(t, req.Header.Get("x-client-os-identifier"))
	assert.Equal(t, "terraform-provider-astro/v1", req.Header.Get("User-Agent"))
}

func TestUnit_NormalizeAPIError(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		resp           *http.Response
		body           []byte
		expectedStatus int
		expectError    bool
		errorContains  string
	}{
		{
			name:           "SuccessfulRequest",
			resp:           &http.Response{StatusCode: http.StatusOK},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "HttpResponseNil",
			resp:           nil,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			errorContains:  "failed to perform request",
		},
		{
			name:           "ResponseNot200",
			resp:           &http.Response{StatusCode: http.StatusNotFound},
			body:           []byte(`{"message": "error", "requestId": "123"}`),
			expectedStatus: http.StatusNotFound,
			expectError:    true,
			errorContains:  "requestId: 123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, diag := clients.NormalizeAPIError(ctx, tt.resp, tt.body)

			assert.Equal(t, tt.expectedStatus, status)
			if tt.expectError {
				assert.NotNil(t, diag)
				if tt.errorContains != "" {
					assert.Contains(t, diag.Detail(), tt.errorContains)
				}
			} else {
				assert.Nil(t, diag)
			}
		})
	}
}

func TestUnit_NormalizeAPIResponseWithBody(t *testing.T) {
	ctx := context.Background()

	type payload struct{ Id string }

	tests := []struct {
		name           string
		resp           *http.Response
		body           []byte
		json200        *payload
		expectedStatus int
		expectError    bool
		errorContains  string
	}{
		{
			name:           "SuccessWithBody",
			resp:           &http.Response{StatusCode: http.StatusOK},
			json200:        &payload{Id: "id"},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "OkWithoutParsedBody",
			resp:           &http.Response{StatusCode: http.StatusOK},
			json200:        nil,
			expectedStatus: http.StatusOK,
			expectError:    true,
			errorContains:  "no response body",
		},
		{
			name:           "NoContentHasNoBody",
			resp:           &http.Response{StatusCode: http.StatusNoContent},
			json200:        nil,
			expectedStatus: http.StatusNoContent,
			expectError:    true,
			errorContains:  "no response body",
		},
		{
			name:           "CreatedWithoutParsedBody",
			resp:           &http.Response{StatusCode: http.StatusCreated},
			json200:        nil,
			expectedStatus: http.StatusCreated,
			expectError:    true,
			errorContains:  "no response body",
		},
		{
			// The status error wins; the body check is never reached
			name:           "HttpResponseNilTakesPrecedence",
			resp:           nil,
			json200:        nil,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			errorContains:  "failed to perform request",
		},
		{
			// 404 must still reach the caller for RemoveResource
			name:           "StatusErrorTakesPrecedence",
			resp:           &http.Response{StatusCode: http.StatusNotFound},
			body:           []byte(`{"message": "not found", "requestId": "123"}`),
			json200:        nil,
			expectedStatus: http.StatusNotFound,
			expectError:    true,
			errorContains:  "requestId: 123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, diag := clients.NormalizeAPIResponseWithBody(ctx, tt.resp, tt.body, tt.json200, "read thing")

			assert.Equal(t, tt.expectedStatus, status)
			if tt.expectError {
				assert.NotNil(t, diag)
				if tt.errorContains != "" {
					assert.Contains(t, diag.Detail(), tt.errorContains)
				}
			} else {
				assert.Nil(t, diag)
			}
		})
	}
}
