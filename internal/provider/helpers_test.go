package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestListToStrings_NullList(t *testing.T) {
	ctx := context.Background()
	null := types.ListNull(types.StringType)
	result := listToStrings(ctx, null)
	if result != nil {
		t.Fatalf("expected nil for null list, got %v", result)
	}
}

func TestListToStrings_UnknownList(t *testing.T) {
	ctx := context.Background()
	unknown := types.ListUnknown(types.StringType)
	result := listToStrings(ctx, unknown)
	if result != nil {
		t.Fatalf("expected nil for unknown list, got %v", result)
	}
}

func TestListToStrings_EmptyList(t *testing.T) {
	ctx := context.Background()
	list, diags := types.ListValueFrom(ctx, types.StringType, []types.String{})
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	result := listToStrings(ctx, list)
	if len(result) != 0 {
		t.Fatalf("expected empty slice, got %v", result)
	}
}

func TestListToStrings_WithValues(t *testing.T) {
	ctx := context.Background()
	elems := []types.String{
		types.StringValue("alpha"),
		types.StringValue("beta"),
		types.StringValue("gamma"),
	}
	list, diags := types.ListValueFrom(ctx, types.StringType, elems)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	result := listToStrings(ctx, list)
	if len(result) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(result))
	}
	expected := []string{"alpha", "beta", "gamma"}
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("index %d: expected %q, got %q", i, expected[i], v)
		}
	}
}

func TestStringsToList_EmptySlice(t *testing.T) {
	ctx := context.Background()
	result := stringsToList(ctx, []string{})
	if !result.IsNull() {
		t.Fatalf("expected null list for empty slice, got %v", result)
	}
}

func TestStringsToList_NilSlice(t *testing.T) {
	ctx := context.Background()
	result := stringsToList(ctx, nil)
	if !result.IsNull() {
		t.Fatalf("expected null list for nil slice, got %v", result)
	}
}

func TestStringsToList_WithValues(t *testing.T) {
	ctx := context.Background()
	result := stringsToList(ctx, []string{"one", "two", "three"})
	if result.IsNull() || result.IsUnknown() {
		t.Fatalf("expected non-null non-unknown list, got null=%v unknown=%v", result.IsNull(), result.IsUnknown())
	}
	// Round-trip: convert back and verify
	back := listToStrings(ctx, result)
	if len(back) != 3 {
		t.Fatalf("expected 3 elements after round-trip, got %d", len(back))
	}
	expected := []string{"one", "two", "three"}
	for i, v := range back {
		if v != expected[i] {
			t.Errorf("round-trip index %d: expected %q, got %q", i, expected[i], v)
		}
	}
}

func TestStringsToList_SingleValue(t *testing.T) {
	ctx := context.Background()
	result := stringsToList(ctx, []string{"solo"})
	back := listToStrings(ctx, result)
	if len(back) != 1 || back[0] != "solo" {
		t.Fatalf("expected [\"solo\"], got %v", back)
	}
}
