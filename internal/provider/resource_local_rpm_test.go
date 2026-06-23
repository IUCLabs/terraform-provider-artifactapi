package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestLocalRPMModelToAPI(t *testing.T) {
	model := localRPMResourceModel{
		Name:        types.StringValue("rpm-internal"),
		Description: types.StringValue("Internal RPM repository"),
	}

	api := localRPMModelToAPI(model)

	if api.Name != "rpm-internal" {
		t.Errorf("Name: expected rpm-internal, got %s", api.Name)
	}
	if api.PackageType != "rpm" {
		t.Errorf("PackageType: expected rpm, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
	if api.Description != "Internal RPM repository" {
		t.Errorf("Description: expected 'Internal RPM repository', got %s", api.Description)
	}
}

func TestLocalRPMModelToAPI_EmptyDescription(t *testing.T) {
	model := localRPMResourceModel{
		Name:        types.StringValue("rpm-empty"),
		Description: types.StringValue(""),
	}

	api := localRPMModelToAPI(model)

	if api.Name != "rpm-empty" {
		t.Errorf("Name: expected rpm-empty, got %s", api.Name)
	}
	if api.Description != "" {
		t.Errorf("Description: expected empty string, got %s", api.Description)
	}
	if api.PackageType != "rpm" {
		t.Errorf("PackageType: expected rpm, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
}

func TestLocalRPMAPIToModel(t *testing.T) {
	api := remoteAPI{
		Name:        "rpm-internal",
		PackageType: "rpm",
		RepoType:    "local",
		Description: "Internal RPM repository",
		ManagedBy:   "terraform",
	}

	model := localRPMAPIToModel(api)

	if model.Name.ValueString() != "rpm-internal" {
		t.Errorf("Name: expected rpm-internal, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Internal RPM repository" {
		t.Errorf("Description: expected 'Internal RPM repository', got %s", model.Description.ValueString())
	}
}

func TestLocalRPMRoundTrip(t *testing.T) {
	original := localRPMResourceModel{
		Name:        types.StringValue("roundtrip-rpm"),
		Description: types.StringValue("Round trip test"),
	}

	api := localRPMModelToAPI(original)
	result := localRPMAPIToModel(api)

	if result.Name.ValueString() != original.Name.ValueString() {
		t.Errorf("Name: expected %s, got %s", original.Name.ValueString(), result.Name.ValueString())
	}
	if result.Description.ValueString() != original.Description.ValueString() {
		t.Errorf("Description: expected %s, got %s", original.Description.ValueString(), result.Description.ValueString())
	}
}

func TestLocalRPMResource_Metadata(t *testing.T) {
	r := NewLocalRPMResource()
	req := resource.MetadataRequest{ProviderTypeName: "artifactapi"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)
	if resp.TypeName != "artifactapi_local_rpm" {
		t.Errorf("expected artifactapi_local_rpm, got %s", resp.TypeName)
	}
}

func TestLocalRPMResource_Schema(t *testing.T) {
	r := NewLocalRPMResource()
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

func TestNewLocalRPMResource_Type(t *testing.T) {
	r := NewLocalRPMResource()
	_, ok := r.(*localRPMResource)
	if !ok {
		t.Error("expected *localRPMResource")
	}
}
