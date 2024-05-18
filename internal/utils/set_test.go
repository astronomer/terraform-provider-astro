package utils_test

import (
	"context"
	"fmt"
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/lucsky/cuid"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/astronomer/terraform-provider-astro/internal/utils"
)

func TestUnit_StringSet(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		s, diags := utils.StringSet(nil)

		assert.False(t, diags.HasError())
		assert.Equal(t, 0, len(s.Elements()))
	})

	t.Run("with values", func(t *testing.T) {
		input := []string{"one", "two", "three"}

		s, diags := utils.StringSet(&input)

		assert.False(t, diags.HasError())
		assert.Equal(t, len(input), len(s.Elements()))
		for i, v := range input {
			assert.Equal(t, fmt.Sprintf(`"%s"`, v), s.Elements()[i].String())
		}
	})
}

func TestUnit_TypesSetToStringSlice(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		s := types.SetValueMust(types.StringType, []attr.Value{})

		result, diags := utils.TypesSetToStringSlice(context.Background(), s)
		assert.Nil(t, diags)
		assert.Empty(t, result)
	})

	t.Run("with values", func(t *testing.T) {
		s := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("string1"), types.StringValue("string2")})

		expected := []string{"string1", "string2"}
		result, diags := utils.TypesSetToStringSlice(context.Background(), s)
		assert.Nil(t, diags)
		assert.Equal(t, expected, result)
	})
}

func TestUnit_TypesSetToObjectSlice(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		s := types.SetNull(types.ObjectType{AttrTypes: schemas.WorkspaceRoleAttributeTypes()})
		result, diags := utils.TypesSetToObjectSlice[iam.WorkspaceRole](context.Background(), s)
		assert.Nil(t, diags)
		assert.Empty(t, result)
	})

	t.Run("with values", func(t *testing.T) {
		workspaceId := cuid.New()
		workspaceRole := iam.WORKSPACEOWNER
		s, diags := utils.ObjectSet(context.Background(), &[]iam.WorkspaceRole{{WorkspaceId: workspaceId, Role: workspaceRole}}, schemas.WorkspaceRoleAttributeTypes(), models.WorkspaceRoleTypesObject)
		assert.Nil(t, diags)
		result, diags := utils.TypesSetToObjectSlice[models.WorkspaceRole](context.Background(), s)
		expected := []models.WorkspaceRole{{WorkspaceId: types.StringValue(workspaceId), Role: types.StringValue(string(workspaceRole))}}
		assert.Nil(t, diags)
		assert.Equal(t, expected, result)
	})
}
