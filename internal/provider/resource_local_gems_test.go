package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestLocalGemsModelToAPI(t *testing.T) {
	model := localGemsResourceModel{
		Name:        types.StringValue("gems-internal"),
		Description: types.StringValue("Internal gems repository"),
	}

	api := localGemsModelToAPI(model)

	if api.Name != "gems-internal" {
		t.Errorf("Name: expected gems-internal, got %s", api.Name)
	}
	if api.PackageType != "gems" {
		t.Errorf("PackageType: expected gems, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
	if api.Description != "Internal gems repository" {
		t.Errorf("Description: expected 'Internal gems repository', got %s", api.Description)
	}
}

func TestLocalGemsModelToAPI_EmptyDescription(t *testing.T) {
	model := localGemsResourceModel{
		Name:        types.StringValue("gems-empty"),
		Description: types.StringValue(""),
	}

	api := localGemsModelToAPI(model)

	if api.Name != "gems-empty" {
		t.Errorf("Name: expected gems-empty, got %s", api.Name)
	}
	if api.Description != "" {
		t.Errorf("Description: expected empty string, got %s", api.Description)
	}
	if api.PackageType != "gems" {
		t.Errorf("PackageType: expected gems, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
}

func TestLocalGemsAPIToModel(t *testing.T) {
	api := remoteAPI{
		Name:        "gems-internal",
		PackageType: "gems",
		RepoType:    "local",
		Description: "Internal gems repository",
		ManagedBy:   "terraform",
	}

	model := localGemsAPIToModel(api)

	if model.Name.ValueString() != "gems-internal" {
		t.Errorf("Name: expected gems-internal, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Internal gems repository" {
		t.Errorf("Description: expected 'Internal gems repository', got %s", model.Description.ValueString())
	}
}

func TestLocalGemsRoundTrip(t *testing.T) {
	original := localGemsResourceModel{
		Name:        types.StringValue("roundtrip-gems"),
		Description: types.StringValue("Round trip test"),
	}

	api := localGemsModelToAPI(original)
	result := localGemsAPIToModel(api)

	if result.Name.ValueString() != original.Name.ValueString() {
		t.Errorf("Name: expected %s, got %s", original.Name.ValueString(), result.Name.ValueString())
	}
	if result.Description.ValueString() != original.Description.ValueString() {
		t.Errorf("Description: expected %s, got %s", original.Description.ValueString(), result.Description.ValueString())
	}
}

func TestLocalGemsResource_Metadata(t *testing.T) {
	r := NewLocalGemsResource()
	req := resource.MetadataRequest{ProviderTypeName: "artifactapi"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)
	if resp.TypeName != "artifactapi_local_gems" {
		t.Errorf("expected artifactapi_local_gems, got %s", resp.TypeName)
	}
}

func TestLocalGemsResource_Schema(t *testing.T) {
	r := NewLocalGemsResource()
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

func TestNewLocalGemsResource_Type(t *testing.T) {
	r := NewLocalGemsResource()
	_, ok := r.(*localGemsResource)
	if !ok {
		t.Error("expected *localGemsResource")
	}
}
