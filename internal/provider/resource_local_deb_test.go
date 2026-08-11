package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestLocalDebModelToAPI(t *testing.T) {
	model := localDebResourceModel{
		Name:        types.StringValue("deb-internal"),
		Description: types.StringValue("Internal Deb repository"),
	}

	api := localDebModelToAPI(model)

	if api.Name != "deb-internal" {
		t.Errorf("Name: expected deb-internal, got %s", api.Name)
	}
	if api.PackageType != "deb" {
		t.Errorf("PackageType: expected deb, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
	if api.Description != "Internal Deb repository" {
		t.Errorf("Description: expected 'Internal Deb repository', got %s", api.Description)
	}
}

func TestLocalDebModelToAPI_EmptyDescription(t *testing.T) {
	model := localDebResourceModel{
		Name:        types.StringValue("deb-empty"),
		Description: types.StringValue(""),
	}

	api := localDebModelToAPI(model)

	if api.Name != "deb-empty" {
		t.Errorf("Name: expected deb-empty, got %s", api.Name)
	}
	if api.Description != "" {
		t.Errorf("Description: expected empty string, got %s", api.Description)
	}
	if api.PackageType != "deb" {
		t.Errorf("PackageType: expected deb, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
}

func TestLocalDebAPIToModel(t *testing.T) {
	api := remoteAPI{
		Name:        "deb-internal",
		PackageType: "deb",
		RepoType:    "local",
		Description: "Internal Deb repository",
		ManagedBy:   "terraform",
	}

	model := localDebAPIToModel(api)

	if model.Name.ValueString() != "deb-internal" {
		t.Errorf("Name: expected deb-internal, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Internal Deb repository" {
		t.Errorf("Description: expected 'Internal Deb repository', got %s", model.Description.ValueString())
	}
}

func TestLocalDebRoundTrip(t *testing.T) {
	original := localDebResourceModel{
		Name:        types.StringValue("roundtrip-deb"),
		Description: types.StringValue("Round trip test"),
	}

	api := localDebModelToAPI(original)
	result := localDebAPIToModel(api)

	if result.Name.ValueString() != original.Name.ValueString() {
		t.Errorf("Name: expected %s, got %s", original.Name.ValueString(), result.Name.ValueString())
	}
	if result.Description.ValueString() != original.Description.ValueString() {
		t.Errorf("Description: expected %s, got %s", original.Description.ValueString(), result.Description.ValueString())
	}
}

func TestLocalDebResource_Metadata(t *testing.T) {
	r := NewLocalDebResource()
	req := resource.MetadataRequest{ProviderTypeName: "artifactapi"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)
	if resp.TypeName != "artifactapi_local_deb" {
		t.Errorf("expected artifactapi_local_deb, got %s", resp.TypeName)
	}
}

func TestLocalDebResource_Schema(t *testing.T) {
	r := NewLocalDebResource()
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

func TestNewLocalDebResource_Type(t *testing.T) {
	r := NewLocalDebResource()
	_, ok := r.(*localDebResource)
	if !ok {
		t.Error("expected *localDebResource")
	}
}
