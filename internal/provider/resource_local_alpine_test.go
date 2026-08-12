package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestLocalAlpineModelToAPI(t *testing.T) {
	model := localAlpineResourceModel{
		Name:        types.StringValue("alpine-internal"),
		Description: types.StringValue("Internal Alpine repository"),
	}

	api := localAlpineModelToAPI(model)

	if api.Name != "alpine-internal" {
		t.Errorf("Name: expected alpine-internal, got %s", api.Name)
	}
	if api.PackageType != "alpine" {
		t.Errorf("PackageType: expected alpine, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
	if api.Description != "Internal Alpine repository" {
		t.Errorf("Description: expected 'Internal Alpine repository', got %s", api.Description)
	}
}

func TestLocalAlpineModelToAPI_EmptyDescription(t *testing.T) {
	model := localAlpineResourceModel{
		Name:        types.StringValue("alpine-empty"),
		Description: types.StringValue(""),
	}

	api := localAlpineModelToAPI(model)

	if api.Name != "alpine-empty" {
		t.Errorf("Name: expected alpine-empty, got %s", api.Name)
	}
	if api.Description != "" {
		t.Errorf("Description: expected empty string, got %s", api.Description)
	}
	if api.PackageType != "alpine" {
		t.Errorf("PackageType: expected alpine, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
}

func TestLocalAlpineAPIToModel(t *testing.T) {
	api := remoteAPI{
		Name:        "alpine-internal",
		PackageType: "alpine",
		RepoType:    "local",
		Description: "Internal Alpine repository",
		ManagedBy:   "terraform",
	}

	model := localAlpineAPIToModel(api)

	if model.Name.ValueString() != "alpine-internal" {
		t.Errorf("Name: expected alpine-internal, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Internal Alpine repository" {
		t.Errorf("Description: expected 'Internal Alpine repository', got %s", model.Description.ValueString())
	}
}

func TestLocalAlpineRoundTrip(t *testing.T) {
	original := localAlpineResourceModel{
		Name:        types.StringValue("roundtrip-alpine"),
		Description: types.StringValue("Round trip test"),
	}

	api := localAlpineModelToAPI(original)
	result := localAlpineAPIToModel(api)

	if result.Name.ValueString() != original.Name.ValueString() {
		t.Errorf("Name: expected %s, got %s", original.Name.ValueString(), result.Name.ValueString())
	}
	if result.Description.ValueString() != original.Description.ValueString() {
		t.Errorf("Description: expected %s, got %s", original.Description.ValueString(), result.Description.ValueString())
	}
}

func TestLocalAlpineResource_Metadata(t *testing.T) {
	r := NewLocalAlpineResource()
	req := resource.MetadataRequest{ProviderTypeName: "artifactapi"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)
	if resp.TypeName != "artifactapi_local_alpine" {
		t.Errorf("expected artifactapi_local_alpine, got %s", resp.TypeName)
	}
}

func TestLocalAlpineResource_Schema(t *testing.T) {
	r := NewLocalAlpineResource()
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

func TestNewLocalAlpineResource_Type(t *testing.T) {
	r := NewLocalAlpineResource()
	_, ok := r.(*localAlpineResource)
	if !ok {
		t.Error("expected *localAlpineResource")
	}
}
