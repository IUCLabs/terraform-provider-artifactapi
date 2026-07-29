package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestLocalGenericModelToAPI(t *testing.T) {
	model := localGenericResourceModel{
		Name:        types.StringValue("rootfs-images"),
		Description: types.StringValue("Internal Generic repository"),
	}

	api := localGenericModelToAPI(model)

	if api.Name != "rootfs-images" {
		t.Errorf("Name: expected rootfs-images, got %s", api.Name)
	}
	if api.PackageType != "generic" {
		t.Errorf("PackageType: expected generic, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
	if api.Description != "Internal Generic repository" {
		t.Errorf("Description: expected 'Internal Generic repository', got %s", api.Description)
	}
}

func TestLocalGenericModelToAPI_EmptyDescription(t *testing.T) {
	model := localGenericResourceModel{
		Name:        types.StringValue("generic-empty"),
		Description: types.StringValue(""),
	}

	api := localGenericModelToAPI(model)

	if api.Name != "generic-empty" {
		t.Errorf("Name: expected generic-empty, got %s", api.Name)
	}
	if api.Description != "" {
		t.Errorf("Description: expected empty string, got %s", api.Description)
	}
	if api.PackageType != "generic" {
		t.Errorf("PackageType: expected generic, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
}

func TestLocalGenericAPIToModel(t *testing.T) {
	api := remoteAPI{
		Name:        "rootfs-images",
		PackageType: "generic",
		RepoType:    "local",
		Description: "Internal Generic repository",
		ManagedBy:   "terraform",
	}

	model := localGenericAPIToModel(api)

	if model.Name.ValueString() != "rootfs-images" {
		t.Errorf("Name: expected rootfs-images, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Internal Generic repository" {
		t.Errorf("Description: expected 'Internal Generic repository', got %s", model.Description.ValueString())
	}
}

func TestLocalGenericRoundTrip(t *testing.T) {
	original := localGenericResourceModel{
		Name:        types.StringValue("roundtrip-generic"),
		Description: types.StringValue("Round trip test"),
	}

	api := localGenericModelToAPI(original)
	result := localGenericAPIToModel(api)

	if result.Name.ValueString() != original.Name.ValueString() {
		t.Errorf("Name: expected %s, got %s", original.Name.ValueString(), result.Name.ValueString())
	}
	if result.Description.ValueString() != original.Description.ValueString() {
		t.Errorf("Description: expected %s, got %s", original.Description.ValueString(), result.Description.ValueString())
	}
}

func TestLocalGenericResource_Metadata(t *testing.T) {
	r := NewLocalGenericResource()
	req := resource.MetadataRequest{ProviderTypeName: "artifactapi"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)
	if resp.TypeName != "artifactapi_local_generic" {
		t.Errorf("expected artifactapi_local_generic, got %s", resp.TypeName)
	}
}

func TestLocalGenericResource_Schema(t *testing.T) {
	r := NewLocalGenericResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	expectedAttrs := []string{"name", "description"}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing expected attribute: %s", attr)
		}
	}

	if len(resp.Schema.Attributes) != len(expectedAttrs) {
		t.Errorf("expected %d attributes, got %d", len(expectedAttrs), len(resp.Schema.Attributes))
	}
}

func TestNewLocalGenericResource_Type(t *testing.T) {
	r := NewLocalGenericResource()
	_, ok := r.(*localGenericResource)
	if !ok {
		t.Error("expected *localGenericResource")
	}
}
