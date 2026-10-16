package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestLocalGoProxyModelToAPI(t *testing.T) {
	model := localGoProxyResourceModel{
		Name:        types.StringValue("go-internal"),
		Description: types.StringValue("Internal Go modules"),
	}

	api := localGoProxyModelToAPI(model)

	if api.Name != "go-internal" {
		t.Errorf("Name: expected go-internal, got %s", api.Name)
	}
	if api.PackageType != "goproxy" {
		t.Errorf("PackageType: expected goproxy, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
	if api.Description != "Internal Go modules" {
		t.Errorf("Description: expected 'Internal Go modules', got %s", api.Description)
	}
}

func TestLocalGoProxyModelToAPI_EmptyDescription(t *testing.T) {
	model := localGoProxyResourceModel{
		Name:        types.StringValue("go-empty"),
		Description: types.StringValue(""),
	}

	api := localGoProxyModelToAPI(model)

	if api.Name != "go-empty" {
		t.Errorf("Name: expected go-empty, got %s", api.Name)
	}
	if api.Description != "" {
		t.Errorf("Description: expected empty string, got %s", api.Description)
	}
	if api.PackageType != "goproxy" {
		t.Errorf("PackageType: expected goproxy, got %s", api.PackageType)
	}
	if api.RepoType != "local" {
		t.Errorf("RepoType: expected local, got %s", api.RepoType)
	}
}

func TestLocalGoProxyAPIToModel(t *testing.T) {
	api := remoteAPI{
		Name:        "go-internal",
		PackageType: "goproxy",
		RepoType:    "local",
		Description: "Internal Go modules",
		ManagedBy:   "terraform",
	}

	model := localGoProxyAPIToModel(api)

	if model.Name.ValueString() != "go-internal" {
		t.Errorf("Name: expected go-internal, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Internal Go modules" {
		t.Errorf("Description: expected 'Internal Go modules', got %s", model.Description.ValueString())
	}
}

func TestLocalGoProxyRoundTrip(t *testing.T) {
	original := localGoProxyResourceModel{
		Name:        types.StringValue("roundtrip-go"),
		Description: types.StringValue("Round trip test"),
	}

	api := localGoProxyModelToAPI(original)
	result := localGoProxyAPIToModel(api)

	if result.Name.ValueString() != original.Name.ValueString() {
		t.Errorf("Name: expected %s, got %s", original.Name.ValueString(), result.Name.ValueString())
	}
	if result.Description.ValueString() != original.Description.ValueString() {
		t.Errorf("Description: expected %s, got %s", original.Description.ValueString(), result.Description.ValueString())
	}
}

func TestLocalGoProxyResource_Metadata(t *testing.T) {
	r := NewLocalGoProxyResource()
	req := resource.MetadataRequest{ProviderTypeName: "artifactapi"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)
	if resp.TypeName != "artifactapi_local_goproxy" {
		t.Errorf("expected artifactapi_local_goproxy, got %s", resp.TypeName)
	}
}

func TestLocalGoProxyResource_Schema(t *testing.T) {
	r := NewLocalGoProxyResource()
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

func TestNewLocalGoProxyResource_Type(t *testing.T) {
	r := NewLocalGoProxyResource()
	_, ok := r.(*localGoProxyResource)
	if !ok {
		t.Error("expected *localGoProxyResource")
	}
}
