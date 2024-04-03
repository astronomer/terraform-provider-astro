package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func CoreRequestEditor(
	ctx context.Context,
	req *http.Request,
	baseUrl, token, version string,
) error {
	os := runtime.GOOS
	arch := runtime.GOARCH
	requestURL, err := url.Parse(baseUrl + req.URL.String())
	if err != nil {
		return fmt.Errorf("baseUrl '%v' is invalid", baseUrl)
	}
	req.URL = requestURL
	req.Header.Add("authorization", fmt.Sprintf("Bearer %v", token))
	req.Header.Add("x-astro-client-identifier", "astronomer-terraform-provider")
	req.Header.Add("x-astro-client-version", version)
	req.Header.Add("x-client-os-identifier", os+"-"+arch)
	req.Header.Add("User-Agent", fmt.Sprintf("astronomer-terraform-provider/%s", version))
	return nil
}

func NormalizeAPIError(
	ctx context.Context,
	httpResp *http.Response,
	body []byte,
) (int, diag.Diagnostic) {
	if httpResp == nil {
		tflog.Error(
			ctx,
			"failed to perform request",
			map[string]interface{}{"error": "http response is nil"},
		)
		return http.StatusInternalServerError, diag.NewErrorDiagnostic(
			"Client error",
			"failed to perform request",
		)
	}
	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNoContent &&
		httpResp.StatusCode != http.StatusCreated {
		type Error struct {
			Message   string `json:"message"`
			RequestId string `json:"requestId"`
		}
		decode := Error{}
		err := json.NewDecoder(bytes.NewReader(body)).Decode(&decode)
		if err != nil {
			tflog.Error(
				ctx,
				"failed to decode error response",
				map[string]interface{}{"error": err.Error()},
			)
			return httpResp.StatusCode, diag.NewErrorDiagnostic(
				"Client error",
				fmt.Sprintf("failed to perform request, status: %v", httpResp.StatusCode),
			)
		}
		tflog.Error(
			ctx,
			"Client error",
			map[string]interface{}{
				"message":   decode.Message,
				"status":    httpResp.StatusCode,
				"requestId": decode.RequestId,
			},
		)
		return httpResp.StatusCode, diag.NewErrorDiagnostic(
			"Client error",
			fmt.Sprintf(
				"%v, status: %v, requestId: %v",
				decode.Message,
				httpResp.StatusCode,
				decode.RequestId,
			),
		)
	}
	return httpResp.StatusCode, nil
}
