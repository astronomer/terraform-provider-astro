package utils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

func StringList(values []string) (types.List, diag.Diagnostics) {
	list, diags := types.ListValue(types.StringType, lo.Map(values, func(v string, _ int) attr.Value {
		return types.StringValue(v)
	}))
	if diags.HasError() {
		return types.List{}, diags
	}
	return list, nil
}

func ObjectList[T any](ctx context.Context, values []T, objectAttributeTypes map[string]attr.Type, transformer func(context.Context, T) (types.Object, diag.Diagnostics)) (types.List, diag.Diagnostics) {
	if len(values) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: objectAttributeTypes}), nil
	}
	objs := make([]attr.Value, len(values))
	for i, envVar := range values {
		obj, diags := transformer(ctx, envVar)
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
