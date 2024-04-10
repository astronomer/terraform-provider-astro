package schemas_test

import (
	"context"

	"github.com/astronomer/astronomer-terraform-provider/internal/provider/schemas"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients/iam"
	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("Common Test", func() {
	var ctx context.Context
	BeforeEach(func() {
		ctx = context.Background()
	})
	Context("SubjectProfileTypesObject", func() {
		DescribeTable("should fail if type passed in is not BasicSubjectProfile", func(value any) {
			_, diags := models.SubjectProfileTypesObject(ctx, value)
			Expect(diags.HasError()).To(BeTrue())
			Expect(
				diags[0].Detail(),
			).To(Equal("SubjectProfileTypesObject expects a BasicSubjectProfile type but did not receive one"))
		},
			Entry("nil", nil),
			Entry("string", "string"),
			Entry("int", 1),
			Entry("bool", true),
			Entry("map[string]interface{}",
				map[string]interface{}{"Id": "id"}),
		)

		DescribeTable(
			"should return subject profile model",
			func(input any, expected models.SubjectProfile) {
				subjectProfileModel, diags := models.SubjectProfileTypesObject(ctx, input)
				Expect(diags.HasError()).To(BeFalse())
				expectedSubjectProfile, diags := types.ObjectValueFrom(ctx, schemas.SubjectProfileAttributeTypes(), expected)
				Expect(diags.HasError()).To(BeFalse())
				Expect(subjectProfileModel).To(Equal(expectedSubjectProfile))
			},
			Entry("user", &platform.BasicSubjectProfile{
				AvatarUrl:   lo.ToPtr("avatar_url"),
				FullName:    lo.ToPtr("full_name"),
				Id:          "id",
				SubjectType: (*platform.BasicSubjectProfileSubjectType)(lo.ToPtr("USER")),
				Username:    lo.ToPtr("username"),
			}, models.SubjectProfile{
				Id:           types.StringValue("id"),
				SubjectType:  types.StringValue("USER"),
				Username:     types.StringValue("username"),
				FullName:     types.StringValue("full_name"),
				AvatarUrl:    types.StringValue("avatar_url"),
				ApiTokenName: types.StringNull(),
			}),
			Entry("token", &iam.BasicSubjectProfile{
				Id:           "id",
				SubjectType:  (*iam.BasicSubjectProfileSubjectType)(lo.ToPtr("SERVICEKEY")),
				ApiTokenName: lo.ToPtr("api_token_name"),
			}, models.SubjectProfile{
				Id:           types.StringValue("id"),
				SubjectType:  types.StringValue("SERVICEKEY"),
				Username:     types.StringNull(),
				FullName:     types.StringNull(),
				AvatarUrl:    types.StringNull(),
				ApiTokenName: types.StringValue("api_token_name"),
			}),
			Entry("just id", &platform.BasicSubjectProfile{
				Id: "id",
			}, models.SubjectProfile{
				Id:           types.StringValue("id"),
				SubjectType:  types.StringNull(),
				Username:     types.StringNull(),
				FullName:     types.StringNull(),
				AvatarUrl:    types.StringNull(),
				ApiTokenName: types.StringNull(),
			}),
			Entry("platform.BasicSubjectProfile", platform.BasicSubjectProfile{
				Id: "id",
			}, models.SubjectProfile{
				Id:           types.StringValue("id"),
				SubjectType:  types.StringNull(),
				Username:     types.StringNull(),
				FullName:     types.StringNull(),
				AvatarUrl:    types.StringNull(),
				ApiTokenName: types.StringNull(),
			}),
			Entry("*platform.BasicSubjectProfile", &platform.BasicSubjectProfile{
				Id: "id",
			}, models.SubjectProfile{
				Id:           types.StringValue("id"),
				SubjectType:  types.StringNull(),
				Username:     types.StringNull(),
				FullName:     types.StringNull(),
				AvatarUrl:    types.StringNull(),
				ApiTokenName: types.StringNull(),
			}),
			Entry("iam.BasicSubjectProfile", iam.BasicSubjectProfile{
				Id: "id",
			}, models.SubjectProfile{
				Id:           types.StringValue("id"),
				SubjectType:  types.StringNull(),
				Username:     types.StringNull(),
				FullName:     types.StringNull(),
				AvatarUrl:    types.StringNull(),
				ApiTokenName: types.StringNull(),
			}),
			Entry("*iam.BasicSubjectProfile", &iam.BasicSubjectProfile{
				Id: "id",
			}, models.SubjectProfile{
				Id:           types.StringValue("id"),
				SubjectType:  types.StringNull(),
				Username:     types.StringNull(),
				FullName:     types.StringNull(),
				AvatarUrl:    types.StringNull(),
				ApiTokenName: types.StringNull(),
			}),
		)
	})
})
