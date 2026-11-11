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

// capsHarness drives the resource's methods against the fake, the way Terraform does.
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

// desired builds the model for a configuration: a nil map or default is an attribute left out.
func desired(defaultUser, defaultApiToken *int64, users, apiTokens map[string]int64) models.AiSpendingCapsResource {
	return models.AiSpendingCapsResource{
		Id:                      types.StringUnknown(),
		DefaultUserCapCents:     int64Value(defaultUser),
		DefaultApiTokenCapCents: int64Value(defaultApiToken),
		UserCapCents:            capsMapValue(users),
		ApiTokenCapCents:        capsMapValue(apiTokens),
	}
}

func (h *capsHarness) null() tftypes.Value { return tftypes.NewValue(h.objType, nil) }

func (h *capsHarness) plan(m models.AiSpendingCapsResource) tfsdk.Plan {
	p := tfsdk.Plan{Schema: h.schema, Raw: h.null()}
	require.False(h.t, p.Set(context.Background(), &m).HasError())
	return p
}

func (h *capsHarness) state(m models.AiSpendingCapsResource) tfsdk.State {
	m.Id = types.StringValue("org")
	st := tfsdk.State{Schema: h.schema, Raw: h.null()}
	require.False(h.t, st.Set(context.Background(), &m).HasError())
	return st
}

func (h *capsHarness) model(st tfsdk.State) models.AiSpendingCapsResource {
	var m models.AiSpendingCapsResource
	require.False(h.t, st.Get(context.Background(), &m).HasError())
	return m
}

func (h *capsHarness) create(m models.AiSpendingCapsResource) *resource.CreateResponse {
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: h.schema, Raw: h.null()}}
	h.r.Create(context.Background(), resource.CreateRequest{Plan: h.plan(m)}, resp)
	return resp
}

func (h *capsHarness) update(plan, state models.AiSpendingCapsResource) *resource.UpdateResponse {
	prior := h.state(state)
	resp := &resource.UpdateResponse{State: prior}
	h.r.Update(context.Background(), resource.UpdateRequest{Plan: h.plan(plan), State: prior}, resp)
	return resp
}

func (h *capsHarness) read(state models.AiSpendingCapsResource) *resource.ReadResponse {
	prior := h.state(state)
	resp := &resource.ReadResponse{State: prior}
	h.r.Read(context.Background(), resource.ReadRequest{State: prior}, resp)
	return resp
}

func (h *capsHarness) destroy(state models.AiSpendingCapsResource) *resource.DeleteResponse {
	prior := h.state(state)
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

func TestUnit_AiSpendingCaps_SchemaIsValid(t *testing.T) {
	s := rschema.Schema{Attributes: schemas.AiSpendingCapsResourceSchemaAttributes()}
	assert.False(t, s.ValidateImplementation(context.Background()).HasError())
}

func TestUnit_AiSpendingCaps_Create_SetsWhatIsConfiguredAndStoresThePlan(t *testing.T) {
	h := newCapsHarness(t)
	plan := desired(ptr(5000), ptr(2500), map[string]int64{"user-a": 100, "user-b": 0}, map[string]int64{"token-a": 300})

	resp := h.create(plan)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Equal(t, map[string]int{"USER": 5000, "SERVICEKEY": 2500}, h.api.Defaults)
	assert.Equal(t, map[string]int{"user-a": 100, "user-b": 0}, h.api.Caps["USER"], "0 is a cap, not an absent one")
	assert.Equal(t, map[string]int{"token-a": 300}, h.api.Caps["SERVICEKEY"])
	stored := h.model(resp.State)
	assert.Equal(t, "org", stored.Id.ValueString())
	assert.Equal(t, int64(5000), stored.DefaultUserCapCents.ValueInt64())
	assert.Equal(t, map[string]int64{"user-a": 100, "user-b": 0}, elements(t, stored.UserCapCents))
}

func TestUnit_AiSpendingCaps_Create_LeavesUnmanagedAttributesAlone(t *testing.T) {
	h := newCapsHarness(t)
	h.api.Seed("USER", map[string]int{"set-in-the-ui": 700})
	h.api.Defaults["SERVICEKEY"] = 99

	resp := h.create(desired(ptr(5000), nil, nil, nil))

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Equal(t, []string{"set-defaults"}, h.api.CallsMade(), "only the one default that is configured is sent")
	assert.Equal(t, map[string]int{"set-in-the-ui": 700}, h.api.Caps["USER"], "a map that is not set is not managed")
	assert.Equal(t, 99, h.api.Defaults["SERVICEKEY"], "a default that is not set is not managed")
	stored := h.model(resp.State)
	assert.True(t, stored.UserCapCents.IsNull())
	assert.True(t, stored.DefaultApiTokenCapCents.IsNull())
}

func TestUnit_AiSpendingCaps_Create_ChunksTheBulkRequests(t *testing.T) {
	h := newCapsHarness(t)
	users := map[string]int64{}
	for i := 0; i < 1100; i++ {
		users[fmt.Sprintf("user-%04d", i)] = int64(i)
	}

	resp := h.create(desired(nil, nil, users, nil))

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Equal(t, []string{"set-caps:500", "set-caps:500", "set-caps:100"}, h.api.CallsMade())
	assert.Len(t, h.api.Caps["USER"], 1100)
}

func TestUnit_AiSpendingCaps_Update_SendsOnlyWhatChangedAndSetsBeforeRemoving(t *testing.T) {
	h := newCapsHarness(t)
	h.api.Seed("USER", map[string]int{"keep": 100, "change": 200, "drop": 300})
	state := desired(ptr(5000), ptr(2500), map[string]int64{"keep": 100, "change": 200, "drop": 300}, nil)
	plan := desired(ptr(6000), nil, map[string]int64{"keep": 100, "change": 250, "new": 50}, nil)
	h.api.Defaults = map[string]int{"USER": 5000, "SERVICEKEY": 2500}

	resp := h.update(plan, state)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Equal(t, []string{"set-caps:2", "set-defaults", "remove-caps:1", "remove-defaults"}, h.api.CallsMade(),
		"the unchanged cap is not sent, overrides are set before defaults, and everything that sets runs before anything that removes")
	assert.Equal(t, map[string]int{"keep": 100, "change": 250, "new": 50}, h.api.Caps["USER"])
	assert.Equal(t, map[string]int{"USER": 6000}, h.api.Defaults, "the API token default left the configuration, so it is removed")
}

func TestUnit_AiSpendingCaps_Update_RemovingAMapRemovesOnlyWhatTerraformHeld(t *testing.T) {
	h := newCapsHarness(t)
	h.api.Seed("USER", map[string]int{"managed": 100, "set-in-the-ui": 700})
	state := desired(nil, nil, map[string]int64{"managed": 100}, nil)

	resp := h.update(desired(nil, nil, nil, nil), state)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Equal(t, map[string]int{"set-in-the-ui": 700}, h.api.Caps["USER"])
	assert.True(t, h.model(resp.State).UserCapCents.IsNull())
}

func TestUnit_AiSpendingCaps_Update_AnEmptyMapRemovesEveryCapItHolds(t *testing.T) {
	h := newCapsHarness(t)
	h.api.Seed("SERVICEKEY", map[string]int{"token-a": 1, "token-b": 2})
	state := desired(nil, nil, nil, map[string]int64{"token-a": 1, "token-b": 2})

	resp := h.update(desired(nil, nil, nil, map[string]int64{}), state)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Empty(t, h.api.Caps["SERVICEKEY"])
	assert.NotNil(t, elements(t, h.model(resp.State).ApiTokenCapCents), "an empty map stays set")
	assert.False(t, h.model(resp.State).ApiTokenCapCents.IsNull())
}

func TestUnit_AiSpendingCaps_Read_AMapIsAuthoritativeAndAnUnsetOneIsNotRead(t *testing.T) {
	h := newCapsHarness(t)
	h.api.Seed("USER", map[string]int{"managed": 100, "set-in-the-ui": 700})
	h.api.Seed("SERVICEKEY", map[string]int{"token-a": 5})
	h.api.Defaults["USER"] = 4000
	state := desired(ptr(5000), ptr(1000), map[string]int64{"managed": 100}, nil)

	resp := h.read(state)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	got := h.model(resp.State)
	assert.Equal(t, map[string]int64{"managed": 100, "set-in-the-ui": 700}, elements(t, got.UserCapCents), "a cap set in the UI shows as drift")
	assert.True(t, got.ApiTokenCapCents.IsNull(), "a map that is not managed stays null")
	assert.Equal(t, int64(4000), got.DefaultUserCapCents.ValueInt64(), "a default changed behind Terraform's back shows as drift")
	assert.True(t, got.DefaultApiTokenCapCents.IsNull(), "a managed default removed behind Terraform's back reads as unset")
	assert.NotContains(t, h.api.CallsMade(), "list:SERVICEKEY")
}

func TestUnit_AiSpendingCaps_Read_PagesThroughEveryCap(t *testing.T) {
	h := newCapsHarness(t)
	for i := 0; i < 250; i++ {
		h.api.Caps["USER"][fmt.Sprintf("user-%03d", i)] = i
	}

	resp := h.read(desired(nil, nil, map[string]int64{}, nil))

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Len(t, elements(t, h.model(resp.State).UserCapCents), 250)
	assert.Equal(t, []string{"list:USER", "list:USER", "list:USER"}, h.api.CallsMade(), "100 per page")
}

func TestUnit_AiSpendingCaps_Read_ErrorsWhenAPageWentMissing(t *testing.T) {
	h := newCapsHarness(t)
	h.api.Seed("USER", map[string]int{"a": 1})
	h.api.TotalCountPadding = 1

	resp := h.read(desired(nil, nil, map[string]int64{}, nil))

	require.True(t, resp.Diagnostics.HasError(), "a list shorter than its count must not read as the whole list")
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "reports 2")
}

func TestUnit_AiSpendingCaps_AFailedEntryIsAnErrorNamingTheSubject(t *testing.T) {
	h := newCapsHarness(t)
	h.api.FailIds["user-b"] = true

	resp := h.create(desired(nil, nil, map[string]int64{"user-a": 1, "user-b": 2}, nil))

	require.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "user-b")
	assert.NotContains(t, resp.Diagnostics.Errors()[0].Detail(), "user-a")
}

func TestUnit_AiSpendingCaps_AnUnknownSubjectIsReportedAsItselfAndNotAsTheGate(t *testing.T) {
	h := newCapsHarness(t)
	h.api.UnknownIds["gone-token"] = true

	create := h.create(desired(nil, nil, nil, map[string]int64{"gone-token": 100}))
	require.True(t, create.Diagnostics.HasError())
	errs := create.Diagnostics.Errors()
	assert.NotContains(t, errs[0].Summary(), "not available", "a 404 for a subject does not mean the feature is off")
	assert.NotContains(t, errs[0].Detail(), "Contact Astronomer")
	assert.Contains(t, errs[0].Detail(), "api token with id gone-token not found")

	h.api.Seed("USER", map[string]int{"gone-user": 5})
	h.api.UnknownIds["gone-user"] = true
	update := h.update(desired(nil, nil, map[string]int64{}, nil), desired(nil, nil, map[string]int64{"gone-user": 5}, nil))
	require.True(t, update.Diagnostics.HasError())
	assert.NotContains(t, update.Diagnostics.Errors()[0].Summary(), "not available")
	assert.Contains(t, update.Diagnostics.Errors()[0].Detail(), "user with id gone-user not found")
}

func TestUnit_AiSpendingCaps_TheGateBeingOffIsAnErrorAndNeverDropsTheResource(t *testing.T) {
	h := newCapsHarness(t)
	h.api.GateOff = true
	state := desired(ptr(5000), nil, map[string]int64{"user-a": 1}, nil)

	read := h.read(state)
	require.True(t, read.Diagnostics.HasError())
	assert.Contains(t, read.Diagnostics.Errors()[0].Summary(), "not available")
	assert.False(t, read.State.Raw.IsNull(), "a 404 here means the feature is off, not that the resource is gone")

	create := h.create(desired(ptr(5000), nil, nil, nil))
	require.True(t, create.Diagnostics.HasError())
	assert.Contains(t, create.Diagnostics.Errors()[0].Detail(), "Contact Astronomer")

	imported := &resource.ImportStateResponse{State: tfsdk.State{Schema: h.schema, Raw: h.null()}}
	h.r.ImportState(context.Background(), resource.ImportStateRequest{ID: "org"}, imported)
	assert.True(t, imported.Diagnostics.HasError())
}

func TestUnit_AiSpendingCaps_Delete_RemovesWhatTheStateHoldsAndNothingElse(t *testing.T) {
	h := newCapsHarness(t)
	h.api.Seed("USER", map[string]int{"managed": 100, "set-in-the-ui": 700})
	h.api.Seed("SERVICEKEY", map[string]int{"token-a": 5})
	h.api.Defaults = map[string]int{"USER": 5000, "SERVICEKEY": 2500}
	state := desired(ptr(5000), nil, map[string]int64{"managed": 100}, nil)

	resp := h.destroy(state)

	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.Equal(t, map[string]int{"set-in-the-ui": 700}, h.api.Caps["USER"])
	assert.Equal(t, map[string]int{"token-a": 5}, h.api.Caps["SERVICEKEY"], "an unmanaged map is left alone")
	assert.Equal(t, map[string]int{"SERVICEKEY": 2500}, h.api.Defaults, "only the managed default is removed")
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
	assert.True(t, got.DefaultUserCapCents.IsNull(), "no default is no default")
	assert.Equal(t, int64(2500), got.DefaultApiTokenCapCents.ValueInt64())
	assert.Equal(t, map[string]int64{"user-a": 100}, elements(t, got.UserCapCents))
	assert.False(t, got.ApiTokenCapCents.IsNull(), "the maps are set, empty or not, so they are managed after an import")
	assert.Empty(t, elements(t, got.ApiTokenCapCents))
}

func TestUnit_AiSpendingCaps_ValidateConfig_RefusesMoreThanTheOrganizationCanHave(t *testing.T) {
	h := newCapsHarness(t)
	validate := func(users, tokens int) *resource.ValidateConfigResponse {
		userCaps, tokenCaps := map[string]int64{}, map[string]int64{}
		for i := 0; i < users; i++ {
			userCaps[fmt.Sprintf("u%d", i)] = 1
		}
		for i := 0; i < tokens; i++ {
			tokenCaps[fmt.Sprintf("t%d", i)] = 1
		}
		m := desired(nil, nil, userCaps, tokenCaps)
		cfg := tfsdk.Config{Schema: h.schema, Raw: h.plan(m).Raw}
		resp := &resource.ValidateConfigResponse{}
		h.r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{Config: cfg}, resp)
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

	assert.Equal(t, []string{"a"}, diffCaps(nil, map[string]int64{"a": 1}).remove, "a map left out removes what the state held")
	assert.Equal(t, []string{"a"}, diffCaps(map[string]int64{}, map[string]int64{"a": 1}).remove, "an empty map does too")
	assert.Empty(t, diffCaps(map[string]int64{"a": 1}, map[string]int64{"a": 1}).set, "an unchanged cap is not sent")
	assert.Equal(t, map[string]int64{"a": 0}, diffCaps(map[string]int64{"a": 0}, nil).set, "a cap of 0 is a cap to set")
}

func TestUnit_diffDefault(t *testing.T) {
	set, remove := diffDefault(types.Int64Value(5), types.Int64Null())
	assert.Equal(t, 5, *set)
	assert.False(t, remove)

	set, remove = diffDefault(types.Int64Value(0), types.Int64Value(5))
	assert.Equal(t, 0, *set, "0 is a default to set")
	assert.False(t, remove)

	set, remove = diffDefault(types.Int64Value(5), types.Int64Value(5))
	assert.Nil(t, set)
	assert.False(t, remove)

	set, remove = diffDefault(types.Int64Null(), types.Int64Value(5))
	assert.Nil(t, set)
	assert.True(t, remove, "left out of the configuration after it was set")

	set, remove = diffDefault(types.Int64Null(), types.Int64Null())
	assert.Nil(t, set)
	assert.False(t, remove, "never managed")
}
