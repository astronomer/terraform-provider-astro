package utils

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

// StringList is a helper that creates a types.List of string values
func StringList(values []string) (types.List, diag.Diagnostics) {
	list, diags := types.ListValue(types.StringType, lo.Map(values, func(v string, _ int) attr.Value {
		return types.StringValue(v)
	}))
	if diags.HasError() {
		return types.List{}, diags
	}
	return list, nil
}

// ObjectList is a helper that creates a types.List of objects where each types.Object is created by the transformer function
func ObjectList[T any](ctx context.Context, values []T, objectAttributeTypes map[string]attr.Type, transformer func(context.Context, T) (types.Object, diag.Diagnostics)) (types.List, diag.Diagnostics) {
	if len(values) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: objectAttributeTypes}), nil
	}
	objs := make([]attr.Value, len(values))
	for i, value := range values {
		obj, diags := transformer(ctx, value)
		if diags.HasError() {
			return types.List{}, diags
		}
		objs[i] = obj
	}
	objectList, diags := types.ListValue(types.ObjectType{AttrTypes: objectAttributeTypes}, objs)
	if diags.HasError() {
		return types.List{}, diags
	}
	return objectList, nil
}

// TypesListToStringSlice converts a types.List to a pointer to a slice of strings
// This is useful for converting a list of strings from the Terraform framework to a slice of strings used for calling the API
func TypesListToStringSlice(list types.List) *[]string {
	elements := list.Elements()
	if len(elements) == 0 {
		return nil
	}
	slice := lo.Map(elements, func(id attr.Value, _ int) string {
		// Terraform includes quotes around the string, so we need to remove them
		return strings.ReplaceAll(id.String(), `"`, "")
	})
	return &slice
}
