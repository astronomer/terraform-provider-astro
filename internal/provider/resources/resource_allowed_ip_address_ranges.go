package resources

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/clients/labs"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Per-request limit enforced by the Core bulk allowed-ip-address-ranges create/delete endpoints.
// The resource auto-chunks larger configurations across multiple requests.
const (
	allowedIpAddressRangesBulkLimit = 1000
	// allowedIpAddressRangesListPageLimit bounds how many ranges we request per list call. The
	// resource pages through the full org list since it authoritatively owns it.
	allowedIpAddressRangesListPageLimit = 1000
)

var (
	_ resource.Resource                = &allowedIpAddressRangesResource{}
	_ resource.ResourceWithConfigure   = &allowedIpAddressRangesResource{}
	_ resource.ResourceWithImportState = &allowedIpAddressRangesResource{}
)

func NewAllowedIpAddressRangesResource() resource.Resource {
	return &allowedIpAddressRangesResource{}
}

// allowedIpAddressRangesResource authoritatively manages an organization's IP access list as a
// single resource. Ranges not present in ip_address_ranges are removed on apply.
//
// Writes go through the labs bulk create/delete endpoints, while reads list through the iam
// v1beta1 endpoint (labs has no list endpoint for this resource).
type allowedIpAddressRangesResource struct {
	iamClient      *iam.ClientWithResponses
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
			"~> **Warning: risk of lockout.** While the access list is empty, access is unrestricted. Adding the " +
			"first range turns enforcement on, and the API will reject that first apply with a `400` unless the " +
			"submitted ranges include the public IP address of the machine running Terraform (for example, a CI " +
			"runner's egress IP). Once enforcement is on, the API does **not** protect you from removing the range " +
			"that covers your own IP: doing so - or narrowing/replacing it so your IP is no longer covered - locks " +
			"you (and this provider) out and will fail the apply mid-way. Always keep a range covering the machine " +
			"that runs Terraform.\n\n" +
			"~> **Note** Do not manage the IP access list with more than one `astro_allowed_ip_address_ranges` " +
			"resource. To adopt an access list that already exists, import it first (see below) rather than " +
			"re-declaring it, otherwise the first apply will conflict with the existing ranges.",
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
	r.iamClient = apiClients.IamClient
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

	// Store the planned ranges verbatim rather than re-listing from the API. ip_address_ranges is a
	// Required attribute, so Terraform enforces that the value saved to state equals the planned
	// (config) value; writing back an API re-list that differs by even a canonicalized CIDR (host
	// bits set, IPv6 form) would raise "Provider produced inconsistent result after apply" - the same
	// class as GH #244/#314. Read reconciles any server-side drift on the next refresh.
	data.Id = types.StringValue(r.organizationId)
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
	data.Id = types.StringValue(r.organizationId)
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

	// Create before delete so the machine running Terraform never loses coverage mid-update. IP
	// enforcement is per-request and the API only guards against lockout on the empty -> first-entry
	// transition (not on deletes), so deleting the caller's covering range before adding its
	// replacement would lock the caller out and fail the remaining requests. Adding ranges to an
	// already non-empty list is always safe.
	if len(toCreate) > 0 {
		resp.Diagnostics.Append(r.bulkCreate(ctx, toCreate)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

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

	// Store the planned ranges verbatim (see Create) so state equals the Required config value and
	// avoids an inconsistent-result error; Read reconciles server-side drift on the next refresh.
	plan.Id = types.StringValue(r.organizationId)
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

// ImportState imports the organization's existing IP access list. The resource is a singleton for
// the organization configured on the provider, so the import ID is only cosmetic (use the
// organization ID) - the subsequent Read populates ip_address_ranges from the API and sets id to
// the organization ID regardless of the value passed here.
func (r *allowedIpAddressRangesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// bulkCreate chunks the given CIDRs by the API's per-request limit and creates them via the labs
// bulk create endpoint.
func (r *allowedIpAddressRangesResource) bulkCreate(ctx context.Context, cidrs []string) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, chunk := range chunkSlice(cidrs, allowedIpAddressRangesBulkLimit) {
		createResp, err := r.labsClient.LabsCreateAllowedIpAddressRangesWithResponse(ctx, r.organizationId, labs.BulkCreateAllowedIpAddressRangesRequest{AllowedIpAddressRanges: chunk})
		if err != nil {
			tflog.Error(ctx, "failed to bulk create allowed IP address ranges", map[string]interface{}{"error": err})
			diags.AddError("Client Error", fmt.Sprintf("Unable to bulk create allowed IP address ranges: %s", err))
			return diags
		}
		// The bulk create is atomic, so a 409 means one or more CIDRs in the chunk already exist.
		// This is reachable when retrying after a multi-chunk create where an earlier chunk committed
		// but a later one failed: the retry re-submits the committed chunk and conflicts. Treat 409 as
		// success only when every CIDR in the chunk is in fact already present (the desired state is
		// met, so the create is idempotent); otherwise fall through and surface the conflict.
		if createResp.StatusCode() == http.StatusConflict {
			existing, d := r.listAllRanges(ctx)
			if d.HasError() {
				diags.Append(d...)
				return diags
			}
			present := make(map[string]bool, len(existing))
			for _, rng := range existing {
				present[rng.IpAddressRange] = true
			}
			allPresent := true
			for _, c := range chunk {
				if !present[c] {
					allPresent = false
					break
				}
			}
			if allPresent {
				tflog.Debug(ctx, "bulk create returned 409 but all ranges are already present; treating as idempotent success", map[string]interface{}{"count": len(chunk)})
				continue
			}
		}
		if _, d := clients.NormalizeAPIError(ctx, createResp.HTTPResponse, createResp.Body); d != nil {
			diags.Append(d)
			return diags
		}
	}
	return diags
}

// bulkDelete chunks the given range IDs by the API's per-request limit and deletes them via the
// labs bulk delete endpoint.
func (r *allowedIpAddressRangesResource) bulkDelete(ctx context.Context, ids []string) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, chunk := range chunkSlice(ids, allowedIpAddressRangesBulkLimit) {
		deleteResp, err := r.labsClient.LabsDeleteAllowedIpAddressRangesWithResponse(ctx, r.organizationId, labs.BulkDeleteAllowedIpAddressRangesRequest{AllowedIpAddressRangeIds: chunk})
		if err != nil {
			tflog.Error(ctx, "failed to bulk delete allowed IP address ranges", map[string]interface{}{"error": err})
			diags.AddError("Client Error", fmt.Sprintf("Unable to bulk delete allowed IP address ranges: %s", err))
			return diags
		}
		if _, d := clients.NormalizeAPIError(ctx, deleteResp.HTTPResponse, deleteResp.Body); d != nil {
			diags.Append(d)
			return diags
		}
	}
	return diags
}

// listAll pages through the organization's full allowed IP address range list (via iam v1beta1)
// and returns the CIDRs.
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
	var missing []string
	for _, c := range cidrs {
		if id, ok := byCidr[c]; ok {
			ids = append(ids, id)
		} else {
			missing = append(missing, c)
		}
	}
	// An authoritative delete that can't resolve a range to an ID would silently under-delete, leaving
	// a range in place on a security-sensitive allowlist with no signal. Warn so the drift is visible.
	if len(missing) > 0 {
		diags.AddWarning(
			"Some allowed IP address ranges could not be resolved for deletion",
			fmt.Sprintf("The following ranges were expected in the organization's current list but were not found, so they were not deleted: %s", strings.Join(missing, ", ")),
		)
	}
	return ids, diags
}

func (r *allowedIpAddressRangesResource) listAllRanges(ctx context.Context) ([]iam.AllowedIpAddressRange, diag.Diagnostics) {
	var diags diag.Diagnostics
	var all []iam.AllowedIpAddressRange
	limit := allowedIpAddressRangesListPageLimit
	offset := 0
	for {
		params := &iam.ListAllowedIpAddressRangesParams{Limit: &limit, Offset: &offset}
		listResp, err := r.iamClient.ListAllowedIpAddressRangesWithResponse(ctx, r.organizationId, params)
		if err != nil {
			tflog.Error(ctx, "failed to list allowed IP address ranges", map[string]interface{}{"error": err})
			diags.AddError("Client Error", fmt.Sprintf("Unable to list allowed IP address ranges: %s", err))
			return all, diags
		}
		// A missing body must not read as "the list is empty" - for this authoritative resource
		// that would wipe every managed range from state and plan their deletion.
		if _, d := clients.NormalizeAPIResponseWithBody(ctx, listResp.HTTPResponse, listResp.Body, listResp.JSON200, "list allowed IP address ranges"); d != nil {
			diags.Append(d)
			return all, diags
		}
		all = append(all, listResp.JSON200.AllowedIpAddressRanges...)
		offset += len(listResp.JSON200.AllowedIpAddressRanges)
		if len(listResp.JSON200.AllowedIpAddressRanges) < limit || offset >= listResp.JSON200.TotalCount {
			break
		}
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
