package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func listToStrings(ctx context.Context, l types.List) []string {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	var result []string
	l.ElementsAs(ctx, &result, false)
	return result
}

func stringsToList(ctx context.Context, ss []string) types.List {
	if len(ss) == 0 {
		return types.ListNull(types.StringType)
	}
	elems := make([]types.String, len(ss))
	for i, s := range ss {
		elems[i] = types.StringValue(s)
	}
	list, _ := types.ListValueFrom(ctx, types.StringType, elems)
	return list
}
