package utils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

// StringList is a helper that creates a types.List of string values
func StringList(values *[]string) (types.List, diag.Diagnostics) {
	if values == nil {
		return types.ListValue(types.StringType, []attr.Value{})
	}
	return types.ListValue(types.StringType, lo.Map(*values, func(v string, _ int) attr.Value {
		return types.StringValue(v)
	}))
}

// ObjectList is a helper that creates a types.List of objects where each types.Object is created by the transformer function
func ObjectList[T any](ctx context.Context, values *[]T, objectAttributeTypes map[string]attr.Type, transformer func(context.Context, T) (types.Object, diag.Diagnostics)) (types.List, diag.Diagnostics) {
	if values == nil {
		// NullList and EmptyList are different in Terraform
		// Sometimes the API returns a null list, sometimes it returns an empty list
		// However, in the Terraform framework, sometimes we need to return a null list, sometimes we need to return an empty list
		// so there are four possible combinations we need to be aware of
		return types.ListNull(types.ObjectType{AttrTypes: objectAttributeTypes}), nil
	}
	objs := make([]attr.Value, len(*values))
	for i, value := range *values {
		obj, diags := transformer(ctx, value)
		if diags.HasError() {
			return types.List{}, diags
		}
		objs[i] = obj
	}
	return types.ListValue(types.ObjectType{AttrTypes: objectAttributeTypes}, objs)
}

// TypesListToStringSlicePtr converts a types.List to a pointer to a slice of strings
// This is useful for converting a list of strings from the Terraform framework to a slice of strings used for calling the API
// We prefer to use a pointer to a slice of strings because our API client query params usually have type *[]string
// and we can easily assign the query param to the result of this function (regardless if the result is nil or not)
func TypesListToStringSlicePtr(ctx context.Context, list types.List) (*[]string, diag.Diagnostics) {
	if len(list.Elements()) == 0 {
		return nil, nil
	}
	var typesStringSlice []types.String
	diags := list.ElementsAs(ctx, &typesStringSlice, false)
	if diags.HasError() {
		return nil, diags
	}
	resp := lo.Map(typesStringSlice, func(v types.String, _ int) string {
		return v.ValueString()
	})
	return &resp, nil
}
