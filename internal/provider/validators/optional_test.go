package validators_test

import (
	"testing"

	"github.com/astronomer/astronomer-terraform-provider/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lucsky/cuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_Validators_Optional(t *testing.T) {
	type testCase struct {
		str          string
		expectedPass bool
	}
	testCases := []testCase{
		{str: cuid.New(), expectedPass: true},
		{str: "abcdef", expectedPass: false},
		{str: "abc!@#", expectedPass: false},
		{str: "12345", expectedPass: false},
		{str: "c123", expectedPass: false},
		{str: "", expectedPass: true},
	}
	for _, tc := range testCases {
		t.Run("validate optional cuid", func(t *testing.T) {
			optionalIsCuidValidator := validators.OptionalString(validators.IsCuid())
			request := validator.StringRequest{
				ConfigValue: types.StringValue(tc.str),
			}
			response := validator.StringResponse{}
			optionalIsCuidValidator.ValidateString(nil, request, &response)
			assert.Equal(t, response.Diagnostics.HasError(), !tc.expectedPass)
		})
	}
}
