package resources

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients/labs"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests call the resource's methods directly, the way the framework does, for what is easier
// to check here than through the Terraform CLI: the requests sent and their order, chunking, paging,
// failed entries and the 404 gate. ai_spending_caps_terraform_test.go runs the real CLI.

// capsHarness drives the resource's methods against the fake.
type capsHarness struct {
	t       *testing.T
	api     *FakeAiSpendingCapsAPI
	r       *aiSpendingCapsResource
	schema  rschema.Schema
	objType tftypes.Object
}

func newCapsHarness(t *testing.T) *capsHarness {
	t.Helper()
	api := NewFakeAiSpendingCapsAPI()
	srv := httptest.NewServer(api)
	t.Cleanup(srv.Close)
	labsClient, err := labs.NewLabsClient(srv.URL, "token", "test")
	require.NoError(t, err)
	s := rschema.Schema{Attributes: schemas.AiSpendingCapsResourceSchemaAttributes()}
	return &capsHarness{
		t: t, api: api, schema: s,
		r:       &aiSpendingCapsResource{labsClient: labsClient, organizationId: "org"},
		objType: s.Type().TerraformType(context.Background()).(tftypes.Object),
	}
}

// defaults is a default_cap_cents object; a nil field is left out of it.
type defaults struct{ user, apiToken *int64 }

// config is a configuration; a nil field is an attribute left out.
type config struct {
	defaults         *defaults
	users, apiTokens map[string]int64
}

func capsMapValue(caps map[string]int64) types.Map {
	if caps == nil {
		return types.MapNull(types.Int64Type)
	}
	value, _ := types.MapValueFrom(context.Background(), types.Int64Type, caps)
	return value
}

func int64Value(value *int64) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*value)
}

func defaultsValue(d *defaults) types.Object {
	if d == nil {
		return types.ObjectNull(models.AiSpendingCapDefaultsAttrTypes)
	}
	value, _ := types.ObjectValueFrom(context.Background(), models.AiSpendingCapDefaultsAttrTypes,
		models.AiSpendingCapDefaults{User: int64Value(d.user), ApiToken: int64Value(d.apiToken)})
	return value
}

// model is the configuration as the resource reads it: what is left out is null.
func (c config) model() models.AiSpendingCapsResource {
	return models.AiSpendingCapsResource{
		Id:               types.StringNull(),
		DefaultCapCents:  defaultsValue(c.defaults),
		UserCapCents:     capsMapValue(c.users),
		ApiTokenCapCents: capsMapValue(c.apiTokens),
	}
}

// planFor is what Terraform plans for the configuration: what it sets, and for what it leaves out,
// the prior state's value (UseStateForUnknown), or unknown on a create.
func planFor(c config, prior *models.AiSpendingCapsResource) models.AiSpendingCapsResource {
	plan := c.model()
	if prior == nil {
		plan.Id = types.StringUnknown()
		if c.defaults == nil {
			plan.DefaultCapCents = types.ObjectUnknown(models.AiSpendingCapDefaultsAttrTypes)
		}
		if c.users == nil {
			plan.UserCapCents = types.MapUnknown(types.Int64Type)
		}
		if c.apiTokens == nil {
			plan.ApiTokenCapCents = types.MapUnknown(types.Int64Type)
		}
		return plan
	}
	plan.Id = prior.Id
	if c.defaults == nil {
		plan.DefaultCapCents = prior.DefaultCapCents
	}
	if c.users == nil {
		plan.UserCapCents = prior.UserCapCents
	}
	if c.apiTokens == nil {
		plan.ApiTokenCapCents = prior.ApiTokenCapCents
	}
	return plan
}

func (h *capsHarness) null() tftypes.Value { return tftypes.NewValue(h.objType, nil) }

func (h *capsHarness) raw(m models.AiSpendingCapsResource) tftypes.Value {
	p := tfsdk.Plan{Schema: h.schema, Raw: h.null()}
	require.False(h.t, p.Set(context.Background(), &m).HasError())
	return p.Raw
}

func (h *capsHarness) tfConfig(c config) tfsdk.Config {
	return tfsdk.Config{Schema: h.schema, Raw: h.raw(c.model())}
}

func (h *capsHarness) tfPlan(m models.AiSpendingCapsResource) tfsdk.Plan {
	return tfsdk.Plan{Schema: h.schema, Raw: h.raw(m)}
}

func (h *capsHarness) tfState(m models.AiSpendingCapsResource) tfsdk.State {
	return tfsdk.State{Schema: h.schema, Raw: h.raw(m)}
}

func (h *capsHarness) model(st tfsdk.State) models.AiSpendingCapsResource {
	var m models.AiSpendingCapsResource
	require.False(h.t, st.Get(context.Background(), &m).HasError())
	return m
}

func (h *capsHarness) create(c config) *resource.CreateResponse {
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: h.schema, Raw: h.null()}}
	h.r.Create(context.Background(), resource.CreateRequest{Config: h.tfConfig(c), Plan: h.tfPlan(planFor(c, nil))}, resp)
	return resp
}

// refresh is the state Terraform hands to Update and Delete: it refreshes before it plans.
func (h *capsHarness) refresh() models.AiSpendingCapsResource {
	resp := h.read()
	require.False(h.t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	return h.model(resp.State)
}

func (h *capsHarness) read() *resource.ReadResponse {
	// Read ignores the prior state, so any will do.
	prior := h.tfState(models.AiSpendingCapsResource{
		Id: types.StringValue("org"), DefaultCapCents: defaultsValue(nil),
		UserCapCents: capsMapValue(nil), ApiTokenCapCents: capsMapValue(nil),
	})
	resp := &resource.ReadResponse{State: prior}
	h.r.Read(context.Background(), resource.ReadRequest{State: prior}, resp)
	return resp
}

func (h *capsHarness) update(c config, state models.AiSpendingCapsResource) *resource.UpdateResponse {
	prior := h.tfState(state)
	resp := &resource.UpdateResponse{State: prior}
	h.r.Update(context.Background(), resource.UpdateRequest{Config: h.tfConfig(c), Plan: h.tfPlan(planFor(c, &state)), State: prior}, resp)
	return resp
}

func (h *capsHarness) planCreate(c config) *resource.ModifyPlanResponse {
	plan := h.tfPlan(planFor(c, nil))
	resp := &resource.ModifyPlanResponse{Plan: plan}
	h.r.ModifyPlan(context.Background(), resource.ModifyPlanRequest{
		Config: h.tfConfig(c), Plan: plan, State: tfsdk.State{Schema: h.schema, Raw: h.null()},
	}, resp)
	return resp
}

func (h *capsHarness) destroy(state models.AiSpendingCapsResource) *resource.DeleteResponse {
	prior := h.tfState(state)
	resp := &resource.DeleteResponse{State: prior}
	h.r.Delete(context.Background(), resource.DeleteRequest{State: prior}, resp)
	return resp
}

func ptr(value int64) *int64 { return &value }

func elements(t *testing.T, m types.Map) map[string]int64 {
	t.Helper()
	out := map[string]int64{}
	require.False(t, m.ElementsAs(context.Background(), &out, false).HasError())
	return out
}

func defaultsIn(t *testing.T, o types.Object) models.AiSpendingCapDefaults {
	t.Helper()
	d, diags := defaultsOf(context.Background(), o)
	require.False(t, diags.HasError())
	return d
}

func TestUnit_AiSpendingCaps_SchemaIsValid(t *testing.T) {
	s := rschema.Schema{Attributes: schemas.AiSpendingCapsResourceSchemaAttributes()}
	assert.False(t, s.ValidateImplementation(context.Background()).HasError())
}

func TestUnit_AiSpendingCaps_Create_SetsWhatIsConfiguredAndReadsWhatIsLeftOut(t *testing.T) {
	h := newCapsHarness(t)
	h.api.Seed("SERVICEKEY", map[string]int{"set-in-the-ui": 5})

	resp := h.create(config{
		defaults: &defaults{user: ptr(5000), apiToken: ptr(2500)},
		users:    map[string]int64{"user-a": 100, "user-b": 0},
	})

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Equal(t, []string{"set-caps:2", "set-defaults", "get-defaults", "list:USER", "list:SERVICEKEY"}, h.api.CallsMade(),
		"overrides before defaults, then a read of what the configuration leaves out")
	assert.Equal(t, map[string]int{"USER": 5000, "SERVICEKEY": 2500}, h.api.Defaults)
	assert.Equal(t, map[string]int{"user-a": 100, "user-b": 0}, h.api.Caps["USER"], "0 is a cap, not an absent one")
	assert.Equal(t, map[string]int{"set-in-the-ui": 5}, h.api.Caps["SERVICEKEY"], "a map left out is not changed")
	stored := h.model(resp.State)
	assert.Equal(t, "org", stored.Id.ValueString())
	assert.Equal(t, map[string]int64{"user-a": 100, "user-b": 0}, elements(t, stored.UserCapCents))
	assert.Equal(t, map[string]int64{"set-in-the-ui": 5}, elements(t, stored.ApiTokenCapCents), "a map left out is read into the state")
}

func TestUnit_AiSpendingCaps_Create_ChunksTheBulkRequests(t *testing.T) {
	h := newCapsHarness(t)
	users := map[string]int64{}
	for i := 0; i < 1100; i++ {
		users[fmt.Sprintf("user-%04d", i)] = int64(i)
	}

	resp := h.create(config{users: users})

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Equal(t, []string{"set-caps:500", "set-caps:500", "set-caps:100"}, h.api.CallsMade()[:3])
	assert.Len(t, h.api.Caps["USER"], 1100)
}

func TestUnit_AiSpendingCaps_Update_SendsOnlyWhatChangedInOrder(t *testing.T) {
	h := newCapsHarness(t)
	h.api.Seed("USER", map[string]int{"keep": 100, "change": 200, "drop": 300})
	h.api.Defaults = map[string]int{"USER": 5000, "SERVICEKEY": 2500}
	state := h.refresh()
	callsBefore := len(h.api.CallsMade())

	resp := h.update(config{
		defaults: &defaults{user: ptr(6000)}, // api_token left out of a set object: removed
		users:    map[string]int64{"keep": 100, "change": 250, "new": 50},
	}, state)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Equal(t, []string{"set-caps:2", "set-defaults", "remove-defaults", "remove-caps:1"}, h.api.CallsMade()[callsBefore:],
		"the unchanged cap is not sent, overrides are set before defaults, and removals come last")
	assert.Equal(t, map[string]int{"keep": 100, "change": 250, "new": 50}, h.api.Caps["USER"])
	assert.Equal(t, map[string]int{"USER": 6000}, h.api.Defaults)
}

func TestUnit_AiSpendingCaps_Update_WhatIsLeftOutIsNotChanged(t *testing.T) {
	h := newCapsHarness(t)
	// An admin gave "set-in-the-ui" a 0 cap, which turns AI off for them.
	h.api.Seed("USER", map[string]int{"managed": 100, "set-in-the-ui": 0})
	h.api.Seed("SERVICEKEY", map[string]int{"token-a": 1})
	h.api.Defaults = map[string]int{"USER": 5000}
	state := h.refresh()
	callsBefore := len(h.api.CallsMade())

	resp := h.update(config{apiTokens: map[string]int64{"token-a": 2}}, state)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Equal(t, []string{"set-caps:1"}, h.api.CallsMade()[callsBefore:], "only the map the configuration sets is sent")
	assert.Equal(t, map[string]int{"managed": 100, "set-in-the-ui": 0}, h.api.Caps["USER"])
	assert.Equal(t, map[string]int{"USER": 5000}, h.api.Defaults)
	got := h.model(resp.State)
	assert.Equal(t, map[string]int64{"managed": 100, "set-in-the-ui": 0}, elements(t, got.UserCapCents), "and keeps its value in the state")
	assert.Equal(t, int64(5000), defaultsIn(t, got.DefaultCapCents).User.ValueInt64())
}

func TestUnit_AiSpendingCaps_Update_EmptyRemovesEverythingItCovers(t *testing.T) {
	h := newCapsHarness(t)
	h.api.Seed("USER", map[string]int{"a": 1, "set-in-the-ui": 2})
	h.api.Seed("SERVICEKEY", map[string]int{"token-a": 3})
	h.api.Defaults = map[string]int{"USER": 5000, "SERVICEKEY": 2500}
	state := h.refresh()

	resp := h.update(config{defaults: &defaults{}, users: map[string]int64{}}, state)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Empty(t, h.api.Defaults, "{} removes both defaults")
	assert.Empty(t, h.api.Caps["USER"], "{} removes every user's override, including ones set in the UI")
	assert.Equal(t, map[string]int{"token-a": 3}, h.api.Caps["SERVICEKEY"], "a map left out is not changed")
}

func TestUnit_AiSpendingCaps_Read_ReadsEverything(t *testing.T) {
	h := newCapsHarness(t)
	h.api.Seed("USER", map[string]int{"a": 100})
	h.api.Seed("SERVICEKEY", map[string]int{"token-a": 5})
	h.api.Defaults["USER"] = 4000

	got := h.refresh()

	assert.Equal(t, "org", got.Id.ValueString())
	assert.Equal(t, map[string]int64{"a": 100}, elements(t, got.UserCapCents))
	assert.Equal(t, map[string]int64{"token-a": 5}, elements(t, got.ApiTokenCapCents))
	d := defaultsIn(t, got.DefaultCapCents)
	assert.False(t, got.DefaultCapCents.IsNull(), "the object is always set, so {} in the configuration matches no defaults")
	assert.Equal(t, int64(4000), d.User.ValueInt64())
	assert.True(t, d.ApiToken.IsNull(), "no default is a null field")
}

func TestUnit_AiSpendingCaps_Read_PagesThroughEveryCap(t *testing.T) {
	h := newCapsHarness(t)
	for i := 0; i < 250; i++ {
		h.api.Caps["USER"][fmt.Sprintf("user-%03d", i)] = i
	}

	got := h.refresh()

	assert.Len(t, elements(t, got.UserCapCents), 250)
	assert.Equal(t, []string{"get-defaults", "list:USER", "list:USER", "list:USER", "list:SERVICEKEY"}, h.api.CallsMade(), "100 per page")
}

func TestUnit_AiSpendingCaps_Read_ErrorsWhenAPageWentMissing(t *testing.T) {
	h := newCapsHarness(t)
	h.api.Seed("USER", map[string]int{"a": 1})
	h.api.TotalCountPadding = 1

	resp := h.read()

	require.True(t, resp.Diagnostics.HasError(), "a list shorter than its count must not read as the whole list")
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "reports 2")
}

func TestUnit_AiSpendingCaps_AFailedEntryIsAnErrorNamingTheSubject(t *testing.T) {
	h := newCapsHarness(t)
	h.api.FailIds["user-b"] = true

	resp := h.create(config{users: map[string]int64{"user-a": 1, "user-b": 2}})

	require.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "user-b")
	assert.NotContains(t, resp.Diagnostics.Errors()[0].Detail(), "user-a")
}

func TestUnit_AiSpendingCaps_AnUnknownSubjectIsReportedAsItselfAndNotAsTheGate(t *testing.T) {
	h := newCapsHarness(t)
	h.api.UnknownIds["gone-token"] = true

	create := h.create(config{apiTokens: map[string]int64{"gone-token": 100}})
	require.True(t, create.Diagnostics.HasError())
	errs := create.Diagnostics.Errors()
	assert.NotContains(t, errs[0].Summary(), "not available", "a 404 for a subject does not mean the feature is off")
	assert.NotContains(t, errs[0].Detail(), "Contact Astronomer")
	assert.Contains(t, errs[0].Detail(), "api token with id gone-token not found")

	h.api.Seed("USER", map[string]int{"gone-user": 5})
	state := h.refresh()
	h.api.UnknownIds["gone-user"] = true
	update := h.update(config{users: map[string]int64{}}, state)
	require.True(t, update.Diagnostics.HasError())
	assert.NotContains(t, update.Diagnostics.Errors()[0].Summary(), "not available")
	assert.Contains(t, update.Diagnostics.Errors()[0].Detail(), "user with id gone-user not found")
}

func TestUnit_AiSpendingCaps_TheGateBeingOffIsAnErrorAndNeverDropsTheResource(t *testing.T) {
	h := newCapsHarness(t)
	h.api.GateOff = true

	read := h.read()
	require.True(t, read.Diagnostics.HasError())
	assert.Contains(t, read.Diagnostics.Errors()[0].Summary(), "not available")
	assert.False(t, read.State.Raw.IsNull(), "a 404 here means the feature is off, not that the resource is gone")

	plan := h.planCreate(config{defaults: &defaults{user: ptr(5000)}})
	require.True(t, plan.Diagnostics.HasError(), "the create check reads, so the plan says so")
	assert.Contains(t, plan.Diagnostics.Errors()[0].Detail(), "Contact Astronomer")

	create := h.create(config{defaults: &defaults{user: ptr(5000)}})
	require.True(t, create.Diagnostics.HasError())
	assert.Contains(t, create.Diagnostics.Errors()[0].Detail(), "Contact Astronomer")

	imported := &resource.ImportStateResponse{State: tfsdk.State{Schema: h.schema, Raw: h.null()}}
	h.r.ImportState(context.Background(), resource.ImportStateRequest{ID: "org"}, imported)
	assert.True(t, imported.Diagnostics.HasError())
}

func TestUnit_AiSpendingCaps_ModifyPlan_CreateFailsOnCapsItWouldChangeOrRemove(t *testing.T) {
	h := newCapsHarness(t)
	h.api.Seed("USER", map[string]int{"set-in-the-ui": 0, "same": 5})
	h.api.Defaults = map[string]int{"USER": 1000, "SERVICEKEY": 2500}

	resp := h.planCreate(config{
		defaults: &defaults{user: ptr(1000)}, // same user default; leaves the API token default out
		users:    map[string]int64{"same": 5, "new": 1},
	})

	require.True(t, resp.Diagnostics.HasError())
	detail := resp.Diagnostics.Errors()[0].Detail()
	assert.Contains(t, detail, "the override of user set-in-the-ui", "an override the map would remove")
	assert.Contains(t, detail, "the API token default", "a default the object would remove")
	assert.NotContains(t, detail, "user same", "an override the configuration keeps as it is")
	assert.NotContains(t, detail, "the user default", "a default the configuration keeps as it is")
	assert.Contains(t, detail, "terraform import")
}

func TestUnit_AiSpendingCaps_ModifyPlan_CreateIsAllowedWhenNothingExistingChanges(t *testing.T) {
	h := newCapsHarness(t)
	h.api.Seed("USER", map[string]int{"set-in-the-ui": 0})
	h.api.Defaults = map[string]int{"USER": 1000}

	assert.False(t, h.planCreate(config{apiTokens: map[string]int64{"token-a": 1}}).Diagnostics.HasError(),
		"what the configuration leaves out never conflicts")
	assert.False(t, h.planCreate(config{users: map[string]int64{"set-in-the-ui": 0}, defaults: &defaults{user: ptr(1000)}}).Diagnostics.HasError(),
		"caps that match the configuration, as after a destroy, do not conflict")
	assert.False(t, newCapsHarness(t).planCreate(config{defaults: &defaults{}, users: map[string]int64{}}).Diagnostics.HasError(),
		"an organization without caps")
}

func TestUnit_AiSpendingCaps_ModifyPlan_CreateListsTheFirstConflictsAndCountsTheRest(t *testing.T) {
	h := newCapsHarness(t)
	for i := 0; i < 15; i++ {
		h.api.Caps["USER"][fmt.Sprintf("user-%02d", i)] = i
	}

	resp := h.planCreate(config{users: map[string]int64{}})

	require.True(t, resp.Diagnostics.HasError())
	detail := resp.Diagnostics.Errors()[0].Detail()
	assert.Contains(t, detail, "user-09")
	assert.NotContains(t, detail, "user-10")
	assert.Contains(t, detail, "and 5 more")
}

func TestUnit_AiSpendingCaps_ModifyPlan_UpdateReadsNothingAndDestroyWarns(t *testing.T) {
	h := newCapsHarness(t)
	h.api.Seed("USER", map[string]int{"set-in-the-ui": 0})
	state := h.refresh()
	callsBefore := len(h.api.CallsMade())

	c := config{users: map[string]int64{}}
	update := &resource.ModifyPlanResponse{Plan: h.tfPlan(planFor(c, &state))}
	h.r.ModifyPlan(context.Background(), resource.ModifyPlanRequest{
		Config: h.tfConfig(c), Plan: h.tfPlan(planFor(c, &state)), State: h.tfState(state),
	}, update)
	assert.False(t, update.Diagnostics.HasError())
	assert.Len(t, h.api.CallsMade(), callsBefore, "an update plans from the refreshed state")

	destroy := &resource.ModifyPlanResponse{Plan: tfsdk.Plan{Schema: h.schema, Raw: h.null()}}
	h.r.ModifyPlan(context.Background(), resource.ModifyPlanRequest{
		Config: tfsdk.Config{Schema: h.schema, Raw: h.null()}, Plan: tfsdk.Plan{Schema: h.schema, Raw: h.null()}, State: h.tfState(state),
	}, destroy)
	require.Equal(t, 1, destroy.Diagnostics.WarningsCount(), "the destroy plan says nothing will be removed")
	assert.Contains(t, destroy.Diagnostics.Warnings()[0].Detail(), "{}")
}

func TestUnit_AiSpendingCaps_Delete_ChangesNothingAndSaysSo(t *testing.T) {
	h := newCapsHarness(t)
	h.api.Seed("USER", map[string]int{"a": 100})
	h.api.Defaults = map[string]int{"USER": 5000, "SERVICEKEY": 2500}
	state := h.refresh()
	callsBefore := len(h.api.CallsMade())

	resp := h.destroy(state)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Equal(t, map[string]int{"a": 100}, h.api.Caps["USER"])
	assert.Equal(t, map[string]int{"USER": 5000, "SERVICEKEY": 2500}, h.api.Defaults)
	assert.Len(t, h.api.CallsMade(), callsBefore, "it sends nothing")
	require.Equal(t, 1, resp.Diagnostics.WarningsCount())
	assert.Contains(t, resp.Diagnostics.Warnings()[0].Detail(), "{}")
}

func TestUnit_AiSpendingCaps_ImportReadsEverything(t *testing.T) {
	h := newCapsHarness(t)
	h.api.Seed("USER", map[string]int{"user-a": 100})
	h.api.Defaults["SERVICEKEY"] = 2500

	resp := &resource.ImportStateResponse{State: tfsdk.State{Schema: h.schema, Raw: h.null()}}
	h.r.ImportState(context.Background(), resource.ImportStateRequest{ID: "org"}, resp)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	got := h.model(resp.State)
	assert.Equal(t, "org", got.Id.ValueString())
	d := defaultsIn(t, got.DefaultCapCents)
	assert.True(t, d.User.IsNull(), "no default is no default")
	assert.Equal(t, int64(2500), d.ApiToken.ValueInt64())
	assert.Equal(t, map[string]int64{"user-a": 100}, elements(t, got.UserCapCents))
	assert.Empty(t, elements(t, got.ApiTokenCapCents))
}

func TestUnit_AiSpendingCaps_ValidateConfig_RefusesMoreThanTheOrganizationCanHave(t *testing.T) {
	h := newCapsHarness(t)
	validate := func(users, tokens int) *resource.ValidateConfigResponse {
		c := config{users: map[string]int64{}, apiTokens: map[string]int64{}}
		for i := 0; i < users; i++ {
			c.users[fmt.Sprintf("u%d", i)] = 1
		}
		for i := 0; i < tokens; i++ {
			c.apiTokens[fmt.Sprintf("t%d", i)] = 1
		}
		resp := &resource.ValidateConfigResponse{}
		h.r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{Config: h.tfConfig(c)}, resp)
		return resp
	}

	assert.False(t, validate(600, 400).Diagnostics.HasError(), "1,000 is allowed")
	over := validate(600, 401)
	require.True(t, over.Diagnostics.HasError())
	assert.Contains(t, over.Diagnostics.Errors()[0].Detail(), "1001")
}

func TestUnit_diffCaps(t *testing.T) {
	changes := diffCaps(map[string]int64{"a": 1, "b": 2, "c": 3}, map[string]int64{"a": 1, "b": 9, "d": 4, "e": 5})
	assert.Equal(t, map[string]int64{"b": 2, "c": 3}, changes.set)
	assert.Equal(t, []string{"d", "e"}, changes.remove)

	assert.Equal(t, []string{"a"}, diffCaps(map[string]int64{}, map[string]int64{"a": 1}).remove, "an empty map removes what the state holds")
	assert.Empty(t, diffCaps(map[string]int64{"a": 1}, map[string]int64{"a": 1}).set, "an unchanged cap is not sent")
	assert.Equal(t, map[string]int64{"a": 0}, diffCaps(map[string]int64{"a": 0}, nil).set, "a cap of 0 is a cap to set")
}

func TestUnit_diffDefault(t *testing.T) {
	assert.Equal(t, 5, *diffDefault(types.Int64Value(5), types.Int64Null()), "a new default is set")
	assert.Equal(t, 0, *diffDefault(types.Int64Value(0), types.Int64Value(5)), "0 is a default to set")
	assert.Nil(t, diffDefault(types.Int64Value(5), types.Int64Value(5)), "an unchanged default is not sent")
	assert.Nil(t, diffDefault(types.Int64Null(), types.Int64Value(5)), "removing is not setting")
	assert.Nil(t, diffDefault(types.Int64Unknown(), types.Int64Value(5)))
}
