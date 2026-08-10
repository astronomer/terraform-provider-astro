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
		{str: "10.0.0.0/8", expectedIsCidr: true},
		{str: "203.0.113.0/24", expectedIsCidr: true},
		{str: "2001:db8::/32", expectedIsCidr: true},
		// Host bits set: the API accepts and stores these verbatim, so they round-trip and must pass.
		{str: "10.1.2.3/8", expectedIsCidr: true},
		// Surrounding whitespace: the API trims before storing, which would break plan/state
		// consistency, so it must be rejected at plan time.
		{str: " 10.0.0.0/8", expectedIsCidr: false},
		{str: "10.0.0.0/8 ", expectedIsCidr: false},
		// Not CIDR notation / malformed.
		{str: "10.0.0.0", expectedIsCidr: false},
		{str: "not-a-cidr", expectedIsCidr: false},
		{str: "10.0.0.0/33", expectedIsCidr: false},
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
			assert.Equal(t, !tc.expectedIsCidr, response.Diagnostics.HasError(), fmt.Sprintf("test case: %q failed", tc.str))
		})
	}
}
