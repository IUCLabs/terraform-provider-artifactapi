package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestLocalAnsibleModelToAPI(t *testing.T) {
	model := localAnsibleResourceModel{
		Name:        types.StringValue("ansible-internal"),
		Description: types.StringValue("Internal ansible repository"),
	}

	api := localAnsibleModelToAPI(model)

	if api.Name != "ansible-internal" {
		t.Errorf("Name: expected ansible-internal, got %s", api.Name)
	}
	if api.PackageType != "ansible" {
		t.Errorf("PackageType: expected ansible, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
	if api.Description != "Internal ansible repository" {
		t.Errorf("Description: expected 'Internal ansible repository', got %s", api.Description)
	}
}

func TestLocalAnsibleModelToAPI_EmptyDescription(t *testing.T) {
	model := localAnsibleResourceModel{
		Name:        types.StringValue("ansible-empty"),
		Description: types.StringValue(""),
	}

	api := localAnsibleModelToAPI(model)

	if api.Name != "ansible-empty" {
		t.Errorf("Name: expected ansible-empty, got %s", api.Name)
	}
	if api.Description != "" {
		t.Errorf("Description: expected empty string, got %s", api.Description)
	}
	if api.PackageType != "ansible" {
		t.Errorf("PackageType: expected ansible, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
}

func TestLocalAnsibleAPIToModel(t *testing.T) {
	api := remoteAPI{
		Name:        "ansible-internal",
		PackageType: "ansible",
		RepoType:    "local",
		Description: "Internal ansible repository",
		ManagedBy:   "terraform",
	}

	model := localAnsibleAPIToModel(api)

	if model.Name.ValueString() != "ansible-internal" {
		t.Errorf("Name: expected ansible-internal, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Internal ansible repository" {
		t.Errorf("Description: expected 'Internal ansible repository', got %s", model.Description.ValueString())
	}
}

func TestLocalAnsibleRoundTrip(t *testing.T) {
	original := localAnsibleResourceModel{
		Name:        types.StringValue("roundtrip-ansible"),
		Description: types.StringValue("Round trip test"),
	}

	api := localAnsibleModelToAPI(original)
	result := localAnsibleAPIToModel(api)

	if result.Name.ValueString() != original.Name.ValueString() {
		t.Errorf("Name: expected %s, got %s", original.Name.ValueString(), result.Name.ValueString())
	}
	if result.Description.ValueString() != original.Description.ValueString() {
		t.Errorf("Description: expected %s, got %s", original.Description.ValueString(), result.Description.ValueString())
	}
}

func TestLocalAnsibleResource_Metadata(t *testing.T) {
	r := NewLocalAnsibleResource()
	req := resource.MetadataRequest{ProviderTypeName: "artifactapi"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)
	if resp.TypeName != "artifactapi_local_ansible" {
		t.Errorf("expected artifactapi_local_ansible, got %s", resp.TypeName)
	}
}

func TestLocalAnsibleResource_Schema(t *testing.T) {
	r := NewLocalAnsibleResource()
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

func TestNewLocalAnsibleResource_Type(t *testing.T) {
	r := NewLocalAnsibleResource()
	_, ok := r.(*localAnsibleResource)
	if !ok {
		t.Error("expected *localAnsibleResource")
	}
}
