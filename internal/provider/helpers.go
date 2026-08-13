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

// stringOrNull maps an empty API string to a null attribute so an unset
// optional string does not show a perpetual diff against a null config.
func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func stringsToList(ctx context.Context, ss []string) types.List {
	if ss == nil {
		return types.ListNull(types.StringType)
	}
	elems := make([]types.String, len(ss))
	for i, s := range ss {
		elems[i] = types.StringValue(s)
	}
	list, _ := types.ListValueFrom(ctx, types.StringType, elems)
	return list
}

// preserveListNullEmptySemantics keeps the prior null/empty distinction when
// the API returns null for a field that was previously an empty list.
// Without this, OpenTofu reports "inconsistent result after apply" because
// the plan/state had [] but the provider returned null.
func preserveListNullEmptySemantics(prior, current types.List) types.List {
	if current.IsNull() && !prior.IsNull() && len(prior.Elements()) == 0 {
		return prior
	}
	return current
}
