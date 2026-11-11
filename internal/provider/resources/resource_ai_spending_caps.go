package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
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

const (
	// aiSpendingCapsBulkLimit is the most entries the API takes in one set or remove request. The
	// resource chunks larger configurations across requests.
	aiSpendingCapsBulkLimit = 500
	// aiSpendingCapsListPageLimit is the page size of the list endpoint, which takes at most 100.
	aiSpendingCapsListPageLimit = 100
)

var (
	_ resource.Resource                   = &aiSpendingCapsResource{}
	_ resource.ResourceWithConfigure      = &aiSpendingCapsResource{}
	_ resource.ResourceWithImportState    = &aiSpendingCapsResource{}
	_ resource.ResourceWithValidateConfig = &aiSpendingCapsResource{}
)

func NewAiSpendingCapsResource() resource.Resource {
	return &aiSpendingCapsResource{}
}

// capSubject is one of the two kinds of subject with a cap, with its spelling in each request type
// the generated client has for it.
type capSubject struct {
	label   string
	set     labs.AiSpendingCapRequestSubjectType
	remove  labs.AiSpendingCapSubjectSubjectType
	list    labs.LabsListAiSpendingCapsParamsSubjectType
	removed labs.DeleteAiSpendingCapDefaultsRequestSubjectTypes
}

var (
	capSubjectUser = capSubject{
		label:   "user",
		set:     labs.AiSpendingCapRequestSubjectTypeUSER,
		remove:  labs.AiSpendingCapSubjectSubjectTypeUSER,
		list:    labs.USER,
		removed: labs.DeleteAiSpendingCapDefaultsRequestSubjectTypesUSER,
	}
	capSubjectApiToken = capSubject{
		label:   "API token",
		set:     labs.AiSpendingCapRequestSubjectTypeSERVICEKEY,
		remove:  labs.AiSpendingCapSubjectSubjectTypeSERVICEKEY,
		list:    labs.SERVICEKEY,
		removed: labs.DeleteAiSpendingCapDefaultsRequestSubjectTypesSERVICEKEY,
	}
)

// aiSpendingCapsResource manages an organization's per-user and per-API-token AI spending caps as a
// single resource: the two defaults, and the override of each user and API token. An attribute that
// is not set is not managed. A map that is set is authoritative for its kind of subject, so a cap
// that is not in it is removed on apply.
//
// It runs over the labs AI spending caps endpoints, which Core serves only to organizations it has
// turned them on for.
type aiSpendingCapsResource struct {
	labsClient     *labs.ClientWithResponses
	organizationId string
}

func (r *aiSpendingCapsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ai_spending_caps"
}

func (r *aiSpendingCapsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage an organization's per-user and per-API-token Astro AI spending caps as a single " +
			"resource: the default cap for users, the default cap for API tokens, and the override of each user and " +
			"API token. A cap is a monthly limit in USD cents on Astro AI spend, and an override replaces the default " +
			"for that user or API token. `0` turns AI off.\n\n" +
			"~> **Labs.** The AI spending caps API is a Labs feature, turned on per organization. If it is not turned " +
			"on for your organization, every request fails with a \"not available\" error.\n\n" +
			"~> **Warning: authoritative maps.** When `user_cap_cents` or `api_token_cap_cents` is set, any user or API " +
			"token with an override that is not in the map has it removed on apply, including overrides set in the " +
			"Astro UI. A removed override makes the user or API token follow the default, and a default of `0` turns AI " +
			"off for them. Leave a map unset to leave that kind of override to the Astro UI, and use at most one " +
			"`astro_ai_spending_caps` resource.\n\n" +
			"~> **Note** Destroying this resource removes the defaults and overrides it manages, so those users and API " +
			"tokens have no per-subject limit afterwards. The organization-wide AI spending cap is not managed here.",
		Attributes: schemas.AiSpendingCapsResourceSchemaAttributes(),
	}
}

func (r *aiSpendingCapsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ValidateConfig refuses a configuration that gives more users and API tokens an override
// than an organization can have, so the apply does not fail partway on the API's limit. The limit
// counts the organization's other caps too, which only the API knows.
func (r *aiSpendingCapsResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data models.AiSpendingCapsResource
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() || data.UserCapCents.IsUnknown() || data.ApiTokenCapCents.IsUnknown() {
		return
	}
	if total := len(data.UserCapCents.Elements()) + len(data.ApiTokenCapCents.Elements()); total > schemas.MaxAiSpendingCapSubjects {
		resp.Diagnostics.AddAttributeError(path.Root("user_cap_cents"), "Too many overrides",
			fmt.Sprintf("An organization can have at most %d users and API tokens with an override, and this configuration has %d.",
				schemas.MaxAiSpendingCapSubjects, total))
	}
}

func (r *aiSpendingCapsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.AiSpendingCapsResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.apply(ctx, plan, models.AiSpendingCapsResource{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Store the planned values rather than re-reading them: a set attribute is Optional, so
	// Terraform requires the saved value to equal the planned one. Read reconciles drift, such as
	// a cap set in the UI for a user the map does not list, on the next refresh.
	plan.Id = types.StringValue(r.organizationId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *aiSpendingCapsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.AiSpendingCapsResource
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only what the resource manages is read: an attribute that is null in state stays null, so a
	// default or cap it does not manage never shows as a change.
	if !data.DefaultUserCapCents.IsNull() || !data.DefaultApiTokenCapCents.IsNull() {
		defaults, diags := r.getDefaults(ctx)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if !data.DefaultUserCapCents.IsNull() {
			data.DefaultUserCapCents = int64FromCap(defaults.UserSpendingCap)
		}
		if !data.DefaultApiTokenCapCents.IsNull() {
			data.DefaultApiTokenCapCents = int64FromCap(defaults.ApiTokenSpendingCap)
		}
	}
	for _, kind := range []struct {
		subject capSubject
		attr    *types.Map
	}{{capSubjectUser, &data.UserCapCents}, {capSubjectApiToken, &data.ApiTokenCapCents}} {
		if kind.attr.IsNull() {
			continue
		}
		caps, diags := r.listCaps(ctx, kind.subject)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		value, diags := capsToMap(ctx, caps)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		*kind.attr = value
	}

	data.Id = types.StringValue(r.organizationId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *aiSpendingCapsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state models.AiSpendingCapsResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.apply(ctx, plan, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(r.organizationId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes what the resource manages: a configuration that sets nothing, applied over the
// state.
func (r *aiSpendingCapsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.AiSpendingCapsResource
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.apply(ctx, models.AiSpendingCapsResource{}, state)...)
}

// ImportState imports the organization's AI spending caps. The resource is a singleton for the
// organization configured on the provider, so the import ID is only cosmetic (use the organization
// ID). Everything is read: both defaults when set, and every user's and API token's override. The
// configuration then has to list what it keeps, since an attribute missing from it is removed.
func (r *aiSpendingCapsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	defaults, diags := r.getDefaults(ctx)
	resp.Diagnostics.Append(diags...)
	userCaps, diags := r.listCaps(ctx, capSubjectUser)
	resp.Diagnostics.Append(diags...)
	apiTokenCaps, diags := r.listCaps(ctx, capSubjectApiToken)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	userMap, diags := capsToMap(ctx, userCaps)
	resp.Diagnostics.Append(diags...)
	apiTokenMap, diags := capsToMap(ctx, apiTokenCaps)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data := models.AiSpendingCapsResource{
		Id:                      types.StringValue(r.organizationId),
		DefaultUserCapCents:     int64FromCap(defaults.UserSpendingCap),
		DefaultApiTokenCapCents: int64FromCap(defaults.ApiTokenSpendingCap),
		UserCapCents:            userMap,
		ApiTokenCapCents:        apiTokenMap,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// apply brings the organization from the state to the plan. A set attribute is set when it is new
// or differs, and an attribute the plan leaves out is removed when the state had it, which is what
// the resource set. Everything that sets runs before anything that removes, and overrides are set
// before defaults, so a user or API token never falls back to a default, which may be 0, while it
// is waiting for its new cap. Every step is idempotent, so an apply that fails partway can be run
// again, and a step that fails stops the ones after it.
func (r *aiSpendingCapsResource) apply(ctx context.Context, plan, state models.AiSpendingCapsResource) diag.Diagnostics {
	var diags diag.Diagnostics

	userSet, userRemoved := diffDefault(plan.DefaultUserCapCents, state.DefaultUserCapCents)
	tokenSet, tokenRemoved := diffDefault(plan.DefaultApiTokenCapCents, state.DefaultApiTokenCapCents)

	var sets []labs.AiSpendingCapRequest
	var removes []labs.AiSpendingCapSubject
	for _, kind := range []struct {
		subject capSubject
		plan    types.Map
		state   types.Map
	}{{capSubjectUser, plan.UserCapCents, state.UserCapCents}, {capSubjectApiToken, plan.ApiTokenCapCents, state.ApiTokenCapCents}} {
		desired, d := capsFromMap(ctx, kind.plan)
		diags.Append(d...)
		current, d := capsFromMap(ctx, kind.state)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		changes := diffCaps(desired, current)
		for _, id := range sortedKeys(changes.set) {
			sets = append(sets, labs.AiSpendingCapRequest{SubjectId: id, SubjectType: kind.subject.set, SpendingCap: int(changes.set[id])})
		}
		for _, id := range changes.remove {
			removes = append(removes, labs.AiSpendingCapSubject{SubjectId: id, SubjectType: kind.subject.remove})
		}
	}

	// A subject's new override goes in before a new default can apply to it.
	if len(sets) > 0 {
		diags.Append(r.setCaps(ctx, sets)...)
	}
	if !diags.HasError() && (userSet != nil || tokenSet != nil) {
		diags.Append(r.setDefaults(ctx, userSet, tokenSet)...)
	}
	if !diags.HasError() && len(removes) > 0 {
		diags.Append(r.removeCaps(ctx, removes)...)
	}
	var removedDefaults []labs.DeleteAiSpendingCapDefaultsRequestSubjectTypes
	if userRemoved {
		removedDefaults = append(removedDefaults, capSubjectUser.removed)
	}
	if tokenRemoved {
		removedDefaults = append(removedDefaults, capSubjectApiToken.removed)
	}
	if !diags.HasError() && len(removedDefaults) > 0 {
		diags.Append(r.removeDefaults(ctx, removedDefaults)...)
	}
	return diags
}

// capChanges is what one map needs: the caps to set and the subjects whose override to remove.
type capChanges struct {
	set    map[string]int64
	remove []string
}

// diffCaps compares the caps the plan wants with the ones in the state. A nil desired map is an
// attribute the plan leaves out, so everything the state held is removed. A map that is set, even
// an empty one, keeps only what it lists.
func diffCaps(desired, current map[string]int64) capChanges {
	changes := capChanges{set: map[string]int64{}}
	for id, value := range desired {
		if existing, ok := current[id]; !ok || existing != value {
			changes.set[id] = value
		}
	}
	for id := range current {
		if _, ok := desired[id]; !ok {
			changes.remove = append(changes.remove, id)
		}
	}
	sort.Strings(changes.remove)
	return changes
}

// diffDefault says what a default needs: a value to set when the plan has one that is new or
// differs, or a removal when the plan leaves it out and the state had it.
func diffDefault(plan, state types.Int64) (set *int, remove bool) {
	switch {
	case plan.IsNull() || plan.IsUnknown():
		return nil, !state.IsNull()
	case state.IsNull() || state.ValueInt64() != plan.ValueInt64():
		return new(int(plan.ValueInt64())), false
	}
	return nil, false
}

func (r *aiSpendingCapsResource) setDefaults(ctx context.Context, user, apiToken *int) diag.Diagnostics {
	var diags diag.Diagnostics
	resp, err := r.labsClient.LabsUpdateAiSpendingCapDefaultsWithResponse(ctx, r.organizationId,
		labs.UpdateAiSpendingCapDefaultsRequest{UserSpendingCap: user, ApiTokenSpendingCap: apiToken})
	if err != nil {
		tflog.Error(ctx, "failed to set the default AI spending caps", map[string]interface{}{"error": err})
		diags.AddError("Client Error", fmt.Sprintf("Unable to set the default AI spending caps: %s", err))
		return diags
	}
	if d := capsAPIProblem(ctx, resp.HTTPResponse, resp.Body, resp.JSON200, "set the default AI spending caps"); d != nil {
		diags.Append(d)
	}
	return diags
}

func (r *aiSpendingCapsResource) removeDefaults(ctx context.Context, subjectTypes []labs.DeleteAiSpendingCapDefaultsRequestSubjectTypes) diag.Diagnostics {
	var diags diag.Diagnostics
	resp, err := r.labsClient.LabsDeleteAiSpendingCapDefaultsWithResponse(ctx, r.organizationId,
		labs.DeleteAiSpendingCapDefaultsRequest{SubjectTypes: subjectTypes})
	if err != nil {
		tflog.Error(ctx, "failed to remove the default AI spending caps", map[string]interface{}{"error": err})
		diags.AddError("Client Error", fmt.Sprintf("Unable to remove the default AI spending caps: %s", err))
		return diags
	}
	if d := capsAPIProblem(ctx, resp.HTTPResponse, resp.Body, resp.JSON200, "remove the default AI spending caps"); d != nil {
		diags.Append(d)
	}
	return diags
}

func (r *aiSpendingCapsResource) getDefaults(ctx context.Context) (labs.AiSpendingCapDefaults, diag.Diagnostics) {
	var diags diag.Diagnostics
	resp, err := r.labsClient.LabsGetAiSpendingCapDefaultsWithResponse(ctx, r.organizationId)
	if err != nil {
		tflog.Error(ctx, "failed to get the default AI spending caps", map[string]interface{}{"error": err})
		diags.AddError("Client Error", fmt.Sprintf("Unable to get the default AI spending caps: %s", err))
		return labs.AiSpendingCapDefaults{}, diags
	}
	// A missing body must not read as "no defaults": for a managed default that would plan setting
	// it again, and a default removed behind Terraform's back is the one thing worth surfacing.
	if d := capsAPIProblem(ctx, resp.HTTPResponse, resp.Body, resp.JSON200, "get the default AI spending caps"); d != nil {
		diags.Append(d)
		return labs.AiSpendingCapDefaults{}, diags
	}
	return *resp.JSON200, diags
}

// setCaps sets the override of each entry, chunked by the API's per-request limit. The response
// reports each entry, so a failed one is an error even though the request succeeded.
func (r *aiSpendingCapsResource) setCaps(ctx context.Context, entries []labs.AiSpendingCapRequest) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, chunk := range chunkSlice(entries, aiSpendingCapsBulkLimit) {
		resp, err := r.labsClient.LabsUpdateAiSpendingCapsWithResponse(ctx, r.organizationId, labs.UpdateAiSpendingCapsRequest{SpendingCaps: chunk})
		if err != nil {
			tflog.Error(ctx, "failed to set AI spending caps", map[string]interface{}{"error": err})
			diags.AddError("Client Error", fmt.Sprintf("Unable to set AI spending caps: %s", err))
			return diags
		}
		if d := capsAPIProblem(ctx, resp.HTTPResponse, resp.Body, resp.JSON200, "set AI spending caps"); d != nil {
			diags.Append(d)
			return diags
		}
		if d := failedEntries("set", resp.JSON200.SpendingCaps); d != nil {
			diags.Append(d)
			return diags
		}
	}
	return diags
}

// removeCaps removes the override of each subject, chunked by the API's per-request limit.
func (r *aiSpendingCapsResource) removeCaps(ctx context.Context, subjects []labs.AiSpendingCapSubject) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, chunk := range chunkSlice(subjects, aiSpendingCapsBulkLimit) {
		resp, err := r.labsClient.LabsDeleteAiSpendingCapsWithResponse(ctx, r.organizationId, labs.DeleteAiSpendingCapsRequest{Subjects: chunk})
		if err != nil {
			tflog.Error(ctx, "failed to remove AI spending caps", map[string]interface{}{"error": err})
			diags.AddError("Client Error", fmt.Sprintf("Unable to remove AI spending caps: %s", err))
			return diags
		}
		if d := capsAPIProblem(ctx, resp.HTTPResponse, resp.Body, resp.JSON200, "remove AI spending caps"); d != nil {
			diags.Append(d)
			return diags
		}
		if d := failedEntries("remove", resp.JSON200.SpendingCaps); d != nil {
			diags.Append(d)
			return diags
		}
	}
	return diags
}

// listCaps pages through the organization's users or API tokens that have an override and
// returns them by ID. The resource owns the list, so a page that goes missing would plan the
// removal of caps that exist or hide ones that should go; the number of subjects read must match
// the count the API reports.
func (r *aiSpendingCapsResource) listCaps(ctx context.Context, subject capSubject) (map[string]int64, diag.Diagnostics) {
	var diags diag.Diagnostics
	caps := map[string]int64{}
	limit := aiSpendingCapsListPageLimit
	offset := 0
	hasOverride := true
	for {
		resp, err := r.labsClient.LabsListAiSpendingCapsWithResponse(ctx, r.organizationId, &labs.LabsListAiSpendingCapsParams{
			SubjectType: subject.list, HasOverride: &hasOverride, Limit: &limit, Offset: &offset,
		})
		if err != nil {
			tflog.Error(ctx, "failed to list AI spending caps", map[string]interface{}{"error": err})
			diags.AddError("Client Error", fmt.Sprintf("Unable to list the AI spending caps of %ss: %s", subject.label, err))
			return nil, diags
		}
		if d := capsAPIProblem(ctx, resp.HTTPResponse, resp.Body, resp.JSON200, fmt.Sprintf("list the AI spending caps of %ss", subject.label)); d != nil {
			diags.Append(d)
			return nil, diags
		}
		page := resp.JSON200
		for _, entry := range page.AiSpendingCaps {
			if entry.SpendingCap != nil {
				caps[entry.SubjectId] = int64(*entry.SpendingCap)
			}
		}
		offset += len(page.AiSpendingCaps)
		if len(page.AiSpendingCaps) < limit || offset >= page.TotalCount {
			if len(caps) != page.TotalCount {
				diags.AddError("Client Error", fmt.Sprintf("Read %d %ss with a spending cap but the API reports %d. Run the plan again.", len(caps), subject.label, page.TotalCount))
				return nil, diags
			}
			return caps, diags
		}
	}
}

// capsAPIProblem turns a response that is not a success with a body into a diagnostic. Core answers
// 404 "Path ... not found" when it is not serving the endpoints to this organization, and never when
// the resource is gone, so that is an error here and the resource is never dropped from state on it.
// Any other 404, such as a user or API token that does not exist, is reported with Core's message.
func capsAPIProblem[T any](ctx context.Context, httpResp *http.Response, body []byte, json200 *T, operation string) diag.Diagnostic {
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound && isGateNotFound(body) {
		return diag.NewErrorDiagnostic(
			"AI spending caps are not available",
			fmt.Sprintf("Unable to %s: the AI spending caps API answered 404. It is a Labs feature that Astro turns on per organization, "+
				"so it is not turned on for this organization yet. Contact Astronomer to have it turned on.", operation),
		)
	}
	_, d := clients.NormalizeAPIResponseWithBody(ctx, httpResp, body, json200, operation)
	return d
}

// isGateNotFound says whether a 404 body is Core's answer for an endpoint it does not serve to the
// organization ("Path ... not found"), as opposed to a 404 for something in the request.
func isGateNotFound(body []byte) bool {
	var decoded struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		return false
	}
	return strings.HasPrefix(decoded.Message, "Path ") && strings.HasSuffix(decoded.Message, " not found")
}

// failedEntries is an error naming the subjects whose entry failed, or nil. A failed entry changed
// nothing, so the apply can be run again.
func failedEntries(verb string, results []labs.AiSpendingCapResult) diag.Diagnostic {
	var failed []string
	for _, result := range results {
		if result.Status == labs.FAILED {
			failed = append(failed, result.SubjectId)
		}
	}
	if len(failed) == 0 {
		return nil
	}
	sort.Strings(failed)
	return diag.NewErrorDiagnostic(
		"Some AI spending caps could not be changed",
		fmt.Sprintf("Unable to %s the AI spending cap of %d users or API tokens: %s. Those entries changed nothing; run the apply again.",
			verb, len(failed), strings.Join(failed, ", ")),
	)
}

func capsFromMap(ctx context.Context, m types.Map) (map[string]int64, diag.Diagnostics) {
	if m.IsNull() || m.IsUnknown() {
		return nil, nil
	}
	caps := map[string]int64{}
	diags := m.ElementsAs(ctx, &caps, false)
	return caps, diags
}

func capsToMap(ctx context.Context, caps map[string]int64) (types.Map, diag.Diagnostics) {
	return types.MapValueFrom(ctx, types.Int64Type, caps)
}

func int64FromCap(value *int) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*value))
}
