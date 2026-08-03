package validators_test

import (
	"fmt"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestUnit_Validators_IsCidr(t *testing.T) {
	type testCase struct {
		str            string
		expectedIsCidr bool
	}
	testCases := []testCase{
		{str: "null", expectedIsCidr: true},
		{str: "unknown", expectedIsCidr: true},
		{str: "203.0.113.0/24", expectedIsCidr: true},
		{str: "198.51.100.5/32", expectedIsCidr: true},
		{str: "10.1.2.3/8", expectedIsCidr: true}, // host bits set is still parseable
		{str: "2001:db8::/64", expectedIsCidr: true},
		{str: "203.0.113.0", expectedIsCidr: false},    // missing prefix
		{str: "203.0.113.0/33", expectedIsCidr: false}, // prefix out of range
		{str: "not-a-cidr", expectedIsCidr: false},
		{str: "", expectedIsCidr: false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("validate cidr %q", tc.str), func(t *testing.T) {
			isCidrValidator := validators.IsCidr()
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
			isCidrValidator.ValidateString(nil, request, &response)
			assert.Equal(t, !tc.expectedIsCidr, response.Diagnostics.HasError(), fmt.Sprintf("test case: %s failed", tc.str))
		})
	}
}
