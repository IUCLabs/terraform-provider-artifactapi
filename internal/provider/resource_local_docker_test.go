package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestLocalDockerModelToAPI(t *testing.T) {
	model := localDockerResourceModel{
		Name:        types.StringValue("docker-internal"),
		Description: types.StringValue("Internal container registry"),
	}

	api := localDockerModelToAPI(model)

	if api.Name != "docker-internal" {
		t.Errorf("Name: expected docker-internal, got %s", api.Name)
	}
	if api.PackageType != "docker" {
		t.Errorf("PackageType: expected docker, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
	if api.Description != "Internal container registry" {
		t.Errorf("Description: expected 'Internal container registry', got %s", api.Description)
	}
}

func TestLocalDockerModelToAPI_EmptyDescription(t *testing.T) {
	model := localDockerResourceModel{
		Name:        types.StringValue("docker-empty"),
		Description: types.StringValue(""),
	}

	api := localDockerModelToAPI(model)

	if api.Name != "docker-empty" {
		t.Errorf("Name: expected docker-empty, got %s", api.Name)
	}
	if api.Description != "" {
		t.Errorf("Description: expected empty string, got %s", api.Description)
	}
	if api.PackageType != "docker" {
		t.Errorf("PackageType: expected docker, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
}

func TestLocalDockerAPIToModel(t *testing.T) {
	api := remoteAPI{
		Name:        "docker-internal",
		PackageType: "docker",
		RepoType:    "local",
		Description: "Internal container registry",
		ManagedBy:   "terraform",
	}

	model := localDockerAPIToModel(api)

	if model.Name.ValueString() != "docker-internal" {
		t.Errorf("Name: expected docker-internal, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Internal container registry" {
		t.Errorf("Description: expected 'Internal container registry', got %s", model.Description.ValueString())
	}
}

func TestLocalDockerRoundTrip(t *testing.T) {
	original := localDockerResourceModel{
		Name:        types.StringValue("roundtrip-docker"),
		Description: types.StringValue("Round trip test"),
	}

	api := localDockerModelToAPI(original)
	result := localDockerAPIToModel(api)

	if result.Name.ValueString() != original.Name.ValueString() {
		t.Errorf("Name: expected %s, got %s", original.Name.ValueString(), result.Name.ValueString())
	}
	if result.Description.ValueString() != original.Description.ValueString() {
		t.Errorf("Description: expected %s, got %s", original.Description.ValueString(), result.Description.ValueString())
	}
}

func TestLocalDockerResource_Metadata(t *testing.T) {
	r := NewLocalDockerResource()
	req := resource.MetadataRequest{ProviderTypeName: "artifactapi"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)
	if resp.TypeName != "artifactapi_local_docker" {
		t.Errorf("expected artifactapi_local_docker, got %s", resp.TypeName)
	}
}

func TestLocalDockerResource_Schema(t *testing.T) {
	r := NewLocalDockerResource()
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

func TestNewLocalDockerResource_Type(t *testing.T) {
	r := NewLocalDockerResource()
	_, ok := r.(*localDockerResource)
	if !ok {
		t.Error("expected *localDockerResource")
	}
}
