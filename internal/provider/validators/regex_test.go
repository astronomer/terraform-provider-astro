package validators_test

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"regexp"
	"testing"

	"github.com/astronomer/astronomer-terraform-provider/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestUnit_Validators_Regex(t *testing.T) {
	t.Run("validate email", func(t *testing.T) {
		type testCase struct {
			str             string
			expectedIsValid bool
		}
		testCases := []testCase{
			{str: "email@email.com", expectedIsValid: true},
			{str: "invalidemail", expectedIsValid: false},
			{str: "", expectedIsValid: false},
		}
		for _, tc := range testCases {
			t.Run("validate email", func(t *testing.T) {
				v := stringvalidator.RegexMatches(regexp.MustCompile(validators.EmailString), tc.str)
				request := validator.StringRequest{
					ConfigValue: types.StringValue(tc.str),
				}
				response := validator.StringResponse{}
				v.ValidateString(nil, request, &response)
				assert.Equal(t, response.Diagnostics.HasError(), !tc.expectedIsValid)
			})
		}
	})

	t.Run("validate kubernetes resource", func(t *testing.T) {
		type testCase struct {
			str             string
			expectedIsValid bool
		}
		testCases := []testCase{
			{str: "0.5Gi", expectedIsValid: true},
			{str: "2", expectedIsValid: true},
			{str: "", expectedIsValid: false},
			{str: "abc", expectedIsValid: false},
		}
		for _, tc := range testCases {
			t.Run("validate kubernetes resource string", func(t *testing.T) {
				v := stringvalidator.RegexMatches(regexp.MustCompile(validators.KubernetesResourceString), tc.str)
				request := validator.StringRequest{
					ConfigValue: types.StringValue(tc.str),
				}
				response := validator.StringResponse{}
				v.ValidateString(nil, request, &response)
				assert.Equal(t, response.Diagnostics.HasError(), !tc.expectedIsValid)
			})
		}
	})
}
