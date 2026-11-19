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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
	_ resource.ResourceWithModifyPlan     = &aiSpendingCapsResource{}
)

func NewAiSpendingCapsResource() resource.Resource {
	return &aiSpendingCapsResource{}
}

// capSubject is one of the two kinds of subject with a cap, with its spelling in each request type
// the generated client has for it.
type capSubject struct {
	label  string
	set    labs.AiSpendingCapRequestSubjectType
	remove labs.AiSpendingCapSubjectSubjectType
	list   labs.LabsListAiSpendingCapsParamsSubjectType
}

var (
	capSubjectUser = capSubject{
		label:  "user",
		set:    labs.AiSpendingCapRequestSubjectTypeUSER,
		remove: labs.AiSpendingCapSubjectSubjectTypeUSER,
		list:   labs.USER,
	}
	capSubjectApiToken = capSubject{
		label:  "API token",
		set:    labs.AiSpendingCapRequestSubjectTypeSERVICEKEY,
		remove: labs.AiSpendingCapSubjectSubjectTypeSERVICEKEY,
		list:   labs.SERVICEKEY,
	}
)

// aiSpendingCapsResource manages an organization's per-user and per-API-token AI spending caps as a
// single resource: the two defaults, and the override of each user and API token.
//
// Each of default_cap_cents, user_cap_cents and api_token_cap_cents is managed only when it is set.
// One left out is not changed. One that is set is authoritative: it is the complete list for what it
// covers, so whatever it leaves out is removed. Read always reads everything, managed or not, so the
// plan shows every change and one apply converges. A destroy changes nothing.
//
// It runs over the labs AI spending caps endpoints, which Core serves only to organizations it has
// turned them on for.
type aiSpendingCapsResource struct {
	labsClient     *labs.ClientWithResponses
	organizationId string
}

// aiSpendingCapsDestroyWarning is shown in the plan and after the apply of a destroy.
const aiSpendingCapsDestroyWarning = "Destroying astro_ai_spending_caps changes nothing in Astro. The plan lists every default " +
	"and override as removed, but they all stay in place; Terraform only stops managing them. To remove them, set " +
	"default_cap_cents, user_cap_cents and api_token_cap_cents to {} and apply before destroying."

// aiSpendingCapsMaxConflictsListed is how many existing caps the create check names before it
// counts the rest.
const aiSpendingCapsMaxConflictsListed = 10

func (r *aiSpendingCapsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ai_spending_caps"
}

func (r *aiSpendingCapsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage an organization's per-user and per-API-token Astro AI spending caps as a single " +
			"resource: the default cap for users and for API tokens, and the override of each user and API token. A cap " +
			"is a monthly limit in USD cents on Astro AI spend, and an override replaces the default for that user or " +
			"API token. `0` turns AI off.\n\n" +
			"Each of `default_cap_cents`, `user_cap_cents` and `api_token_cap_cents` is managed only when it is set:\n\n" +
			"- **Left out:** not changed. Terraform reads the current values into the state but never changes them, " +
			"and leaving the attribute out, or removing it from the configuration, shows no change in the plan.\n" +
			"- **Set:** authoritative. It is the complete list for what it covers: a default left out of a set " +
			"`default_cap_cents`, and the override of a user or API token that is not in a set map, are removed on " +
			"apply, including ones set in the Astro UI. `{}` removes everything it covers.\n\n" +
			"~> **Labs.** The AI spending caps API is a Labs feature, turned on per organization. If it is not turned " +
			"on for your organization, every request fails with a \"not available\" error.\n\n" +
			"~> **Warning: authoritative.** A set map removes overrides set in the Astro UI, and a removed override " +
			"makes the user or API token follow the default (a default of `0` turns AI off for them; removing an " +
			"override of `0` turns AI back on). Every removal is listed in the plan before it happens. Use at most one " +
			"`astro_ai_spending_caps` resource per organization.\n\n" +
			"~> **Creating the resource.** If the organization already has caps that the configuration would change " +
			"or remove, the plan fails and asks you to import the resource first (`terraform import`, or an `import` " +
			"block). After the import, the plan shows those changes like any other.\n\n" +
			"~> **Destroying the resource changes nothing in Astro.** It stops managing the caps and leaves every " +
			"default and override in place, although the destroy plan lists them as removed. To remove them, set " +
			"`default_cap_cents`, `user_cap_cents` and `api_token_cap_cents` to `{}` and apply before destroying.\n\n" +
			"-> **Users who leave the organization, and deleted API tokens.** If a user in `user_cap_cents` leaves the " +
			"organization, or an API token in `api_token_cap_cents` is deleted, the apply fails on it until you take it " +
			"out of the map. Its override stays in Astro, where neither Terraform nor the Astro UI shows it.\n\n" +
			"The organization-wide AI spending cap is not managed here.",
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

// ModifyPlan fails a create that would change or remove caps the organization already has: with no
// prior state, the plan cannot show them, so the resource has to be imported first. On a destroy it
// warns that nothing is removed.
func (r *aiSpendingCapsResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		resp.Diagnostics.AddWarning("Destroying astro_ai_spending_caps changes nothing", aiSpendingCapsDestroyWarning)
		return
	}
	if !req.State.Raw.IsNull() || r.labsClient == nil {
		return
	}
	var config models.AiSpendingCapsResource
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	current, diags := r.readAll(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	conflicts, diags := createConflicts(ctx, config, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || len(conflicts) == 0 {
		return
	}
	listed := conflicts
	if len(listed) > aiSpendingCapsMaxConflictsListed {
		listed = append(append([]string{}, conflicts[:aiSpendingCapsMaxConflictsListed]...),
			fmt.Sprintf("and %d more", len(conflicts)-aiSpendingCapsMaxConflictsListed))
	}
	resp.Diagnostics.AddError("The organization already has AI spending caps",
		fmt.Sprintf("Creating astro_ai_spending_caps would change or remove caps the organization already has (%s), and a "+
			"plan cannot show them before the resource is in the state. Import it first, with `terraform import "+
			"<address> %s` or an import block; the next plan then shows what this configuration would change or remove.",
			strings.Join(listed, "; "), r.organizationId))
}

// createConflicts names the existing caps that applying config would change or remove: for each
// attribute the config sets, a value that differs or is missing from it.
func createConflicts(ctx context.Context, config, current models.AiSpendingCapsResource) ([]string, diag.Diagnostics) {
	var conflicts []string
	var diags diag.Diagnostics
	if !config.DefaultCapCents.IsNull() && !config.DefaultCapCents.IsUnknown() {
		want, d := defaultsOf(ctx, config.DefaultCapCents)
		diags.Append(d...)
		have, d := defaultsOf(ctx, current.DefaultCapCents)
		diags.Append(d...)
		if !have.User.IsNull() && !have.User.Equal(want.User) {
			conflicts = append(conflicts, "the user default")
		}
		if !have.ApiToken.IsNull() && !have.ApiToken.Equal(want.ApiToken) {
			conflicts = append(conflicts, "the API token default")
		}
	}
	for _, kind := range []struct {
		label         string
		want, current types.Map
	}{{"user", config.UserCapCents, current.UserCapCents}, {"API token", config.ApiTokenCapCents, current.ApiTokenCapCents}} {
		if kind.want.IsNull() || kind.want.IsUnknown() {
			continue
		}
		want, d := capsFromMap(ctx, kind.want)
		diags.Append(d...)
		have, d := capsFromMap(ctx, kind.current)
		diags.Append(d...)
		for _, id := range sortedKeys(have) {
			if value, ok := want[id]; !ok || value != have[id] {
				conflicts = append(conflicts, fmt.Sprintf("the override of %s %s", kind.label, id))
			}
		}
	}
	return conflicts, diags
}

func (r *aiSpendingCapsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config models.AiSpendingCapsResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.apply(ctx, config, plan, models.AiSpendingCapsResource{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	// An attribute the config sets keeps its planned value, as Terraform requires. One it leaves out
	// is unknown in the plan, so it is read.
	current, diags := r.readAll(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.DefaultCapCents.IsNull() {
		plan.DefaultCapCents = current.DefaultCapCents
	}
	if config.UserCapCents.IsNull() {
		plan.UserCapCents = current.UserCapCents
	}
	if config.ApiTokenCapCents.IsNull() {
		plan.ApiTokenCapCents = current.ApiTokenCapCents
	}
	plan.Id = types.StringValue(r.organizationId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads everything, managed or not, so the state holds the organization's current caps.
func (r *aiSpendingCapsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	data, diags := r.readAll(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *aiSpendingCapsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state, config models.AiSpendingCapsResource
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.apply(ctx, config, plan, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Id = types.StringValue(r.organizationId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete changes nothing in Astro: destroy is not the way to remove caps, setting them to {} is.
func (r *aiSpendingCapsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("The AI spending caps were left in place", aiSpendingCapsDestroyWarning)
}

// ImportState reads everything. The resource is a singleton for the organization configured on the
// provider, so the import ID is only cosmetic (use the organization ID).
func (r *aiSpendingCapsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	data, diags := r.readAll(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readAll reads both defaults and every override of both kinds.
func (r *aiSpendingCapsResource) readAll(ctx context.Context) (models.AiSpendingCapsResource, diag.Diagnostics) {
	var diags diag.Diagnostics
	data := models.AiSpendingCapsResource{Id: types.StringValue(r.organizationId)}
	defaults, d := r.getDefaults(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return data, diags
	}
	data.DefaultCapCents, d = types.ObjectValueFrom(ctx, models.AiSpendingCapDefaultsAttrTypes, models.AiSpendingCapDefaults{
		User:     int64FromCap(defaults.UserSpendingCap),
		ApiToken: int64FromCap(defaults.ApiTokenSpendingCap),
	})
	diags.Append(d...)
	for _, kind := range []struct {
		subject capSubject
		attr    *types.Map
	}{{capSubjectUser, &data.UserCapCents}, {capSubjectApiToken, &data.ApiTokenCapCents}} {
		caps, d := r.listCaps(ctx, kind.subject)
		diags.Append(d...)
		if diags.HasError() {
			return data, diags
		}
		*kind.attr, d = capsToMap(ctx, caps)
		diags.Append(d...)
	}
	return data, diags
}

// apply brings the organization from the state to the plan for each attribute the config sets. One
// the config leaves out is not managed, whatever the plan holds for it (its state value).
// Overrides are set first, then defaults, so a user or API token never falls under a new default,
// which may be 0, while it waits for its new override. Removals come last. Every step is
// idempotent, so an apply that fails partway can be run again.
func (r *aiSpendingCapsResource) apply(ctx context.Context, config, plan, state models.AiSpendingCapsResource) diag.Diagnostics {
	var diags diag.Diagnostics

	var userSet, tokenSet *int
	var removeDefaults []labs.DeleteAiSpendingCapDefaultsRequestSubjectTypes
	if !config.DefaultCapCents.IsNull() {
		want, d := defaultsOf(ctx, plan.DefaultCapCents)
		diags.Append(d...)
		have, d := defaultsOf(ctx, state.DefaultCapCents)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		userSet = diffDefault(want.User, have.User)
		tokenSet = diffDefault(want.ApiToken, have.ApiToken)
		if want.User.IsNull() && !have.User.IsNull() {
			removeDefaults = append(removeDefaults, labs.DeleteAiSpendingCapDefaultsRequestSubjectTypesUSER)
		}
		if want.ApiToken.IsNull() && !have.ApiToken.IsNull() {
			removeDefaults = append(removeDefaults, labs.DeleteAiSpendingCapDefaultsRequestSubjectTypesSERVICEKEY)
		}
	}

	var sets []labs.AiSpendingCapRequest
	var removes []labs.AiSpendingCapSubject
	for _, kind := range []struct {
		subject             capSubject
		config, plan, state types.Map
	}{{capSubjectUser, config.UserCapCents, plan.UserCapCents, state.UserCapCents}, {capSubjectApiToken, config.ApiTokenCapCents, plan.ApiTokenCapCents, state.ApiTokenCapCents}} {
		if kind.config.IsNull() {
			continue
		}
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

	if len(sets) > 0 {
		diags.Append(r.setCaps(ctx, sets)...)
	}
	if !diags.HasError() && (userSet != nil || tokenSet != nil) {
		diags.Append(r.setDefaults(ctx, userSet, tokenSet)...)
	}
	if !diags.HasError() && len(removeDefaults) > 0 {
		diags.Append(r.removeDefaults(ctx, removeDefaults)...)
	}
	if !diags.HasError() && len(removes) > 0 {
		diags.Append(r.removeCaps(ctx, removes)...)
	}
	return diags
}

// defaultsOf reads default_cap_cents; a null or unknown object reads as no defaults.
func defaultsOf(ctx context.Context, o types.Object) (models.AiSpendingCapDefaults, diag.Diagnostics) {
	if o.IsNull() || o.IsUnknown() {
		return models.AiSpendingCapDefaults{User: types.Int64Null(), ApiToken: types.Int64Null()}, nil
	}
	var d models.AiSpendingCapDefaults
	diags := o.As(ctx, &d, basetypes.ObjectAsOptions{})
	return d, diags
}

// capChanges is what one map needs: the caps to set and the subjects whose override to remove.
type capChanges struct {
	set    map[string]int64
	remove []string
}

// diffCaps compares the caps the plan wants with the ones in the state. A map that is set, even an
// empty one, keeps only what it lists. The caller does not call it for a map the config leaves out.
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

// diffDefault is the value to set when the plan has one that is new or differs, nil otherwise.
// Removing a default is the caller's job.
func diffDefault(plan, state types.Int64) *int {
	if plan.IsNull() || plan.IsUnknown() {
		return nil
	}
	if state.IsNull() || state.ValueInt64() != plan.ValueInt64() {
		return new(int(plan.ValueInt64()))
	}
	return nil
}

func (r *aiSpendingCapsResource) removeDefaults(ctx context.Context, kinds []labs.DeleteAiSpendingCapDefaultsRequestSubjectTypes) diag.Diagnostics {
	var diags diag.Diagnostics
	resp, err := r.labsClient.LabsDeleteAiSpendingCapDefaultsWithResponse(ctx, r.organizationId, labs.DeleteAiSpendingCapDefaultsRequest{SubjectTypes: kinds})
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
