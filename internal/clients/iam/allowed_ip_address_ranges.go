package iam

// Hand-authored client binding for the iam allowed-ip-address-ranges list endpoint
// (GET /organizations/{organizationId}/allowed-ip-address-ranges, tag "AllowedIpAddressRange").
//
// This is maintained by hand rather than generated because adding "AllowedIpAddressRange" to
// this package's oapi-codegen -include-tags list (see the Makefile) causes oapi-codegen v2.1.0 to
// rename the unrelated ListUsersParamsSorts enum constants (their overlapping value strings with
// ListAllowedIpAddressRangesParamsSorts collide in the tool's constant-naming pass), which breaks
// internal/provider/datasources/data_source_users_list.go. That's outside the scope of this
// change, so this binding is hand-written instead - matching the shape oapi-codegen produces
// elsewhere in this package - until the coegen/tag collision is resolved upstream (or the two
// enums are reconciled) and `make api_client_gen` can pick this endpoint up cleanly.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/oapi-codegen/runtime"
)

// AllowedIpAddressRange defines model for AllowedIpAddressRange.
type AllowedIpAddressRange struct {
	// CreatedAt The time when the allowed IP address range was created in UTC, formatted as `YYYY-MM-DDTHH:MM:SSZ`.
	CreatedAt string `json:"createdAt"`

	// CreatedBy The entity that created the range.
	CreatedBy *BasicSubjectProfile `json:"createdBy,omitempty"`

	// Id The allowed IP address range's ID.
	Id string `json:"id"`

	// IpAddressRange The allowed IP address range in CIDR format.
	IpAddressRange string `json:"ipAddressRange"`

	// OrganizationId The allowed IP address range's Organization ID.
	OrganizationId string `json:"organizationId"`

	// UpdatedAt The time when the allowed IP address range was last updated in UTC, formatted as `YYYY-MM-DDTHH:MM:SSZ`.
	UpdatedAt string `json:"updatedAt"`

	// UpdatedBy The entity that last updated the range.
	UpdatedBy *BasicSubjectProfile `json:"updatedBy,omitempty"`
}

// AllowedIpAddressRangesPaginated defines model for AllowedIpAddressRangesPaginated.
type AllowedIpAddressRangesPaginated struct {
	AllowedIpAddressRanges []AllowedIpAddressRange `json:"allowedIpAddressRanges"`
	Limit                  int                     `json:"limit"`
	Offset                 int                     `json:"offset"`
	TotalCount             int                     `json:"totalCount"`
}

// ListAllowedIpAddressRangesParamsSorts defines parameters for ListAllowedIpAddressRanges.
type ListAllowedIpAddressRangesParamsSorts string

const (
	ListAllowedIpAddressRangesParamsSortsCreatedAtAsc  ListAllowedIpAddressRangesParamsSorts = "createdAt:asc"
	ListAllowedIpAddressRangesParamsSortsCreatedAtDesc ListAllowedIpAddressRangesParamsSorts = "createdAt:desc"
	ListAllowedIpAddressRangesParamsSortsIpAddressAsc  ListAllowedIpAddressRangesParamsSorts = "ipAddress:asc"
	ListAllowedIpAddressRangesParamsSortsIpAddressDesc ListAllowedIpAddressRangesParamsSorts = "ipAddress:desc"
	ListAllowedIpAddressRangesParamsSortsUpdatedAtAsc  ListAllowedIpAddressRangesParamsSorts = "updatedAt:asc"
	ListAllowedIpAddressRangesParamsSortsUpdatedAtDesc ListAllowedIpAddressRangesParamsSorts = "updatedAt:desc"
)

// ListAllowedIpAddressRangesParams defines parameters for ListAllowedIpAddressRanges.
type ListAllowedIpAddressRangesParams struct {
	Offset *int                                     `form:"offset,omitempty" json:"offset,omitempty"`
	Limit  *int                                     `form:"limit,omitempty" json:"limit,omitempty"`
	Sorts  *[]ListAllowedIpAddressRangesParamsSorts `form:"sorts,omitempty" json:"sorts,omitempty"`
}

// NewListAllowedIpAddressRangesRequest generates a request for ListAllowedIpAddressRanges.
func NewListAllowedIpAddressRangesRequest(server string, organizationId string, params *ListAllowedIpAddressRangesParams) (*http.Request, error) {
	pathParam0, err := runtime.StyleParamWithLocation("simple", false, "organizationId", runtime.ParamLocationPath, organizationId)
	if err != nil {
		return nil, err
	}
	operationPath := fmt.Sprintf("/organizations/%s/allowed-ip-address-ranges", pathParam0)
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
		if params.Sorts != nil {
			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "sorts", runtime.ParamLocationQuery, *params.Sorts); err != nil {
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

func (c *Client) ListAllowedIpAddressRanges(ctx context.Context, organizationId string, params *ListAllowedIpAddressRangesParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListAllowedIpAddressRangesRequest(c.Server, organizationId, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

type ListAllowedIpAddressRangesResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *AllowedIpAddressRangesPaginated
	JSON400      *Error
	JSON401      *Error
	JSON403      *Error
	JSON500      *Error
}

func (r ListAllowedIpAddressRangesResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

func (r ListAllowedIpAddressRangesResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

func (c *ClientWithResponses) ListAllowedIpAddressRangesWithResponse(ctx context.Context, organizationId string, params *ListAllowedIpAddressRangesParams, reqEditors ...RequestEditorFn) (*ListAllowedIpAddressRangesResponse, error) {
	rsp, err := c.ListAllowedIpAddressRanges(ctx, organizationId, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListAllowedIpAddressRangesResponse(rsp)
}

// ParseListAllowedIpAddressRangesResponse parses an HTTP response from a ListAllowedIpAddressRangesWithResponse call.
func ParseListAllowedIpAddressRangesResponse(rsp *http.Response) (*ListAllowedIpAddressRangesResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListAllowedIpAddressRangesResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest AllowedIpAddressRangesPaginated
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
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest
	}

	return response, nil
}
