package validators_test

import (
	"fmt"
	"testing"

	"github.com/astronomer/astronomer-terraform-provider/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lucsky/cuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_Validators_IsCuid(t *testing.T) {
	type testCase struct {
		str            string
		expectedIsCuid bool
	}
	testCases := []testCase{
		{str: "null", expectedIsCuid: true},
		{str: "unknown", expectedIsCuid: true},
		{str: cuid.New(), expectedIsCuid: true},
		{str: "abcdef", expectedIsCuid: false},
		{str: "abc!@#", expectedIsCuid: false},
		{str: "12345", expectedIsCuid: false},
		{str: "c123", expectedIsCuid: false},
		{str: "", expectedIsCuid: false},
	}
	for _, tc := range testCases {
		t.Run("validate cuid", func(t *testing.T) {
			isCuidValidator := validators.IsCuid()
			request := validator.StringRequest{
				ConfigValue: types.StringValue(tc.str),
			}
			if tc.str == "null" {
				request.ConfigValue = types.StringNull()
			}
			if tc.str == "unknown" {
				request.ConfigValue = types.StringUnknown()
			}
			response := validator.StringResponse{}
			isCuidValidator.ValidateString(nil, request, &response)
			assert.Equal(t, response.Diagnostics.HasError(), !tc.expectedIsCuid, fmt.Sprintf("test case: %s failed", tc.str))
		})
	}
}
