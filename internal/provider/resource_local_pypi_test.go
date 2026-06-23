package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestLocalPyPIModelToAPI(t *testing.T) {
	model := localPyPIResourceModel{
		Name:        types.StringValue("pypi-internal"),
		Description: types.StringValue("Internal PyPI registry"),
	}

	api := localPyPIModelToAPI(model)

	if api.Name != "pypi-internal" {
		t.Errorf("Name: expected pypi-internal, got %s", api.Name)
	}
	if api.PackageType != "pypi" {
		t.Errorf("PackageType: expected pypi, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
	if api.Description != "Internal PyPI registry" {
		t.Errorf("Description: expected 'Internal PyPI registry', got %s", api.Description)
	}
}

func TestLocalPyPIModelToAPI_EmptyDescription(t *testing.T) {
	model := localPyPIResourceModel{
		Name:        types.StringValue("pypi-empty"),
		Description: types.StringValue(""),
	}

	api := localPyPIModelToAPI(model)

	if api.Name != "pypi-empty" {
		t.Errorf("Name: expected pypi-empty, got %s", api.Name)
	}
	if api.Description != "" {
		t.Errorf("Description: expected empty string, got %s", api.Description)
	}
	if api.PackageType != "pypi" {
		t.Errorf("PackageType: expected pypi, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
}

func TestLocalPyPIAPIToModel(t *testing.T) {
	api := remoteAPI{
		Name:        "pypi-internal",
		PackageType: "pypi",
		RepoType:    "local",
		Description: "Internal PyPI registry",
		ManagedBy:   "terraform",
	}

	model := localPyPIAPIToModel(api)

	if model.Name.ValueString() != "pypi-internal" {
		t.Errorf("Name: expected pypi-internal, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Internal PyPI registry" {
		t.Errorf("Description: expected 'Internal PyPI registry', got %s", model.Description.ValueString())
	}
}

func TestLocalPyPIRoundTrip(t *testing.T) {
	original := localPyPIResourceModel{
		Name:        types.StringValue("roundtrip-pypi"),
		Description: types.StringValue("Round trip test"),
	}

	api := localPyPIModelToAPI(original)
	result := localPyPIAPIToModel(api)

	if result.Name.ValueString() != original.Name.ValueString() {
		t.Errorf("Name: expected %s, got %s", original.Name.ValueString(), result.Name.ValueString())
	}
	if result.Description.ValueString() != original.Description.ValueString() {
		t.Errorf("Description: expected %s, got %s", original.Description.ValueString(), result.Description.ValueString())
	}
}

func TestLocalPyPIResource_Metadata(t *testing.T) {
	r := NewLocalPyPIResource()
	req := resource.MetadataRequest{ProviderTypeName: "artifactapi"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)
	if resp.TypeName != "artifactapi_local_pypi" {
		t.Errorf("expected artifactapi_local_pypi, got %s", resp.TypeName)
	}
}

func TestLocalPyPIResource_Schema(t *testing.T) {
	r := NewLocalPyPIResource()
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

func TestNewLocalPyPIResource_Type(t *testing.T) {
	r := NewLocalPyPIResource()
	_, ok := r.(*localPyPIResource)
	if !ok {
		t.Error("expected *localPyPIResource")
	}
}
