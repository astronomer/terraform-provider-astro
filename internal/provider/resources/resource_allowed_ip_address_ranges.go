package resources

import (
	"context"
	"fmt"
	"net/http"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/labs"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Per-request limit enforced by the Core bulk allowed-ip-address-ranges create endpoint. The
// resource auto-chunks larger configurations across multiple requests. The API has no documented
// limit on the delete endpoint, so deletes are chunked at the same size defensively.
const (
	allowedIpAddressRangesBulkLimit = 1000
	// allowedIpAddressRangesListPageLimit bounds how many ranges we request per list call. The
	// resource pages through the full org list since it authoritatively owns it.
	allowedIpAddressRangesListPageLimit = 1000
)

var (
	_ resource.Resource              = &allowedIpAddressRangesResource{}
	_ resource.ResourceWithConfigure = &allowedIpAddressRangesResource{}
)

func NewAllowedIpAddressRangesResource() resource.Resource {
	return &allowedIpAddressRangesResource{}
}

// allowedIpAddressRangesResource authoritatively manages an organization's IP access list as a
// single resource, via the labs bulk create/delete endpoints. Ranges not present in
// ip_address_ranges are removed on apply.
type allowedIpAddressRangesResource struct {
	labsClient     *labs.ClientWithResponses
	organizationId string
}

func (r *allowedIpAddressRangesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_allowed_ip_address_ranges"
}

func (r *allowedIpAddressRangesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage an organization's IP access list as a single resource. This resource is " +
			"authoritative: any allowed IP address ranges not present in `ip_address_ranges` are removed on apply.\n\n" +
			"~> **Note** Do not manage the IP access list with more than one `astro_allowed_ip_address_ranges` " +
			"resource, and be careful not to remove the range that includes the machine applying the Terraform " +
			"configuration - the API rejects changes that would lock out the current caller when the access list " +
			"is non-empty.",
		Attributes: schemas.AllowedIpAddressRangesResourceSchemaAttributes(),
	}
}

func (r *allowedIpAddressRangesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiClients, ok := req.ProviderData.(models.ApiClientsModel)
	if !ok {
		utils.ResourceApiClientConfigureError(ctx, req, resp)
		return
	}
	r.labsClient = apiClients.LabsClient
	r.organizationId = apiClients.OrganizationId
}

func (r *allowedIpAddressRangesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.AllowedIpAddressRangesResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cidrs, diags := utils.TypesSetToStringSlice(ctx, data.IpAddressRanges)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.bulkCreate(ctx, cidrs)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, diags := r.listAll(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	setVal, d := utils.StringSet(&result)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.IpAddressRanges = setVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *allowedIpAddressRangesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.AllowedIpAddressRangesResource
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, diags := r.listAll(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	setVal, d := utils.StringSet(&result)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.IpAddressRanges = setVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *allowedIpAddressRangesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state models.AllowedIpAddressRangesResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planCidrs, diags := utils.TypesSetToStringSlice(ctx, plan.IpAddressRanges)
	resp.Diagnostics.Append(diags...)
	stateCidrs, diags := utils.TypesSetToStringSlice(ctx, state.IpAddressRanges)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	toCreate, toDelete := diffCidrs(planCidrs, stateCidrs)

	if len(toDelete) > 0 {
		ids, diags := r.idsForCidrs(ctx, toDelete)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(r.bulkDelete(ctx, ids)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if len(toCreate) > 0 {
		resp.Diagnostics.Append(r.bulkCreate(ctx, toCreate)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	result, diags := r.listAll(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	setVal, d := utils.StringSet(&result)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.IpAddressRanges = setVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *allowedIpAddressRangesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.AllowedIpAddressRangesResource
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cidrs, diags := utils.TypesSetToStringSlice(ctx, data.IpAddressRanges)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(cidrs) == 0 {
		return
	}

	ids, diags := r.idsForCidrs(ctx, cidrs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.bulkDelete(ctx, ids)...)
}

// bulkCreate chunks the given CIDRs by the API's per-request limit and creates them.
func (r *allowedIpAddressRangesResource) bulkCreate(ctx context.Context, cidrs []string) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, chunk := range chunkSlice(cidrs, allowedIpAddressRangesBulkLimit) {
		createResp, err := r.labsClient.LabsCreateAllowedIpAddressRangesWithResponse(ctx, r.organizationId, labs.CreateAllowedIpAddressRangesRequest{AllowedIpAddressRanges: chunk})
		if err != nil {
			tflog.Error(ctx, "failed to bulk create allowed IP address ranges", map[string]interface{}{"error": err})
			diags.AddError("Client Error", fmt.Sprintf("Unable to bulk create allowed IP address ranges: %s", err))
			return diags
		}
		if _, d := clients.NormalizeAPIError(ctx, createResp.HTTPResponse, createResp.Body); d != nil {
			diags.Append(d)
			return diags
		}
	}
	return diags
}

// bulkDelete chunks the given range IDs by the API's per-request limit and deletes them.
func (r *allowedIpAddressRangesResource) bulkDelete(ctx context.Context, ids []string) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, chunk := range chunkSlice(ids, allowedIpAddressRangesBulkLimit) {
		deleteResp, err := r.labsClient.LabsDeleteAllowedIpAddressRangesWithResponse(ctx, r.organizationId, labs.DeleteAllowedIpAddressRangesRequest{AllowedIpAddressRangeIds: chunk})
		if err != nil {
			tflog.Error(ctx, "failed to bulk delete allowed IP address ranges", map[string]interface{}{"error": err})
			diags.AddError("Client Error", fmt.Sprintf("Unable to bulk delete allowed IP address ranges: %s", err))
			return diags
		}
		statusCode, d := clients.NormalizeAPIError(ctx, deleteResp.HTTPResponse, deleteResp.Body)
		if statusCode != http.StatusNotFound && d != nil {
			diags.Append(d)
			return diags
		}
	}
	return diags
}

// listAll pages through the organization's full allowed IP address range list and returns the CIDRs.
func (r *allowedIpAddressRangesResource) listAll(ctx context.Context) ([]string, diag.Diagnostics) {
	ranges, diags := r.listAllRanges(ctx)
	if diags.HasError() {
		return nil, diags
	}
	cidrs := make([]string, 0, len(ranges))
	for _, rng := range ranges {
		cidrs = append(cidrs, rng.IpAddressRange)
	}
	return cidrs, diags
}

// idsForCidrs looks up the range IDs for the given CIDRs via the list endpoint.
func (r *allowedIpAddressRangesResource) idsForCidrs(ctx context.Context, cidrs []string) ([]string, diag.Diagnostics) {
	ranges, diags := r.listAllRanges(ctx)
	if diags.HasError() {
		return nil, diags
	}
	byCidr := make(map[string]string, len(ranges))
	for _, rng := range ranges {
		byCidr[rng.IpAddressRange] = rng.Id
	}
	ids := make([]string, 0, len(cidrs))
	for _, c := range cidrs {
		if id, ok := byCidr[c]; ok {
			ids = append(ids, id)
		}
	}
	return ids, diags
}

func (r *allowedIpAddressRangesResource) listAllRanges(ctx context.Context) ([]labs.AllowedIpAddressRange, diag.Diagnostics) {
	var diags diag.Diagnostics
	var all []labs.AllowedIpAddressRange
	limit := allowedIpAddressRangesListPageLimit
	offset := 0
	for {
		params := &labs.ListAllowedIpAddressRangesParams{Limit: &limit, Offset: &offset}
		listResp, err := r.labsClient.LabsListAllowedIpAddressRangesWithResponse(ctx, r.organizationId, params)
		if err != nil {
			tflog.Error(ctx, "failed to list allowed IP address ranges", map[string]interface{}{"error": err})
			diags.AddError("Client Error", fmt.Sprintf("Unable to list allowed IP address ranges: %s", err))
			return all, diags
		}
		if _, d := clients.NormalizeAPIError(ctx, listResp.HTTPResponse, listResp.Body); d != nil {
			diags.Append(d)
			return all, diags
		}
		if listResp.JSON200 == nil {
			break
		}
		all = append(all, listResp.JSON200.AllowedIpAddressRanges...)
		if len(listResp.JSON200.AllowedIpAddressRanges) < limit {
			break
		}
		offset += limit
	}
	return all, diags
}

// diffCidrs partitions plan/state CIDR sets into the ranges to create and the ranges to delete.
func diffCidrs(planCidrs, stateCidrs []string) (toCreate, toDelete []string) {
	stateSet := make(map[string]bool, len(stateCidrs))
	for _, c := range stateCidrs {
		stateSet[c] = true
	}
	planSet := make(map[string]bool, len(planCidrs))
	for _, c := range planCidrs {
		planSet[c] = true
	}

	for _, c := range planCidrs {
		if !stateSet[c] {
			toCreate = append(toCreate, c)
		}
	}
	for _, c := range stateCidrs {
		if !planSet[c] {
			toDelete = append(toDelete, c)
		}
	}
	return toCreate, toDelete
}
