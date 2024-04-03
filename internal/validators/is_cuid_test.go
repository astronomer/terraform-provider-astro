package validators_test

import (
	"github.com/astronomer/astronomer-terraform-provider/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lucsky/cuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IsCuid Validator Test", func() {
	var isCuidValidator validator.String
	BeforeEach(func() {
		isCuidValidator = validators.IsCuid()
	})

	DescribeTable("validate cuid", func(str string, isCuid bool) {
		var request validator.StringRequest
		request.ConfigValue = types.StringValue(str)
		var response validator.StringResponse
		isCuidValidator.ValidateString(nil, request, &response)
		Expect(response.Diagnostics.HasError()).To(Equal(!isCuid))
	},
		Entry("cuid", cuid.New(), true),
		Entry("invalid", "abcdef", false),
		Entry("invalid", "abc!@#", false),
		Entry("invalid", "12345", false),
		Entry("invalid", "c123", false),
	)
})
