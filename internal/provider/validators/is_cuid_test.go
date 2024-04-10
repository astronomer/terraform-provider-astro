package validators_test

import (
	"testing"

	"github.com/astronomer/astronomer-terraform-provider/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lucsky/cuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestUnit_Validators_ListIsCuids(t *testing.T) {
	type testCase struct {
		list                []string
		expectedListIsCuids bool
	}
	testCases := []testCase{
		{list: []string{cuid.New(), cuid.New()}, expectedListIsCuids: true},
		{list: []string{cuid.New(), "abcdef"}, expectedListIsCuids: false},
		{list: []string{"abc!@#"}, expectedListIsCuids: false},
		{list: []string{"12345"}, expectedListIsCuids: false},
		{list: []string{"c123"}, expectedListIsCuids: false},
	}
	for _, tc := range testCases {
		t.Run("validate cuid", func(t *testing.T) {
			listIsCuidsValidator := validators.ListIsCuids()
			values := lo.Map(tc.list, func(v string, _ int) attr.Value {
				return types.StringValue(v)
			})
			request := validator.ListRequest{
				ConfigValue: types.ListValueMust(types.StringType, values),
			}
			response := validator.ListResponse{}
			listIsCuidsValidator.ValidateList(nil, request, &response)
			assert.Equal(t, response.Diagnostics.HasError(), !tc.expectedListIsCuids)
		})
	}
}

func TestUnit_Validators_IsCuid(t *testing.T) {
	type testCase struct {
		str            string
		expectedIsCuid bool
	}
	testCases := []testCase{
		{str: cuid.New(), expectedIsCuid: true},
		{str: "abcdef", expectedIsCuid: false},
		{str: "abc!@#", expectedIsCuid: false},
		{str: "12345", expectedIsCuid: false},
		{str: "c123", expectedIsCuid: false},
	}
	for _, tc := range testCases {
		t.Run("validate cuid", func(t *testing.T) {
			isCuidValidator := validators.IsCuid()
			request := validator.StringRequest{
				ConfigValue: types.StringValue(tc.str),
			}
			response := validator.StringResponse{}
			isCuidValidator.ValidateString(nil, request, &response)
			assert.Equal(t, response.Diagnostics.HasError(), !tc.expectedIsCuid)
		})
	}
}
