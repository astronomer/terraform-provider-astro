package schemas_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lucsky/cuid"
	"github.com/stretchr/testify/assert"

	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
)

func validateCapInt64(t *testing.T, attribute string, value int64) bool {
	t.Helper()
	attr := schemas.AiSpendingCapsResourceSchemaAttributes()[attribute].(resourceSchema.Int64Attribute)
	resp := &validator.Int64Response{}
	for _, v := range attr.Validators {
		v.ValidateInt64(context.Background(), validator.Int64Request{Path: path.Root(attribute), ConfigValue: types.Int64Value(value)}, resp)
	}
	return !resp.Diagnostics.HasError()
}

func validateCapMap(t *testing.T, attribute string, caps map[string]int64) bool {
	t.Helper()
	attr := schemas.AiSpendingCapsResourceSchemaAttributes()[attribute].(resourceSchema.MapAttribute)
	value, diags := types.MapValueFrom(context.Background(), types.Int64Type, caps)
	assert.False(t, diags.HasError())
	resp := &validator.MapResponse{}
	for _, v := range attr.Validators {
		v.ValidateMap(context.Background(), validator.MapRequest{Path: path.Root(attribute), ConfigValue: value}, resp)
	}
	return !resp.Diagnostics.HasError()
}

func TestUnit_AiSpendingCapsSchema_DefaultsAreBetweenZeroAndTheMaximum(t *testing.T) {
	for _, attribute := range []string{"default_user_cap_cents", "default_api_token_cap_cents"} {
		assert.True(t, validateCapInt64(t, attribute, 0), "%s: 0 turns AI off and is allowed", attribute)
		assert.True(t, validateCapInt64(t, attribute, schemas.MaxAiSpendingCapCents), "%s: the maximum is allowed", attribute)
		assert.False(t, validateCapInt64(t, attribute, -1), "%s: a negative cap is refused", attribute)
		assert.False(t, validateCapInt64(t, attribute, schemas.MaxAiSpendingCapCents+1), "%s: above the maximum is refused", attribute)
	}
}

func TestUnit_AiSpendingCapsSchema_CapMapsKeepToCuidsAndTheCapRange(t *testing.T) {
	for _, attribute := range []string{"user_cap_cents", "api_token_cap_cents"} {
		assert.True(t, validateCapMap(t, attribute, map[string]int64{}), "%s: an empty map is allowed, it removes every override", attribute)
		assert.True(t, validateCapMap(t, attribute, map[string]int64{cuid.New(): 0, cuid.New(): schemas.MaxAiSpendingCapCents}), "%s: valid IDs and caps", attribute)
		assert.False(t, validateCapMap(t, attribute, map[string]int64{"not-a-cuid": 100}), "%s: a key that is not a cuid is refused", attribute)
		assert.False(t, validateCapMap(t, attribute, map[string]int64{cuid.New(): -1}), "%s: a negative cap is refused", attribute)
		assert.False(t, validateCapMap(t, attribute, map[string]int64{cuid.New(): schemas.MaxAiSpendingCapCents + 1}), "%s: a cap above the maximum is refused", attribute)
	}
}

func TestUnit_AiSpendingCapsSchema_CapMapsHoldAtMostTheOrganizationLimit(t *testing.T) {
	caps := func(n int) map[string]int64 {
		m := make(map[string]int64, n)
		for i := 0; i < n; i++ {
			m[cuid.New()] = int64(i)
		}
		return m
	}
	for _, attribute := range []string{"user_cap_cents", "api_token_cap_cents"} {
		assert.True(t, validateCapMap(t, attribute, caps(schemas.MaxAiSpendingCapSubjects)), fmt.Sprintf("%s: %d entries are allowed", attribute, schemas.MaxAiSpendingCapSubjects))
		assert.False(t, validateCapMap(t, attribute, caps(schemas.MaxAiSpendingCapSubjects+1)), fmt.Sprintf("%s: %d entries are refused", attribute, schemas.MaxAiSpendingCapSubjects+1))
	}
}
