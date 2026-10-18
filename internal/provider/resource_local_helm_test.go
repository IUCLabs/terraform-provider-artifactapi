package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestLocalHelmModelToAPI(t *testing.T) {
	model := localHelmResourceModel{
		Name:        types.StringValue("helm-internal"),
		Description: types.StringValue("Internal Helm repository"),
	}

	api := localHelmModelToAPI(model)

	if api.Name != "helm-internal" {
		t.Errorf("Name: expected helm-internal, got %s", api.Name)
	}
	if api.PackageType != "helm" {
		t.Errorf("PackageType: expected helm, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
	if api.Description != "Internal Helm repository" {
		t.Errorf("Description: expected 'Internal Helm repository', got %s", api.Description)
	}
}

func TestLocalHelmModelToAPI_EmptyDescription(t *testing.T) {
	model := localHelmResourceModel{
		Name:        types.StringValue("helm-empty"),
		Description: types.StringValue(""),
	}

	api := localHelmModelToAPI(model)

	if api.Name != "helm-empty" {
		t.Errorf("Name: expected helm-empty, got %s", api.Name)
	}
	if api.Description != "" {
		t.Errorf("Description: expected empty string, got %s", api.Description)
	}
	if api.PackageType != "helm" {
		t.Errorf("PackageType: expected helm, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
}

func TestLocalHelmAPIToModel(t *testing.T) {
	api := remoteAPI{
		Name:        "helm-internal",
		PackageType: "helm",
		RepoType:    "local",
		Description: "Internal Helm repository",
		ManagedBy:   "terraform",
	}

	model := localHelmAPIToModel(api)

	if model.Name.ValueString() != "helm-internal" {
		t.Errorf("Name: expected helm-internal, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Internal Helm repository" {
		t.Errorf("Description: expected 'Internal Helm repository', got %s", model.Description.ValueString())
	}
}

func TestLocalHelmRoundTrip(t *testing.T) {
	original := localHelmResourceModel{
		Name:        types.StringValue("roundtrip-helm"),
		Description: types.StringValue("Round trip test"),
	}

	api := localHelmModelToAPI(original)
	result := localHelmAPIToModel(api)

	if result.Name.ValueString() != original.Name.ValueString() {
		t.Errorf("Name: expected %s, got %s", original.Name.ValueString(), result.Name.ValueString())
	}
	if result.Description.ValueString() != original.Description.ValueString() {
		t.Errorf("Description: expected %s, got %s", original.Description.ValueString(), result.Description.ValueString())
	}
}

func TestLocalHelmResource_Metadata(t *testing.T) {
	r := NewLocalHelmResource()
	req := resource.MetadataRequest{ProviderTypeName: "artifactapi"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)
	if resp.TypeName != "artifactapi_local_helm" {
		t.Errorf("expected artifactapi_local_helm, got %s", resp.TypeName)
	}
}

func TestLocalHelmResource_Schema(t *testing.T) {
	r := NewLocalHelmResource()
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

func TestNewLocalHelmResource_Type(t *testing.T) {
	r := NewLocalHelmResource()
	_, ok := r.(*localHelmResource)
	if !ok {
		t.Error("expected *localHelmResource")
	}
}
