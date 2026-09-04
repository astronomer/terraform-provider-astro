package resources

import (
	"context"
	"fmt"
	"net/http"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/labs"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
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

// Per-request limits enforced by the Core bulk notification channel endpoints. The resource
// auto-chunks larger configurations across multiple requests.
const (
	notificationChannelsBulkCreateLimit = 30
	notificationChannelsBulkDeleteLimit = 30
	// notificationChannelsListPageLimit bounds how many channel IDs we request per
	// ListNotificationChannels call during Read.
	notificationChannelsListPageLimit = 100
)

var (
	_ resource.Resource              = &notificationChannelsResource{}
	_ resource.ResourceWithConfigure = &notificationChannelsResource{}
)

func NewNotificationChannelsResource() resource.Resource {
	return &notificationChannelsResource{}
}

// notificationChannelsResource manages a collection of notification channels as a single resource,
// batching the underlying create/delete calls.
//
// Creates and deletes go through the labs API bulk endpoints. Labs has no bulk-update endpoint for
// notification channels (unlike alerts), so updates go through the single-channel platform endpoint,
// one request per changed channel — tracked in CPP-968. Reads also go through the platform
// v1beta1 list endpoint (labs has no list endpoint); the labs create response is used only to
// capture server-assigned ids.
type notificationChannelsResource struct {
	labsClient     *labs.ClientWithResponses
	platformClient *platform.ClientWithResponses
	organizationId string
}

func (r *notificationChannelsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_channels"
}

func (r *notificationChannelsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage a collection of notification channels as a single resource. The resource batches " +
			"create/delete calls and automatically chunks requests that exceed the API's per-request limit (30). " +
			"Updates are applied one channel at a time via the single-channel endpoint.\n\n" +
			"~> **Note** Do not manage the same notification channel with both `astro_notification_channel` and " +
			"`astro_notification_channels`. Each resource claims ownership of the channels it manages, so overlapping " +
			"definitions conflict and cause churn on every apply. Use one resource or the other for a given channel.",
		Attributes: schemas.NotificationChannelsResourceSchemaAttributes(),
	}
}

func (r *notificationChannelsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiClients, ok := req.ProviderData.(models.ApiClientsModel)
	if !ok {
		utils.ResourceApiClientConfigureError(ctx, req, resp)
		return
	}
	r.labsClient = apiClients.LabsClient
	r.platformClient = apiClients.PlatformClient
	r.organizationId = apiClients.OrganizationId
}

func (r *notificationChannelsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.NotificationChannelsResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planElems := make(map[string]models.NotificationChannelsResourceElementModel)
	resp.Diagnostics.Append(data.NotificationChannels.ElementsAs(ctx, &planElems, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keys := sortedKeys(planElems)
	createReqs := make([]labs.CreateNotificationChannelRequest, 0, len(keys))
	for _, k := range keys {
		cr, diags := BuildLabsCreateNotificationChannelRequest(ctx, planElems[k])
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReqs = append(createReqs, cr)
	}

	createdIds, diags := r.bulkCreate(ctx, createReqs)
	// Map the channels that were created (in request order) back to their keys, even on partial
	// failure, so Terraform records what exists.
	keyToId := make(map[string]string, len(createdIds))
	for i := range createdIds {
		keyToId[keys[i]] = createdIds[i]
	}

	result, refreshDiags := r.refreshState(ctx, keyToId, planElems)
	resp.Diagnostics.Append(refreshDiags...)

	mapVal, d := types.MapValueFrom(ctx, models.NotificationChannelsElementObjectType(), result)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.NotificationChannels = mapVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	// Surface any bulk error after persisting partial state.
	resp.Diagnostics.Append(diags...)
}

func (r *notificationChannelsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.NotificationChannelsResource
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateElems := make(map[string]models.NotificationChannelsResourceElementModel)
	resp.Diagnostics.Append(data.NotificationChannels.ElementsAs(ctx, &stateElems, false)...)
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

	result, diags := r.refreshState(ctx, keyToId, stateElems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(result) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	mapVal, d := types.MapValueFrom(ctx, models.NotificationChannelsElementObjectType(), result)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.NotificationChannels = mapVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *notificationChannelsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state models.NotificationChannelsResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planElems := make(map[string]models.NotificationChannelsResourceElementModel)
	stateElems := make(map[string]models.NotificationChannelsResourceElementModel)
	resp.Diagnostics.Append(plan.NotificationChannels.ElementsAs(ctx, &planElems, false)...)
	resp.Diagnostics.Append(state.NotificationChannels.ElementsAs(ctx, &stateElems, false)...)
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
			r.persistFromIds(ctx, &resp.State, &resp.Diagnostics, keyToId, planElems)
			resp.Diagnostics.Append(d...)
			return
		}
	}

	if len(createKeys) > 0 {
		createReqs := make([]labs.CreateNotificationChannelRequest, 0, len(createKeys))
		for _, k := range createKeys {
			cr, diags := BuildLabsCreateNotificationChannelRequest(ctx, planElems[k])
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
			r.persistFromIds(ctx, &resp.State, &resp.Diagnostics, keyToId, planElems)
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	// desired holds the values whose sensitive fields (which the API never returns) should be
	// written to state per key. It starts from the plan; a channel whose update fails is reverted
	// to its prior state below so a failed update — including a sensitive-only change — is not
	// persisted as if it had succeeded.
	desired := make(map[string]models.NotificationChannelsResourceElementModel, len(planElems))
	for k, e := range planElems {
		desired[k] = e
	}

	// Labs has no bulk-update endpoint for notification channels; update each changed channel
	// through the single-channel platform endpoint (CPP-968). Channels whose configuration is
	// unchanged are skipped so an edit to one channel does not re-write the others.
	for _, k := range updateKeys {
		id := stateElems[k].Id.ValueString()
		if id == "" {
			continue
		}
		if notificationChannelElementUnchanged(planElems[k], stateElems[k]) {
			continue
		}
		if d := r.updateChannel(ctx, id, planElems[k]); d.HasError() {
			writeDiags.Append(d...)
			// The channel still holds its prior server-side values; keep them (including sensitive
			// fields) instead of persisting the failed new values.
			desired[k] = stateElems[k]
		}
	}

	r.persistFromIds(ctx, &resp.State, &resp.Diagnostics, keyToId, desired)
	resp.Diagnostics.Append(writeDiags...)
}

func (r *notificationChannelsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.NotificationChannelsResource
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateElems := make(map[string]models.NotificationChannelsResourceElementModel)
	resp.Diagnostics.Append(data.NotificationChannels.ElementsAs(ctx, &stateElems, false)...)
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

// notificationChannelElementUnchanged reports whether two elements carry the same configurable
// values, so an unchanged channel can be skipped during a bulk update.
func notificationChannelElementUnchanged(a, b models.NotificationChannelsResourceElementModel) bool {
	return a.Name.Equal(b.Name) &&
		a.Type.Equal(b.Type) &&
		a.EntityId.Equal(b.EntityId) &&
		a.EntityType.Equal(b.EntityType) &&
		a.IsShared.Equal(b.IsShared) &&
		a.Definition.Equal(b.Definition)
}

// refreshState fetches the current server state for the channels identified by keyToId and maps each
// back to its key. desired carries the configured/prior elements so sensitive fields the API never
// returns are preserved. Channels that no longer exist server-side are dropped from the result.
func (r *notificationChannelsResource) refreshState(
	ctx context.Context,
	keyToId map[string]string,
	desired map[string]models.NotificationChannelsResourceElementModel,
) (map[string]models.NotificationChannelsResourceElementModel, diag.Diagnostics) {
	ids := make([]string, 0, len(keyToId))
	for _, id := range keyToId {
		ids = append(ids, id)
	}
	result := make(map[string]models.NotificationChannelsResourceElementModel, len(keyToId))
	if len(ids) == 0 {
		return result, nil
	}

	found, diags := r.listByIds(ctx, ids)
	if diags.HasError() {
		return result, diags
	}
	for k, id := range keyToId {
		channel, ok := found[id]
		if !ok {
			// Channel deleted outside Terraform (or never created); drop it from state.
			continue
		}
		var elem models.NotificationChannelsResourceElementModel
		if d := elem.ReadFromResponse(ctx, &channel, desired[k].Definition); d.HasError() {
			diags.Append(d...)
			return result, diags
		}
		result[k] = elem
	}
	return result, diags
}

// persistFromIds refreshes state for keyToId and writes the result back to Terraform state. Used on
// the update path (including partial-failure exits) so state reflects what exists server-side.
func (r *notificationChannelsResource) persistFromIds(
	ctx context.Context,
	state *tfsdk.State,
	diags *diag.Diagnostics,
	keyToId map[string]string,
	desired map[string]models.NotificationChannelsResourceElementModel,
) {
	result, d := r.refreshState(ctx, keyToId, desired)
	diags.Append(d...)
	mapVal, md := types.MapValueFrom(ctx, models.NotificationChannelsElementObjectType(), result)
	diags.Append(md...)
	if md.HasError() {
		return
	}
	diags.Append(state.Set(ctx, &models.NotificationChannelsResource{NotificationChannels: mapVal})...)
}

// bulkCreate chunks create requests by the API limit and returns the created channel IDs in request
// order. On a chunk failure it returns the IDs created so far plus the error.
func (r *notificationChannelsResource) bulkCreate(ctx context.Context, reqs []labs.CreateNotificationChannelRequest) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createdIds []string
	for _, chunk := range chunkSlice(reqs, notificationChannelsBulkCreateLimit) {
		ncResp, err := r.labsClient.LabsCreateNotificationChannelsWithResponse(ctx, r.organizationId, labs.BulkCreateNotificationChannelsRequest{NotificationChannels: chunk})
		if err != nil {
			tflog.Error(ctx, "failed to bulk create notification channels", map[string]interface{}{"error": err})
			diags.AddError("Client Error", fmt.Sprintf("Unable to bulk create notification channels: %s", err))
			return createdIds, diags
		}
		// Callers zip these IDs against request keys by index, so skipping a chunk or taking a
		// short one would map every later key to the wrong channel. Stop instead, which leaves
		// createdIds a valid prefix of the requests.
		if _, d := clients.NormalizeAPIResponseWithBody(ctx, ncResp.HTTPResponse, ncResp.Body, ncResp.JSON200, "bulk create notification channels"); d != nil {
			diags.Append(d)
			return createdIds, diags
		}
		if len(ncResp.JSON200.NotificationChannels) != len(chunk) {
			tflog.Error(ctx, "bulk create notification channels returned an unexpected number of channels", map[string]interface{}{
				"requested": len(chunk), "returned": len(ncResp.JSON200.NotificationChannels),
			})
			diags.AddError("Client Error", fmt.Sprintf(
				"Unable to bulk create notification channels, requested %d but the API returned %d",
				len(chunk), len(ncResp.JSON200.NotificationChannels)))
			return createdIds, diags
		}
		for _, nc := range ncResp.JSON200.NotificationChannels {
			createdIds = append(createdIds, nc.Id)
		}
	}
	return createdIds, diags
}

// updateChannel updates a single notification channel through the platform single-channel endpoint.
func (r *notificationChannelsResource) updateChannel(ctx context.Context, id string, elem models.NotificationChannelsResourceElementModel) diag.Diagnostics {
	updateBody, diags := BuildPlatformUpdateNotificationChannelBody(ctx, elem)
	if diags.HasError() {
		return diags
	}
	ncResp, err := r.platformClient.UpdateNotificationChannelWithResponse(ctx, r.organizationId, id, updateBody)
	if err != nil {
		tflog.Error(ctx, "failed to update notification channel", map[string]interface{}{"error": err, "id": id})
		diags.AddError("Client Error", fmt.Sprintf("Unable to update notification channel %s: %s", id, err))
		return diags
	}
	if _, d := clients.NormalizeAPIResponseWithBody(ctx, ncResp.HTTPResponse, ncResp.Body, ncResp.JSON200, "update notification channel"); d != nil {
		diags.Append(d)
		return diags
	}
	return diags
}

// bulkDelete chunks channel IDs by the API limit and deletes them.
func (r *notificationChannelsResource) bulkDelete(ctx context.Context, ids []string) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, chunk := range chunkSlice(ids, notificationChannelsBulkDeleteLimit) {
		ncResp, err := r.labsClient.LabsDeleteNotificationChannelsWithResponse(ctx, r.organizationId, labs.BulkDeleteNotificationChannelsRequest{NotificationChannelIds: chunk})
		if err != nil {
			tflog.Error(ctx, "failed to bulk delete notification channels", map[string]interface{}{"error": err})
			diags.AddError("Client Error", fmt.Sprintf("Unable to bulk delete notification channels: %s", err))
			return diags
		}
		statusCode, d := clients.NormalizeAPIError(ctx, ncResp.HTTPResponse, ncResp.Body)
		if statusCode != http.StatusNotFound && d != nil {
			diags.Append(d)
			return diags
		}
	}
	return diags
}

// listByIds fetches notification channels by ID (chunked) via the platform list endpoint and returns
// them keyed by channel ID. Labs has no list endpoint, so reads route through platform v1beta1.
func (r *notificationChannelsResource) listByIds(ctx context.Context, ids []string) (map[string]platform.NotificationChannel, diag.Diagnostics) {
	var diags diag.Diagnostics
	found := make(map[string]platform.NotificationChannel, len(ids))
	for _, chunk := range chunkSlice(ids, notificationChannelsListPageLimit) {
		channelIds := chunk
		limit := len(chunk)
		params := &platform.ListNotificationChannelsParams{
			NotificationChannelIds: &channelIds,
			Limit:                  &limit,
		}
		listResp, err := r.platformClient.ListNotificationChannelsWithResponse(ctx, r.organizationId, params)
		if err != nil {
			tflog.Error(ctx, "failed to list notification channels", map[string]interface{}{"error": err})
			diags.AddError("Client Error", fmt.Sprintf("Unable to list notification channels: %s", err))
			return found, diags
		}
		if _, d := clients.NormalizeAPIError(ctx, listResp.HTTPResponse, listResp.Body); d != nil {
			diags.Append(d)
			return found, diags
		}
		if listResp.JSON200 != nil {
			for _, nc := range listResp.JSON200.NotificationChannels {
				found[nc.Id] = nc
			}
		}
	}
	return found, diags
}
