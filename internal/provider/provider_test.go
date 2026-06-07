package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestNew_ReturnsFactory(t *testing.T) {
	factory := New("1.0.0")
	if factory == nil {
		t.Fatal("expected non-nil factory")
	}
	p := factory()
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
	_, ok := p.(*ArtifactAPIProvider)
	if !ok {
		t.Fatalf("expected *ArtifactAPIProvider, got %T", p)
	}
}

func TestNew_VersionPropagated(t *testing.T) {
	factory := New("2.3.4")
	p := factory().(*ArtifactAPIProvider)
	if p.version != "2.3.4" {
		t.Errorf("expected version 2.3.4, got %s", p.version)
	}
}

func TestProvider_Metadata(t *testing.T) {
	p := &ArtifactAPIProvider{version: "1.2.3"}
	req := fwprovider.MetadataRequest{}
	var resp fwprovider.MetadataResponse
	p.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "artifactapi" {
		t.Errorf("expected type name 'artifactapi', got %s", resp.TypeName)
	}
	if resp.Version != "1.2.3" {
		t.Errorf("expected version '1.2.3', got %s", resp.Version)
	}
}

func TestProvider_Schema(t *testing.T) {
	p := &ArtifactAPIProvider{version: "1.0.0"}
	req := fwprovider.SchemaRequest{}
	var resp fwprovider.SchemaResponse
	p.Schema(context.Background(), req, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema description")
	}

	_, ok := resp.Schema.Attributes["endpoint"]
	if !ok {
		t.Fatal("missing 'endpoint' attribute in schema")
	}
}

func TestProvider_Resources(t *testing.T) {
	p := &ArtifactAPIProvider{version: "1.0.0"}
	resources := p.Resources(context.Background())

	// 10 remote resource types + 1 virtual = 11
	expectedCount := 11
	if len(resources) != expectedCount {
		t.Fatalf("expected %d resources, got %d", expectedCount, len(resources))
	}

	// Verify each factory produces a valid resource
	for i, factory := range resources {
		r := factory()
		if r == nil {
			t.Errorf("resource factory %d returned nil", i)
		}
	}
}

func TestProvider_Resources_ContainsExpectedTypes(t *testing.T) {
	p := &ArtifactAPIProvider{version: "1.0.0"}
	resources := p.Resources(context.Background())

	// Collect type names by calling Metadata on each
	typeNames := make(map[string]bool)
	for _, factory := range resources {
		r := factory()
		req := resource.MetadataRequest{ProviderTypeName: "artifactapi"}
		var resp resource.MetadataResponse
		r.Metadata(context.Background(), req, &resp)
		typeNames[resp.TypeName] = true
	}

	expected := []string{
		"artifactapi_remote_generic",
		"artifactapi_remote_docker",
		"artifactapi_remote_helm",
		"artifactapi_remote_pypi",
		"artifactapi_remote_npm",
		"artifactapi_remote_rpm",
		"artifactapi_remote_alpine",
		"artifactapi_remote_puppet",
		"artifactapi_remote_terraform",
		"artifactapi_remote_goproxy",
		"artifactapi_virtual",
	}

	for _, name := range expected {
		if !typeNames[name] {
			t.Errorf("missing expected resource type: %s", name)
		}
	}
}

func TestProvider_DataSources(t *testing.T) {
	p := &ArtifactAPIProvider{version: "1.0.0"}
	dataSources := p.DataSources(context.Background())

	expectedCount := 2
	if len(dataSources) != expectedCount {
		t.Fatalf("expected %d data sources, got %d", expectedCount, len(dataSources))
	}

	// Verify each factory produces a valid data source
	for i, factory := range dataSources {
		ds := factory()
		if ds == nil {
			t.Errorf("data source factory %d returned nil", i)
		}
	}
}

func TestProvider_DataSources_ContainsExpectedTypes(t *testing.T) {
	p := &ArtifactAPIProvider{version: "1.0.0"}
	dataSources := p.DataSources(context.Background())

	typeNames := make(map[string]bool)
	for _, factory := range dataSources {
		ds := factory()
		req := datasource.MetadataRequest{ProviderTypeName: "artifactapi"}
		var resp datasource.MetadataResponse
		ds.Metadata(context.Background(), req, &resp)
		typeNames[resp.TypeName] = true
	}

	expected := []string{
		"artifactapi_remote",
		"artifactapi_virtual",
	}

	for _, name := range expected {
		if !typeNames[name] {
			t.Errorf("missing expected data source type: %s", name)
		}
	}
}

func TestProvider_ImplementsInterface(t *testing.T) {
	var _ fwprovider.Provider = &ArtifactAPIProvider{}
}
