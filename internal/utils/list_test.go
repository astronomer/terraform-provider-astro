package utils_test

import (
	"context"
	"fmt"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/models"
	"github.com/astronomer/astronomer-terraform-provider/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/astronomer/astronomer-terraform-provider/internal/utils"
)

func TestUnit_StringList(t *testing.T) {
	input := []string{"one", "two", "three"}

	list, diags := utils.StringList(input)

	assert.False(t, diags.HasError())
	assert.Equal(t, len(input), len(list.Elements()))
	for i, v := range input {
		assert.Equal(t, fmt.Sprintf(`"%s"`, v), list.Elements()[i].String())
	}
}

func TestUnit_ObjectList(t *testing.T) {
	ctx := context.Background()
	input := []models.DeploymentEnvironmentVariable{
		{
			Key:       types.StringValue("key1"),
			Value:     types.StringValue("value1"),
			UpdatedAt: types.StringValue("date1"),
			IsSecret:  types.BoolValue(false),
		},
		{
			Key:       types.StringValue("key2"),
			Value:     types.StringValue("value2"),
			UpdatedAt: types.StringValue("date2"),
			IsSecret:  types.BoolValue(true),
		},
	}
	transformer := func(ctx context.Context, value models.DeploymentEnvironmentVariable) (types.Object, diag.Diagnostics) {
		obj, diags := types.ObjectValue(schemas.DeploymentEnvironmentVariableAttributeTypes(), map[string]attr.Value{
			"key":        value.Key,
			"value":      value.Value,
			"updated_at": value.UpdatedAt,
			"is_secret":  value.IsSecret,
		})
		if diags.HasError() {
			return types.Object{}, diags
		}
		return obj, nil
	}
	list, diags := utils.ObjectList(ctx, input, schemas.DeploymentEnvironmentVariableAttributeTypes(), transformer)

	assert.False(t, diags.HasError())
	assert.Equal(t, len(input), len(list.Elements()))
	for i, v := range input {
		objString := list.Elements()[i].String()
		assert.Contains(t, objString, v.Key.String())
		assert.Contains(t, objString, v.Value.String())
		assert.Contains(t, objString, v.UpdatedAt.String())
		assert.Contains(t, objString, v.IsSecret.String())
	}
}

func TestUnit_TypesListToStringSlicePtr(t *testing.T) {
	list := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("string1"), types.StringValue("string2")})

	expected := &[]string{"string1", "string2"}
	result := utils.TypesListToStringSlicePtr(list)
	assert.Equal(t, expected, result)
}
