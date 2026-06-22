package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestLocalTerraformModelToAPI(t *testing.T) {
	model := localTerraformResourceModel{
		Name:        types.StringValue("tf-internal"),
		Description: types.StringValue("Internal terraform registry"),
	}

	api := localTerraformModelToAPI(model)

	if api.Name != "tf-internal" {
		t.Errorf("Name: expected tf-internal, got %s", api.Name)
	}
	if api.PackageType != "terraform" {
		t.Errorf("PackageType: expected terraform, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
	if api.Description != "Internal terraform registry" {
		t.Errorf("Description: expected 'Internal terraform registry', got %s", api.Description)
	}
}

func TestLocalTerraformModelToAPI_EmptyDescription(t *testing.T) {
	model := localTerraformResourceModel{
		Name:        types.StringValue("tf-empty"),
		Description: types.StringValue(""),
	}

	api := localTerraformModelToAPI(model)

	if api.Name != "tf-empty" {
		t.Errorf("Name: expected tf-empty, got %s", api.Name)
	}
	if api.Description != "" {
		t.Errorf("Description: expected empty string, got %s", api.Description)
	}
	if api.PackageType != "terraform" {
		t.Errorf("PackageType: expected terraform, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
}

func TestLocalTerraformAPIToModel(t *testing.T) {
	api := remoteAPI{
		Name:        "tf-internal",
		PackageType: "terraform",
		RepoType:    "local",
		Description: "Internal terraform registry",
		ManagedBy:   "terraform",
	}

	model := localTerraformAPIToModel(api)

	if model.Name.ValueString() != "tf-internal" {
		t.Errorf("Name: expected tf-internal, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Internal terraform registry" {
		t.Errorf("Description: expected 'Internal terraform registry', got %s", model.Description.ValueString())
	}
}

func TestLocalTerraformRoundTrip(t *testing.T) {
	original := localTerraformResourceModel{
		Name:        types.StringValue("roundtrip-local"),
		Description: types.StringValue("Round trip test"),
	}

	api := localTerraformModelToAPI(original)
	result := localTerraformAPIToModel(api)

	if result.Name.ValueString() != original.Name.ValueString() {
		t.Errorf("Name: expected %s, got %s", original.Name.ValueString(), result.Name.ValueString())
	}
	if result.Description.ValueString() != original.Description.ValueString() {
		t.Errorf("Description: expected %s, got %s", original.Description.ValueString(), result.Description.ValueString())
	}
}

func TestLocalTerraformResource_Metadata(t *testing.T) {
	r := NewLocalTerraformResource()
	req := resource.MetadataRequest{ProviderTypeName: "artifactapi"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)
	if resp.TypeName != "artifactapi_local_terraform" {
		t.Errorf("expected artifactapi_local_terraform, got %s", resp.TypeName)
	}
}

func TestLocalTerraformResource_Schema(t *testing.T) {
	r := NewLocalTerraformResource()
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

func TestNewLocalTerraformResource_Type(t *testing.T) {
	r := NewLocalTerraformResource()
	_, ok := r.(*localTerraformResource)
	if !ok {
		t.Error("expected *localTerraformResource")
	}
}
