package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestVirtualModelToAPI(t *testing.T) {
	ctx := context.Background()

	model := virtualResourceModel{
		Name:        types.StringValue("my-virtual"),
		PackageType: types.StringValue("helm"),
		Description: types.StringValue("Virtual helm repo"),
		Members:     stringsToList(ctx, []string{"remote-a", "remote-b"}),
	}

	api := virtualModelToAPI(ctx, model)

	if api.Name != "my-virtual" {
		t.Errorf("Name: expected my-virtual, got %s", api.Name)
	}
	if api.PackageType != "helm" {
		t.Errorf("PackageType: expected helm, got %s", api.PackageType)
	}
	if api.Description != "Virtual helm repo" {
		t.Errorf("Description: expected 'Virtual helm repo', got %s", api.Description)
	}
	if len(api.Members) != 2 {
		t.Fatalf("Members: expected 2, got %d", len(api.Members))
	}
	if api.Members[0] != "remote-a" || api.Members[1] != "remote-b" {
		t.Errorf("Members: expected [remote-a remote-b], got %v", api.Members)
	}
}

func TestVirtualModelToAPI_NullMembers(t *testing.T) {
	ctx := context.Background()

	model := virtualResourceModel{
		Name:        types.StringValue("no-members"),
		PackageType: types.StringValue("pypi"),
		Description: types.StringValue(""),
		Members:     types.ListNull(types.StringType),
	}

	api := virtualModelToAPI(ctx, model)

	if api.Members != nil {
		t.Errorf("Members: expected nil for null list, got %v", api.Members)
	}
}

func TestVirtualAPIToModel(t *testing.T) {
	ctx := context.Background()

	api := virtualAPI{
		Name:        "my-virtual",
		PackageType: "helm",
		Description: "Virtual helm repo",
		Members:     []string{"remote-a", "remote-b"},
		ManagedBy:   "terraform",
	}

	model := virtualAPIToModel(ctx, api)

	if model.Name.ValueString() != "my-virtual" {
		t.Errorf("Name: expected my-virtual, got %s", model.Name.ValueString())
	}
	if model.PackageType.ValueString() != "helm" {
		t.Errorf("PackageType: expected helm, got %s", model.PackageType.ValueString())
	}
	if model.Description.ValueString() != "Virtual helm repo" {
		t.Errorf("Description: expected 'Virtual helm repo', got %s", model.Description.ValueString())
	}

	members := listToStrings(ctx, model.Members)
	if len(members) != 2 {
		t.Fatalf("Members: expected 2, got %d", len(members))
	}
	if members[0] != "remote-a" || members[1] != "remote-b" {
		t.Errorf("Members: expected [remote-a remote-b], got %v", members)
	}
}

func TestVirtualAPIToModel_EmptyMembers(t *testing.T) {
	ctx := context.Background()

	api := virtualAPI{
		Name:        "empty-members",
		PackageType: "pypi",
		Members:     nil,
	}

	model := virtualAPIToModel(ctx, api)

	if !model.Members.IsNull() {
		t.Errorf("Members: expected null for nil/empty slice, got %v", model.Members)
	}
}

func TestVirtualRoundTrip(t *testing.T) {
	ctx := context.Background()

	original := virtualAPI{
		Name:        "roundtrip-virtual",
		PackageType: "helm",
		Description: "Round trip test",
		Members:     []string{"a", "b", "c"},
	}

	model := virtualAPIToModel(ctx, original)
	result := virtualModelToAPI(ctx, model)

	if result.Name != original.Name {
		t.Errorf("Name: expected %s, got %s", original.Name, result.Name)
	}
	if result.PackageType != original.PackageType {
		t.Errorf("PackageType: expected %s, got %s", original.PackageType, result.PackageType)
	}
	if result.Description != original.Description {
		t.Errorf("Description: expected %s, got %s", original.Description, result.Description)
	}
	if len(result.Members) != len(original.Members) {
		t.Fatalf("Members length: expected %d, got %d", len(original.Members), len(result.Members))
	}
	for i := range original.Members {
		if result.Members[i] != original.Members[i] {
			t.Errorf("Members[%d]: expected %s, got %s", i, original.Members[i], result.Members[i])
		}
	}
}

func TestVirtualResource_Metadata(t *testing.T) {
	r := NewVirtualResource()
	req := resource.MetadataRequest{ProviderTypeName: "artifactapi"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)
	if resp.TypeName != "artifactapi_virtual" {
		t.Errorf("expected artifactapi_virtual, got %s", resp.TypeName)
	}
}

func TestVirtualResource_Schema(t *testing.T) {
	r := NewVirtualResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	expectedAttrs := []string{"name", "package_type", "description", "members"}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing expected attribute: %s", attr)
		}
	}
}

func TestNewVirtualResource_Type(t *testing.T) {
	r := NewVirtualResource()
	_, ok := r.(*virtualResource)
	if !ok {
		t.Error("expected *virtualResource")
	}
}
