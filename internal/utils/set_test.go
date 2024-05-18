package utils_test

import (
	"context"
	"fmt"
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
