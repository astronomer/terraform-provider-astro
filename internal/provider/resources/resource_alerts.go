package resources

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/labs"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Per-request limits enforced by the Core bulk alert endpoints. The resource auto-chunks larger
// configurations across multiple requests.
const (
	alertsBulkCreateLimit = 30
	alertsBulkUpdateLimit = 30
	alertsBulkDeleteLimit = 20
	// alertsListPageLimit bounds how many alert IDs we request per ListAlerts call during Read.
	alertsListPageLimit = 100
)

var (
	_ resource.Resource              = &alertsResource{}
	_ resource.ResourceWithConfigure = &alertsResource{}
)

func NewAlertsResource() resource.Resource {
	return &alertsResource{}
}

// alertsResource manages a collection of alerts as a single resource, batching the underlying
// create/update/delete calls.
//
// Writes go through the labs API (which exposes the bulk create/update/delete endpoints), while
// reads go through the platform v1beta1 list endpoint (labs has no list endpoint). The labs write
// responses are used only to capture server-assigned ids; full state is always refreshed from the
// platform list so a single, well-tested response mapper is reused.
type alertsResource struct {
	labsClient     *labs.ClientWithResponses
	organizationId string
}

func (r *alertsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alerts"
}

func (r *alertsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage a collection of alerts as a single resource. The resource batches " +
			"create/update/delete calls and automatically chunks requests that exceed the API's per-request " +
			"limits (30 for create/update, 20 for delete).\n\n" +
			"~> **Note** Do not manage the same alert with both `astro_alert` and `astro_alerts`. Each resource " +
			"claims ownership of the alerts it manages, so overlapping definitions conflict and cause churn on " +
			"every apply. Use one resource or the other for a given alert.",
		Attributes: schemas.AlertsResourceSchemaAttributes(),
	}
}

func (r *alertsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *alertsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.AlertsResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planElems := make(map[string]models.AlertsResourceElementModel)
	resp.Diagnostics.Append(data.Alerts.ElementsAs(ctx, &planElems, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keys := sortedKeys(planElems)
	createReqs := make([]labs.CreateAlertRequest, 0, len(keys))
	for _, k := range keys {
		cr, diags := BuildLabsCreateAlertRequest(ctx, planElems[k].ToAlertResource())
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReqs = append(createReqs, cr)
	}

	createdIds, diags := r.bulkCreate(ctx, createReqs)
	// Map the alerts that were created (in request order) back to their keys, even on partial
	// failure, so Terraform records what exists.
	keyToId := make(map[string]string, len(createdIds))
	for i := range createdIds {
		keyToId[keys[i]] = createdIds[i]
	}

	result, refreshDiags := r.refreshState(ctx, keyToId)
	resp.Diagnostics.Append(refreshDiags...)

	mapVal, d := types.MapValueFrom(ctx, models.AlertsElementObjectType(), result)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Alerts = mapVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	// Surface any bulk error after persisting partial state.
	resp.Diagnostics.Append(diags...)
}

func (r *alertsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.AlertsResource
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateElems := make(map[string]models.AlertsResourceElementModel)
	resp.Diagnostics.Append(data.Alerts.ElementsAs(ctx, &stateElems, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyToId := make(map[string]string, len(stateElems))
	for k, e := range stateElems {
		if id := e.Id.ValueString(); id != "" {
			keyToId[k] = id
		}
	}
	if len(keyToId) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	result, diags := r.refreshState(ctx, keyToId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(result) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	mapVal, d := types.MapValueFrom(ctx, models.AlertsElementObjectType(), result)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Alerts = mapVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *alertsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state models.AlertsResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planElems := make(map[string]models.AlertsResourceElementModel)
	stateElems := make(map[string]models.AlertsResourceElementModel)
	resp.Diagnostics.Append(plan.Alerts.ElementsAs(ctx, &planElems, false)...)
	resp.Diagnostics.Append(state.Alerts.ElementsAs(ctx, &stateElems, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Partition keys into create / update / delete sets.
	var createKeys, updateKeys []string
	for _, k := range sortedKeys(planElems) {
		if _, ok := stateElems[k]; ok {
			updateKeys = append(updateKeys, k)
		} else {
			createKeys = append(createKeys, k)
		}
	}
	var deleteIds []string
	for k, se := range stateElems {
		if _, ok := planElems[k]; !ok {
			if id := se.Id.ValueString(); id != "" {
				deleteIds = append(deleteIds, id)
			}
		}
	}

	// keyToId tracks the server-assigned id for every key that should survive this update, so we can
	// refresh authoritative state from the platform list afterwards (even on partial failure).
	keyToId := make(map[string]string, len(planElems))
	for _, k := range updateKeys {
		if id := stateElems[k].Id.ValueString(); id != "" {
			keyToId[k] = id
		}
	}

	var writeDiags diag.Diagnostics

	// Deletes first to free capacity, then creates, then updates.
	if len(deleteIds) > 0 {
		if d := r.bulkDelete(ctx, deleteIds); d.HasError() {
			r.persistFromIds(ctx, &resp.State, &resp.Diagnostics, keyToId)
			resp.Diagnostics.Append(d...)
			return
		}
	}

	if len(createKeys) > 0 {
		createReqs := make([]labs.CreateAlertRequest, 0, len(createKeys))
		for _, k := range createKeys {
			cr, diags := BuildLabsCreateAlertRequest(ctx, planElems[k].ToAlertResource())
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			createReqs = append(createReqs, cr)
		}
		createdIds, diags := r.bulkCreate(ctx, createReqs)
		for i := range createdIds {
			keyToId[createKeys[i]] = createdIds[i]
		}
		if diags.HasError() {
			r.persistFromIds(ctx, &resp.State, &resp.Diagnostics, keyToId)
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	if len(updateKeys) > 0 {
		updateReqs := make([]labs.UpdateAlertRequest, 0, len(updateKeys))
		for _, k := range updateKeys {
			elem := planElems[k]
			elem.Id = stateElems[k].Id // carry the server-assigned id required by bulk update
			ur, diags := BuildLabsUpdateAlertRequest(ctx, elem.ToAlertResource())
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			updateReqs = append(updateReqs, ur)
		}
		if diags := r.bulkUpdate(ctx, updateReqs); diags.HasError() {
			writeDiags.Append(diags...)
		}
	}

	r.persistFromIds(ctx, &resp.State, &resp.Diagnostics, keyToId)
	resp.Diagnostics.Append(writeDiags...)
}

func (r *alertsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.AlertsResource
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateElems := make(map[string]models.AlertsResourceElementModel)
	resp.Diagnostics.Append(data.Alerts.ElementsAs(ctx, &stateElems, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ids := make([]string, 0, len(stateElems))
	for _, e := range stateElems {
		if id := e.Id.ValueString(); id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return
	}
	resp.Diagnostics.Append(r.bulkDelete(ctx, ids)...)
}

// refreshState fetches the current server state for the alerts identified by keyToId and maps each
// back to its key. Alerts that no longer exist server-side are dropped from the result.
func (r *alertsResource) refreshState(ctx context.Context, keyToId map[string]string) (map[string]models.AlertsResourceElementModel, diag.Diagnostics) {
	ids := make([]string, 0, len(keyToId))
	for _, id := range keyToId {
		ids = append(ids, id)
	}
	result := make(map[string]models.AlertsResourceElementModel, len(keyToId))
	if len(ids) == 0 {
		return result, nil
	}

	found, diags := r.listByIds(ctx, ids)
	if diags.HasError() {
		return result, diags
	}
	for k, id := range keyToId {
		alert, ok := found[id]
		if !ok {
			// Alert deleted outside Terraform (or never created); drop it from state.
			continue
		}
		var elem models.AlertsResourceElementModel
		if d := elem.ReadFromResponse(ctx, &alert); d.HasError() {
			diags.Append(d...)
			return result, diags
		}
		result[k] = elem
	}
	return result, diags
}

// persistFromIds refreshes state for keyToId and writes the result back to Terraform state. Used on
// the update path (including partial-failure exits) so state reflects what exists server-side.
func (r *alertsResource) persistFromIds(ctx context.Context, state *tfsdk.State, diags *diag.Diagnostics, keyToId map[string]string) {
	result, d := r.refreshState(ctx, keyToId)
	diags.Append(d...)
	mapVal, md := types.MapValueFrom(ctx, models.AlertsElementObjectType(), result)
	diags.Append(md...)
	if md.HasError() {
		return
	}
	diags.Append(state.Set(ctx, &models.AlertsResource{Alerts: mapVal})...)
}

// bulkCreate chunks create requests by the API limit and returns the created alert IDs in request
// order. On a chunk failure it returns the IDs created so far plus the error.
func (r *alertsResource) bulkCreate(ctx context.Context, reqs []labs.CreateAlertRequest) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createdIds []string
	for _, chunk := range chunkSlice(reqs, alertsBulkCreateLimit) {
		alertResp, err := r.labsClient.LabsCreateAlertsWithResponse(ctx, r.organizationId, labs.CreateAlertsRequest{Alerts: chunk})
		if err != nil {
			tflog.Error(ctx, "failed to bulk create alerts", map[string]interface{}{"error": err})
			diags.AddError("Client Error", fmt.Sprintf("Unable to bulk create alerts: %s", err))
			return createdIds, diags
		}
		if _, d := clients.NormalizeAPIError(ctx, alertResp.HTTPResponse, alertResp.Body); d != nil {
			diags.Append(d)
			return createdIds, diags
		}
		if alertResp.JSON200 != nil {
			for _, a := range alertResp.JSON200.Alerts {
				createdIds = append(createdIds, a.Id)
			}
		}
	}
	return createdIds, diags
}

// bulkUpdate chunks update requests by the API limit and applies them.
func (r *alertsResource) bulkUpdate(ctx context.Context, reqs []labs.UpdateAlertRequest) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, chunk := range chunkSlice(reqs, alertsBulkUpdateLimit) {
		alertResp, err := r.labsClient.LabsUpdateAlertsWithResponse(ctx, r.organizationId, labs.UpdateAlertsRequest{Alerts: chunk})
		if err != nil {
			tflog.Error(ctx, "failed to bulk update alerts", map[string]interface{}{"error": err})
			diags.AddError("Client Error", fmt.Sprintf("Unable to bulk update alerts: %s", err))
			return diags
		}
		if _, d := clients.NormalizeAPIError(ctx, alertResp.HTTPResponse, alertResp.Body); d != nil {
			diags.Append(d)
			return diags
		}
	}
	return diags
}

// bulkDelete chunks alert IDs by the API limit and deletes them.
func (r *alertsResource) bulkDelete(ctx context.Context, ids []string) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, chunk := range chunkSlice(ids, alertsBulkDeleteLimit) {
		alertResp, err := r.labsClient.LabsDeleteAlertsWithResponse(ctx, r.organizationId, labs.DeleteAlertsRequest{AlertIds: chunk})
		if err != nil {
			tflog.Error(ctx, "failed to bulk delete alerts", map[string]interface{}{"error": err})
			diags.AddError("Client Error", fmt.Sprintf("Unable to bulk delete alerts: %s", err))
			return diags
		}
		statusCode, d := clients.NormalizeAPIError(ctx, alertResp.HTTPResponse, alertResp.Body)
		if statusCode != http.StatusNotFound && d != nil {
			diags.Append(d)
			return diags
		}
	}
	return diags
}

// listByIds fetches alerts by ID (chunked) via the platform list endpoint and returns them keyed by
// alert ID. Labs has no list endpoint, so reads route through platform v1beta1.
func (r *alertsResource) listByIds(ctx context.Context, ids []string) (map[string]labs.Alert, diag.Diagnostics) {
	var diags diag.Diagnostics
	found := make(map[string]labs.Alert, len(ids))
	for _, chunk := range chunkSlice(ids, alertsListPageLimit) {
		alertIds := chunk
		limit := len(chunk)
		params := &labs.LabsListAlertsParams{
			AlertIds: &alertIds,
			Limit:    &limit,
		}
		listResp, err := r.labsClient.LabsListAlertsWithResponse(ctx, r.organizationId, params)
		if err != nil {
			tflog.Error(ctx, "failed to list alerts", map[string]interface{}{"error": err})
			diags.AddError("Client Error", fmt.Sprintf("Unable to list alerts: %s", err))
			return found, diags
		}
		if _, d := clients.NormalizeAPIError(ctx, listResp.HTTPResponse, listResp.Body); d != nil {
			diags.Append(d)
			return found, diags
		}
		if listResp.JSON200 != nil {
			for _, a := range listResp.JSON200.Alerts {
				found[a.Id] = a
			}
		}
	}
	return found, diags
}

// sortedKeys returns the keys of m in deterministic order so that positional mapping of bulk-create
// responses is stable.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// chunkSlice splits items into consecutive chunks of at most size elements.
func chunkSlice[T any](items []T, size int) [][]T {
	if size <= 0 {
		return [][]T{items}
	}
	var chunks [][]T
	for i := 0; i < len(items); i += size {
		end := i + size
		if end > len(items) {
			end = len(items)
		}
		chunks = append(chunks, items[i:end])
	}
	return chunks
}
