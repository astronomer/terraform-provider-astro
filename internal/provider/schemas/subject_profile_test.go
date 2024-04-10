package schemas_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/astronomer/astronomer-terraform-provider/internal/provider/schemas"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients/iam"
	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

func TestSubjectProfileTypesObject(t *testing.T) {
	ctx := context.Background()

	t.Run("should fail if type passed in is not BasicSubjectProfile", func(t *testing.T) {
		tests := []struct {
			name  string
			input any
		}{
			{"nil", nil},
			{"string", "string"},
			{"int", 1},
			{"bool", true},
			{"map[string]interface{}", map[string]interface{}{"Id": "id"}},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				_, diags := models.SubjectProfileTypesObject(ctx, tc.input)
				assert.True(t, diags.HasError())
				assert.Contains(t, diags[0].Detail(), "SubjectProfileTypesObject expects a BasicSubjectProfile type but did not receive one")
			})
		}
	})

	t.Run("should return subject profile model", func(t *testing.T) {
		tests := []struct {
			name     string
			input    any
			expected models.SubjectProfile
		}{
			{
				"user - &platform.BasicSubjectProfile",
				&platform.BasicSubjectProfile{
					AvatarUrl:   lo.ToPtr("avatar_url"),
					FullName:    lo.ToPtr("full_name"),
					Id:          "id",
					SubjectType: (*platform.BasicSubjectProfileSubjectType)(lo.ToPtr("USER")),
					Username:    lo.ToPtr("username"),
				},
				models.SubjectProfile{
					Id:           types.StringValue("id"),
					SubjectType:  types.StringValue("USER"),
					Username:     types.StringValue("username"),
					FullName:     types.StringValue("full_name"),
					AvatarUrl:    types.StringValue("avatar_url"),
					ApiTokenName: types.StringNull(),
				},
			},
			{
				"token - &iam.BasicSubjectProfile",
				&iam.BasicSubjectProfile{
					Id:           "id",
					SubjectType:  (*iam.BasicSubjectProfileSubjectType)(lo.ToPtr("SERVICEKEY")),
					ApiTokenName: lo.ToPtr("api_token_name"),
				},
				models.SubjectProfile{
					Id:           types.StringValue("id"),
					SubjectType:  types.StringValue("SERVICEKEY"),
					Username:     types.StringNull(),
					FullName:     types.StringNull(),
					AvatarUrl:    types.StringNull(),
					ApiTokenName: types.StringValue("api_token_name"),
				},
			},
			{
				"just id - platform.BasicSubjectProfile",
				platform.BasicSubjectProfile{
					Id: "id",
				},
				models.SubjectProfile{
					Id:           types.StringValue("id"),
					SubjectType:  types.StringNull(),
					Username:     types.StringNull(),
					FullName:     types.StringNull(),
					AvatarUrl:    types.StringNull(),
					ApiTokenName: types.StringNull(),
				},
			},
			{
				"just id - iam.BasicSubjectProfile",
				iam.BasicSubjectProfile{Id: "id"},
				models.SubjectProfile{
					Id:           types.StringValue("id"),
					SubjectType:  types.StringNull(),
					Username:     types.StringNull(),
					FullName:     types.StringNull(),
					AvatarUrl:    types.StringNull(),
					ApiTokenName: types.StringNull(),
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				subjectProfileModel, diags := models.SubjectProfileTypesObject(ctx, tc.input)
				assert.False(t, diags.HasError())

				expectedSubjectProfile, diags := types.ObjectValueFrom(ctx, schemas.SubjectProfileAttributeTypes(), tc.expected)
				assert.False(t, diags.HasError())

				assert.Equal(t, expectedSubjectProfile, subjectProfileModel)
			})
		}
	})
}
