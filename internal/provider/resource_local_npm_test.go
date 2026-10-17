package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestLocalNPMModelToAPI(t *testing.T) {
	model := localNPMResourceModel{
		Name:        types.StringValue("npm-internal"),
		Description: types.StringValue("Internal NPM registry"),
	}

	api := localNPMModelToAPI(model)

	if api.Name != "npm-internal" {
		t.Errorf("Name: expected npm-internal, got %s", api.Name)
	}
	if api.PackageType != "npm" {
		t.Errorf("PackageType: expected npm, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
	if api.Description != "Internal NPM registry" {
		t.Errorf("Description: expected 'Internal NPM registry', got %s", api.Description)
	}
}

func TestLocalNPMModelToAPI_EmptyDescription(t *testing.T) {
	model := localNPMResourceModel{
		Name:        types.StringValue("npm-empty"),
		Description: types.StringValue(""),
	}

	api := localNPMModelToAPI(model)

	if api.Name != "npm-empty" {
		t.Errorf("Name: expected npm-empty, got %s", api.Name)
	}
	if api.Description != "" {
		t.Errorf("Description: expected empty string, got %s", api.Description)
	}
	if api.PackageType != "npm" {
		t.Errorf("PackageType: expected npm, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
}

func TestLocalNPMAPIToModel(t *testing.T) {
	api := remoteAPI{
		Name:        "npm-internal",
		PackageType: "npm",
		RepoType:    "local",
		Description: "Internal NPM registry",
		ManagedBy:   "terraform",
	}

	model := localNPMAPIToModel(api)

	if model.Name.ValueString() != "npm-internal" {
		t.Errorf("Name: expected npm-internal, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Internal NPM registry" {
		t.Errorf("Description: expected 'Internal NPM registry', got %s", model.Description.ValueString())
	}
}

func TestLocalNPMRoundTrip(t *testing.T) {
	original := localNPMResourceModel{
		Name:        types.StringValue("roundtrip-npm"),
		Description: types.StringValue("Round trip test"),
	}

	api := localNPMModelToAPI(original)
	result := localNPMAPIToModel(api)

	if result.Name.ValueString() != original.Name.ValueString() {
		t.Errorf("Name: expected %s, got %s", original.Name.ValueString(), result.Name.ValueString())
	}
	if result.Description.ValueString() != original.Description.ValueString() {
		t.Errorf("Description: expected %s, got %s", original.Description.ValueString(), result.Description.ValueString())
	}
}

func TestLocalNPMResource_Metadata(t *testing.T) {
	r := NewLocalNPMResource()
	req := resource.MetadataRequest{ProviderTypeName: "artifactapi"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)
	if resp.TypeName != "artifactapi_local_npm" {
		t.Errorf("expected artifactapi_local_npm, got %s", resp.TypeName)
	}
}

func TestLocalNPMResource_Schema(t *testing.T) {
	r := NewLocalNPMResource()
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

func TestNewLocalNPMResource_Type(t *testing.T) {
	r := NewLocalNPMResource()
	_, ok := r.(*localNPMResource)
	if !ok {
		t.Error("expected *localNPMResource")
	}
}
