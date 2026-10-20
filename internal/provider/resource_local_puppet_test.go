package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestLocalPuppetModelToAPI(t *testing.T) {
	model := localPuppetResourceModel{
		Name:        types.StringValue("puppet-internal"),
		Description: types.StringValue("Internal Puppet forge"),
	}

	api := localPuppetModelToAPI(model)

	if api.Name != "puppet-internal" {
		t.Errorf("Name: expected puppet-internal, got %s", api.Name)
	}
	if api.PackageType != "puppet" {
		t.Errorf("PackageType: expected puppet, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
	if api.Description != "Internal Puppet forge" {
		t.Errorf("Description: expected 'Internal Puppet forge', got %s", api.Description)
	}
}

func TestLocalPuppetModelToAPI_EmptyDescription(t *testing.T) {
	model := localPuppetResourceModel{
		Name:        types.StringValue("puppet-empty"),
		Description: types.StringValue(""),
	}

	api := localPuppetModelToAPI(model)

	if api.Name != "puppet-empty" {
		t.Errorf("Name: expected puppet-empty, got %s", api.Name)
	}
	if api.Description != "" {
		t.Errorf("Description: expected empty string, got %s", api.Description)
	}
	if api.PackageType != "puppet" {
		t.Errorf("PackageType: expected puppet, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
}

func TestLocalPuppetAPIToModel(t *testing.T) {
	api := remoteAPI{
		Name:        "puppet-internal",
		PackageType: "puppet",
		RepoType:    "local",
		Description: "Internal Puppet forge",
		ManagedBy:   "terraform",
	}

	model := localPuppetAPIToModel(api)

	if model.Name.ValueString() != "puppet-internal" {
		t.Errorf("Name: expected puppet-internal, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Internal Puppet forge" {
		t.Errorf("Description: expected 'Internal Puppet forge', got %s", model.Description.ValueString())
	}
}

func TestLocalPuppetRoundTrip(t *testing.T) {
	original := localPuppetResourceModel{
		Name:        types.StringValue("roundtrip-puppet"),
		Description: types.StringValue("Round trip test"),
	}

	api := localPuppetModelToAPI(original)
	result := localPuppetAPIToModel(api)

	if result.Name.ValueString() != original.Name.ValueString() {
		t.Errorf("Name: expected %s, got %s", original.Name.ValueString(), result.Name.ValueString())
	}
	if result.Description.ValueString() != original.Description.ValueString() {
		t.Errorf("Description: expected %s, got %s", original.Description.ValueString(), result.Description.ValueString())
	}
}

func TestLocalPuppetResource_Metadata(t *testing.T) {
	r := NewLocalPuppetResource()
	req := resource.MetadataRequest{ProviderTypeName: "artifactapi"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)
	if resp.TypeName != "artifactapi_local_puppet" {
		t.Errorf("expected artifactapi_local_puppet, got %s", resp.TypeName)
	}
}

func TestLocalPuppetResource_Schema(t *testing.T) {
	r := NewLocalPuppetResource()
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

func TestNewLocalPuppetResource_Type(t *testing.T) {
	r := NewLocalPuppetResource()
	_, ok := r.(*localPuppetResource)
	if !ok {
		t.Error("expected *localPuppetResource")
	}
}
