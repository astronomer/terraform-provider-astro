package labs

// Hand-authored client bindings for the labs allowed-ip-address-ranges bulk endpoints
// (astronomer/astro#39781). These follow the same shape oapi-codegen produces elsewhere in
// this package, but are maintained by hand because the "AllowedIpAddressRanges" tag hasn't
// been added to the labs OpenAPI include-tags in the Makefile yet (the labs spec isn't
// vendored in this repo). Once that tag is added, re-run `make generate-labs-client` and
// delete this file in favor of the generated types/client methods.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/oapi-codegen/runtime"
)

// AllowedIpAddressRange defines model for AllowedIpAddressRange.
type AllowedIpAddressRange struct {
	// CreatedAt The time when the allowed IP address range was created in UTC, formatted as `YYYY-MM-DDTHH:MM:SSZ`.
	CreatedAt time.Time           `json:"createdAt"`
	CreatedBy BasicSubjectProfile `json:"createdBy"`

	// Id The allowed IP address range's ID.
	Id string `json:"id"`

	// IpAddressRange The allowed IP address range in CIDR format.
	IpAddressRange string `json:"ipAddressRange"`

	// OrganizationId The allowed IP address range's Organization ID.
	OrganizationId string `json:"organizationId"`

	// UpdatedAt The time when the allowed IP address range was last updated in UTC, formatted as `YYYY-MM-DDTHH:MM:SSZ`.
	UpdatedAt time.Time           `json:"updatedAt"`
	UpdatedBy BasicSubjectProfile `json:"updatedBy"`
}

// AllowedIpAddressRangesList defines model for the bulk create response body.
type AllowedIpAddressRangesList struct {
	AllowedIpAddressRanges []AllowedIpAddressRange `json:"allowedIpAddressRanges"`
}

// CreateAllowedIpAddressRangesRequest defines model for CreateAllowedIpAddressRangesRequest.
// At most 1000 CIDRs per request; the request is applied atomically.
type CreateAllowedIpAddressRangesRequest struct {
	AllowedIpAddressRanges []string `json:"allowedIpAddressRanges"`
}

// DeleteAllowedIpAddressRangesRequest defines model for DeleteAllowedIpAddressRangesRequest.
type DeleteAllowedIpAddressRangesRequest struct {
	AllowedIpAddressRangeIds []string `json:"allowedIpAddressRangeIds"`
}

// ListAllowedIpAddressRangesParams defines parameters for ListAllowedIpAddressRanges.
type ListAllowedIpAddressRangesParams struct {
	Offset *int `form:"offset,omitempty" json:"offset,omitempty"`
	Limit  *int `form:"limit,omitempty" json:"limit,omitempty"`
}

type LabsCreateAllowedIpAddressRangesJSONRequestBody = CreateAllowedIpAddressRangesRequest
type LabsDeleteAllowedIpAddressRangesJSONRequestBody = DeleteAllowedIpAddressRangesRequest

func allowedIpAddressRangesPath(organizationId string) (string, error) {
	pathParam0, err := runtime.StyleParamWithLocation("simple", false, "organizationId", runtime.ParamLocationPath, organizationId)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/organizations/%s/allowed-ip-address-ranges", pathParam0), nil
}

// NewLabsListAllowedIpAddressRangesRequest generates a request for LabsListAllowedIpAddressRanges.
func NewLabsListAllowedIpAddressRangesRequest(server string, organizationId string, params *ListAllowedIpAddressRangesParams) (*http.Request, error) {
	operationPath, err := allowedIpAddressRangesPath(organizationId)
	if err != nil {
		return nil, err
	}
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	if params != nil {
		queryValues := queryURL.Query()
		if params.Offset != nil {
			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "offset", runtime.ParamLocationQuery, *params.Offset); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}
		}
		if params.Limit != nil {
			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "limit", runtime.ParamLocationQuery, *params.Limit); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}
		}
		queryURL.RawQuery = queryValues.Encode()
	}

	return http.NewRequest("GET", queryURL.String(), nil)
}

// NewLabsCreateAllowedIpAddressRangesRequest calls the generic builder with an application/json body.
func NewLabsCreateAllowedIpAddressRangesRequest(server string, organizationId string, body LabsCreateAllowedIpAddressRangesJSONRequestBody) (*http.Request, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return NewLabsCreateAllowedIpAddressRangesRequestWithBody(server, organizationId, "application/json", bytes.NewReader(buf))
}

// NewLabsCreateAllowedIpAddressRangesRequestWithBody generates requests with any type of body.
func NewLabsCreateAllowedIpAddressRangesRequestWithBody(server string, organizationId string, contentType string, body io.Reader) (*http.Request, error) {
	operationPath, err := allowedIpAddressRangesPath(organizationId)
	if err != nil {
		return nil, err
	}
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", contentType)
	return req, nil
}

// NewLabsDeleteAllowedIpAddressRangesRequest calls the generic builder with an application/json body.
func NewLabsDeleteAllowedIpAddressRangesRequest(server string, organizationId string, body LabsDeleteAllowedIpAddressRangesJSONRequestBody) (*http.Request, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return NewLabsDeleteAllowedIpAddressRangesRequestWithBody(server, organizationId, "application/json", bytes.NewReader(buf))
}

// NewLabsDeleteAllowedIpAddressRangesRequestWithBody generates requests with any type of body.
func NewLabsDeleteAllowedIpAddressRangesRequestWithBody(server string, organizationId string, contentType string, body io.Reader) (*http.Request, error) {
	operationPath, err := allowedIpAddressRangesPath(organizationId)
	if err != nil {
		return nil, err
	}
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", contentType)
	return req, nil
}

func (c *Client) LabsListAllowedIpAddressRanges(ctx context.Context, organizationId string, params *ListAllowedIpAddressRangesParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewLabsListAllowedIpAddressRangesRequest(c.Server, organizationId, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) LabsCreateAllowedIpAddressRangesWithBody(ctx context.Context, organizationId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewLabsCreateAllowedIpAddressRangesRequestWithBody(c.Server, organizationId, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) LabsCreateAllowedIpAddressRanges(ctx context.Context, organizationId string, body LabsCreateAllowedIpAddressRangesJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewLabsCreateAllowedIpAddressRangesRequest(c.Server, organizationId, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) LabsDeleteAllowedIpAddressRangesWithBody(ctx context.Context, organizationId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewLabsDeleteAllowedIpAddressRangesRequestWithBody(c.Server, organizationId, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) LabsDeleteAllowedIpAddressRanges(ctx context.Context, organizationId string, body LabsDeleteAllowedIpAddressRangesJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewLabsDeleteAllowedIpAddressRangesRequest(c.Server, organizationId, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

type LabsListAllowedIpAddressRangesResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *AllowedIpAddressRangesList
	JSON400      *Error
	JSON401      *Error
	JSON403      *Error
	JSON404      *Error
	JSON500      *Error
}

func (r LabsListAllowedIpAddressRangesResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

func (r LabsListAllowedIpAddressRangesResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type LabsCreateAllowedIpAddressRangesResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *AllowedIpAddressRangesList
	JSON400      *Error
	JSON401      *Error
	JSON403      *Error
	JSON404      *Error
	JSON409      *Error
	JSON500      *Error
}

func (r LabsCreateAllowedIpAddressRangesResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

func (r LabsCreateAllowedIpAddressRangesResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type LabsDeleteAllowedIpAddressRangesResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON400      *Error
	JSON401      *Error
	JSON403      *Error
	JSON404      *Error
	JSON500      *Error
}

func (r LabsDeleteAllowedIpAddressRangesResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

func (r LabsDeleteAllowedIpAddressRangesResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

func (c *ClientWithResponses) LabsListAllowedIpAddressRangesWithResponse(ctx context.Context, organizationId string, params *ListAllowedIpAddressRangesParams, reqEditors ...RequestEditorFn) (*LabsListAllowedIpAddressRangesResponse, error) {
	rsp, err := c.LabsListAllowedIpAddressRanges(ctx, organizationId, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseLabsListAllowedIpAddressRangesResponse(rsp)
}

func (c *ClientWithResponses) LabsCreateAllowedIpAddressRangesWithBodyWithResponse(ctx context.Context, organizationId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*LabsCreateAllowedIpAddressRangesResponse, error) {
	rsp, err := c.LabsCreateAllowedIpAddressRangesWithBody(ctx, organizationId, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseLabsCreateAllowedIpAddressRangesResponse(rsp)
}

func (c *ClientWithResponses) LabsCreateAllowedIpAddressRangesWithResponse(ctx context.Context, organizationId string, body LabsCreateAllowedIpAddressRangesJSONRequestBody, reqEditors ...RequestEditorFn) (*LabsCreateAllowedIpAddressRangesResponse, error) {
	rsp, err := c.LabsCreateAllowedIpAddressRanges(ctx, organizationId, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseLabsCreateAllowedIpAddressRangesResponse(rsp)
}

func (c *ClientWithResponses) LabsDeleteAllowedIpAddressRangesWithBodyWithResponse(ctx context.Context, organizationId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*LabsDeleteAllowedIpAddressRangesResponse, error) {
	rsp, err := c.LabsDeleteAllowedIpAddressRangesWithBody(ctx, organizationId, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseLabsDeleteAllowedIpAddressRangesResponse(rsp)
}

func (c *ClientWithResponses) LabsDeleteAllowedIpAddressRangesWithResponse(ctx context.Context, organizationId string, body LabsDeleteAllowedIpAddressRangesJSONRequestBody, reqEditors ...RequestEditorFn) (*LabsDeleteAllowedIpAddressRangesResponse, error) {
	rsp, err := c.LabsDeleteAllowedIpAddressRanges(ctx, organizationId, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseLabsDeleteAllowedIpAddressRangesResponse(rsp)
}

// ParseLabsListAllowedIpAddressRangesResponse parses an HTTP response from a LabsListAllowedIpAddressRangesWithResponse call.
func ParseLabsListAllowedIpAddressRangesResponse(rsp *http.Response) (*LabsListAllowedIpAddressRangesResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &LabsListAllowedIpAddressRangesResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest AllowedIpAddressRangesList
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 401:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON401 = &dest
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 403:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON403 = &dest
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 404:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON404 = &dest
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest
	}

	return response, nil
}

// ParseLabsCreateAllowedIpAddressRangesResponse parses an HTTP response from a LabsCreateAllowedIpAddressRangesWithResponse call.
func ParseLabsCreateAllowedIpAddressRangesResponse(rsp *http.Response) (*LabsCreateAllowedIpAddressRangesResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &LabsCreateAllowedIpAddressRangesResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest AllowedIpAddressRangesList
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 401:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON401 = &dest
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 403:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON403 = &dest
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 409:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON409 = &dest
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest
	}

	return response, nil
}

// ParseLabsDeleteAllowedIpAddressRangesResponse parses an HTTP response from a LabsDeleteAllowedIpAddressRangesWithResponse call.
func ParseLabsDeleteAllowedIpAddressRangesResponse(rsp *http.Response) (*LabsDeleteAllowedIpAddressRangesResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &LabsDeleteAllowedIpAddressRangesResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 401:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON401 = &dest
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 403:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON403 = &dest
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 404:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON404 = &dest
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest
	}

	return response, nil
}
