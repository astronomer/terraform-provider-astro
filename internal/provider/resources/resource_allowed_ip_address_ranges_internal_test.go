package resources

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/clients/labs"
	mocks_iam "github.com/astronomer/terraform-provider-astro/internal/mocks/iam"
	mocks_labs "github.com/astronomer/terraform-provider-astro/internal/mocks/labs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testOrgId = "test-org-id"

// makeRanges builds n AllowedIpAddressRange records with predictable CIDR/ID pairs.
func makeRanges(n int) []iam.AllowedIpAddressRange {
	out := make([]iam.AllowedIpAddressRange, n)
	for i := 0; i < n; i++ {
		out[i] = iam.AllowedIpAddressRange{
			Id:             fmt.Sprintf("id-%d", i),
			IpAddressRange: fmt.Sprintf("10.%d.%d.0/24", i/256, i%256),
		}
	}
	return out
}

func offsetIs(want int) interface{} {
	return mock.MatchedBy(func(p *iam.ListAllowedIpAddressRangesParams) bool {
		return p != nil && p.Offset != nil && *p.Offset == want
	})
}

func okCreateResp() (*labs.LabsCreateAllowedIpAddressRangesResponse, error) {
	return &labs.LabsCreateAllowedIpAddressRangesResponse{HTTPResponse: &http.Response{StatusCode: http.StatusOK}}, nil
}

func okDeleteResp() (*labs.LabsDeleteAllowedIpAddressRangesResponse, error) {
	return &labs.LabsDeleteAllowedIpAddressRangesResponse{HTTPResponse: &http.Response{StatusCode: http.StatusOK}}, nil
}

func listResp(ranges []iam.AllowedIpAddressRange, total int) (*iam.ListAllowedIpAddressRangesResponse, error) {
	return &iam.ListAllowedIpAddressRangesResponse{
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
		JSON200: &iam.AllowedIpAddressRangesPaginated{
			AllowedIpAddressRanges: ranges,
			TotalCount:             total,
		},
	}, nil
}

func TestUnit_listAllRanges_paginatesAcrossPages(t *testing.T) {
	ctx := context.Background()
	iamMock := new(mocks_iam.ClientWithResponsesInterface)

	page1 := makeRanges(allowedIpAddressRangesListPageLimit) // a full page => keep paging
	page2 := makeRanges(3)
	total := len(page1) + len(page2)

	iamMock.On("ListAllowedIpAddressRangesWithResponse", mock.Anything, testOrgId, offsetIs(0)).
		Return(listResp(page1, total))
	iamMock.On("ListAllowedIpAddressRangesWithResponse", mock.Anything, testOrgId, offsetIs(allowedIpAddressRangesListPageLimit)).
		Return(listResp(page2, total))

	r := &allowedIpAddressRangesResource{iamClient: iamMock, organizationId: testOrgId}
	got, diags := r.listAllRanges(ctx)

	assert.False(t, diags.HasError())
	assert.Len(t, got, total)
	iamMock.AssertNumberOfCalls(t, "ListAllowedIpAddressRangesWithResponse", 2)
}

func TestUnit_listAllRanges_singlePageStops(t *testing.T) {
	ctx := context.Background()
	iamMock := new(mocks_iam.ClientWithResponsesInterface)

	ranges := makeRanges(2)
	iamMock.On("ListAllowedIpAddressRangesWithResponse", mock.Anything, testOrgId, mock.Anything).
		Return(listResp(ranges, len(ranges)))

	r := &allowedIpAddressRangesResource{iamClient: iamMock, organizationId: testOrgId}
	got, diags := r.listAllRanges(ctx)

	assert.False(t, diags.HasError())
	assert.Len(t, got, 2)
	iamMock.AssertNumberOfCalls(t, "ListAllowedIpAddressRangesWithResponse", 1)
}

// A full first page whose length equals TotalCount must stop without a second call - otherwise the
// resource would page forever against a server that returns exactly a page's worth.
func TestUnit_listAllRanges_totalCountEarlyStop(t *testing.T) {
	ctx := context.Background()
	iamMock := new(mocks_iam.ClientWithResponsesInterface)

	page := makeRanges(allowedIpAddressRangesListPageLimit)
	iamMock.On("ListAllowedIpAddressRangesWithResponse", mock.Anything, testOrgId, mock.Anything).
		Return(listResp(page, len(page)))

	r := &allowedIpAddressRangesResource{iamClient: iamMock, organizationId: testOrgId}
	got, diags := r.listAllRanges(ctx)

	assert.False(t, diags.HasError())
	assert.Len(t, got, allowedIpAddressRangesListPageLimit)
	iamMock.AssertNumberOfCalls(t, "ListAllowedIpAddressRangesWithResponse", 1)
}

// An OK response with no parseable body (nil JSON200) must error rather than silently return an
// empty list, which would wipe the authoritative set from state.
func TestUnit_listAllRanges_okButNilBodyErrors(t *testing.T) {
	ctx := context.Background()
	iamMock := new(mocks_iam.ClientWithResponsesInterface)

	iamMock.On("ListAllowedIpAddressRangesWithResponse", mock.Anything, testOrgId, mock.Anything).
		Return(&iam.ListAllowedIpAddressRangesResponse{HTTPResponse: &http.Response{StatusCode: http.StatusOK}}, nil)

	r := &allowedIpAddressRangesResource{iamClient: iamMock, organizationId: testOrgId}
	_, diags := r.listAllRanges(ctx)

	assert.True(t, diags.HasError())
}

func TestUnit_bulkCreate_chunksByLimit(t *testing.T) {
	ctx := context.Background()
	labsMock := new(mocks_labs.ClientWithResponsesInterface)

	var chunkSizes []int
	labsMock.On("LabsCreateAllowedIpAddressRangesWithResponse", mock.Anything, testOrgId, mock.Anything).
		Run(func(args mock.Arguments) {
			body := args.Get(2).(labs.LabsCreateAllowedIpAddressRangesJSONRequestBody)
			chunkSizes = append(chunkSizes, len(body.AllowedIpAddressRanges))
		}).
		Return(okCreateResp())

	cidrs := make([]string, allowedIpAddressRangesBulkLimit+500)
	for i := range cidrs {
		cidrs[i] = fmt.Sprintf("10.%d.%d.0/24", i/256, i%256)
	}

	r := &allowedIpAddressRangesResource{labsClient: labsMock, organizationId: testOrgId}
	diags := r.bulkCreate(ctx, cidrs)

	assert.False(t, diags.HasError())
	assert.Equal(t, []int{allowedIpAddressRangesBulkLimit, 500}, chunkSizes)
}

func TestUnit_bulkDelete_chunksByLimit(t *testing.T) {
	ctx := context.Background()
	labsMock := new(mocks_labs.ClientWithResponsesInterface)

	var chunkSizes []int
	labsMock.On("LabsDeleteAllowedIpAddressRangesWithResponse", mock.Anything, testOrgId, mock.Anything).
		Run(func(args mock.Arguments) {
			body := args.Get(2).(labs.LabsDeleteAllowedIpAddressRangesJSONRequestBody)
			chunkSizes = append(chunkSizes, len(body.AllowedIpAddressRangeIds))
		}).
		Return(okDeleteResp())

	ids := make([]string, allowedIpAddressRangesBulkLimit+1)
	for i := range ids {
		ids[i] = fmt.Sprintf("id-%d", i)
	}

	r := &allowedIpAddressRangesResource{labsClient: labsMock, organizationId: testOrgId}
	diags := r.bulkDelete(ctx, ids)

	assert.False(t, diags.HasError())
	assert.Equal(t, []int{allowedIpAddressRangesBulkLimit, 1}, chunkSizes)
}

// A CIDR tracked in state but absent from the live list is skipped (not resolvable to an ID) and
// surfaced as a warning rather than silently dropped.
func TestUnit_idsForCidrs_warnsOnUnresolved(t *testing.T) {
	ctx := context.Background()
	iamMock := new(mocks_iam.ClientWithResponsesInterface)

	live := []iam.AllowedIpAddressRange{
		{Id: "id-a", IpAddressRange: "10.0.0.0/8"},
		{Id: "id-b", IpAddressRange: "192.168.0.0/16"},
	}
	iamMock.On("ListAllowedIpAddressRangesWithResponse", mock.Anything, testOrgId, mock.Anything).
		Return(listResp(live, len(live)))

	r := &allowedIpAddressRangesResource{iamClient: iamMock, organizationId: testOrgId}
	ids, diags := r.idsForCidrs(ctx, []string{"10.0.0.0/8", "172.16.0.0/12"})

	assert.False(t, diags.HasError())
	assert.Len(t, diags.Warnings(), 1)
	assert.Equal(t, []string{"id-a"}, ids)
}

// applyChanges must create the added ranges before deleting the removed ones so the caller keeps
// coverage throughout the update.
func TestUnit_applyChanges_createsBeforeDeletes(t *testing.T) {
	ctx := context.Background()
	labsMock := new(mocks_labs.ClientWithResponsesInterface)
	iamMock := new(mocks_iam.ClientWithResponsesInterface)

	iamMock.On("ListAllowedIpAddressRangesWithResponse", mock.Anything, testOrgId, mock.Anything).
		Return(listResp([]iam.AllowedIpAddressRange{{Id: "id-old", IpAddressRange: "192.168.0.0/16"}}, 1))

	// Record the order in which the write calls land, then assert create precedes delete.
	var callOrder []string
	labsMock.On("LabsCreateAllowedIpAddressRangesWithResponse", mock.Anything, testOrgId, mock.Anything).
		Run(func(mock.Arguments) { callOrder = append(callOrder, "create") }).
		Return(okCreateResp())
	labsMock.On("LabsDeleteAllowedIpAddressRangesWithResponse", mock.Anything, testOrgId, mock.Anything).
		Run(func(mock.Arguments) { callOrder = append(callOrder, "delete") }).
		Return(okDeleteResp())

	r := &allowedIpAddressRangesResource{iamClient: iamMock, labsClient: labsMock, organizationId: testOrgId}
	diags := r.applyChanges(ctx, []string{"10.0.0.0/8"}, []string{"192.168.0.0/16"})

	assert.False(t, diags.HasError())
	assert.Equal(t, []string{"create", "delete"}, callOrder)
}

func TestUnit_applyChanges_createOnlySkipsDelete(t *testing.T) {
	ctx := context.Background()
	labsMock := new(mocks_labs.ClientWithResponsesInterface)

	labsMock.On("LabsCreateAllowedIpAddressRangesWithResponse", mock.Anything, testOrgId, mock.Anything).
		Return(okCreateResp())

	r := &allowedIpAddressRangesResource{labsClient: labsMock, organizationId: testOrgId}
	diags := r.applyChanges(ctx, []string{"10.0.0.0/8"}, nil)

	assert.False(t, diags.HasError())
	labsMock.AssertNumberOfCalls(t, "LabsCreateAllowedIpAddressRangesWithResponse", 1)
	labsMock.AssertNotCalled(t, "LabsDeleteAllowedIpAddressRangesWithResponse")
}

func TestUnit_applyChanges_deleteOnlySkipsCreate(t *testing.T) {
	ctx := context.Background()
	labsMock := new(mocks_labs.ClientWithResponsesInterface)
	iamMock := new(mocks_iam.ClientWithResponsesInterface)

	iamMock.On("ListAllowedIpAddressRangesWithResponse", mock.Anything, testOrgId, mock.Anything).
		Return(listResp([]iam.AllowedIpAddressRange{{Id: "id-old", IpAddressRange: "192.168.0.0/16"}}, 1))
	labsMock.On("LabsDeleteAllowedIpAddressRangesWithResponse", mock.Anything, testOrgId, mock.Anything).
		Return(okDeleteResp())

	r := &allowedIpAddressRangesResource{iamClient: iamMock, labsClient: labsMock, organizationId: testOrgId}
	diags := r.applyChanges(ctx, nil, []string{"192.168.0.0/16"})

	assert.False(t, diags.HasError())
	labsMock.AssertNotCalled(t, "LabsCreateAllowedIpAddressRangesWithResponse")
	labsMock.AssertNumberOfCalls(t, "LabsDeleteAllowedIpAddressRangesWithResponse", 1)
}
