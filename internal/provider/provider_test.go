package provider

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
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

func TestProvider_Resources_Invariants(t *testing.T) {
	p := &ArtifactAPIProvider{version: "1.0.0"}
	var names []string
	for _, f := range p.Resources(context.Background()) {
		names = append(names, resourceTypeName(f))
	}
	checkRegistered(t, names, "resources")
}

func TestProvider_DataSources_Invariants(t *testing.T) {
	p := &ArtifactAPIProvider{version: "1.0.0"}
	var names []string
	for _, f := range p.DataSources(context.Background()) {
		names = append(names, dataSourceTypeName(f))
	}
	checkRegistered(t, names, "data-sources")
}

var exampleFile = map[string]string{"resources": "resource.tf", "data-sources": "data-source.tf"}

func checkRegistered(t *testing.T, names []string, kind string) {
	t.Helper()
	if len(names) == 0 {
		t.Fatalf("no %s registered", kind)
	}
	if !slices.IsSorted(names) {
		t.Errorf("%s not sorted by type name: %v", kind, names)
	}
	seen := map[string]bool{}
	for _, n := range names {
		if !strings.HasPrefix(n, "artifactapi_") {
			t.Errorf("%s: type name %q lacks artifactapi_ prefix", kind, n)
		}
		if seen[n] {
			t.Errorf("%s: duplicate type name %q", kind, n)
		}
		seen[n] = true
		if _, err := os.Stat(filepath.Join("../../examples", kind, n, exampleFile[kind])); err != nil {
			t.Errorf("%s: %s has no example: %v", kind, n, err)
		}
		if _, err := os.Stat(filepath.Join("../../docs", kind, strings.TrimPrefix(n, "artifactapi_")+".md")); err != nil {
			t.Errorf("%s: %s has no docs page: %v", kind, n, err)
		}
	}

	dirs, err := os.ReadDir(filepath.Join("../../examples", kind))
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range dirs {
		if !seen[d.Name()] {
			t.Errorf("examples/%s/%s does not match a registered type", kind, d.Name())
		}
	}
	docs, err := filepath.Glob(filepath.Join("../../docs", kind, "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range docs {
		if n := "artifactapi_" + strings.TrimSuffix(filepath.Base(d), ".md"); !seen[n] {
			t.Errorf("docs/%s/%s does not match a registered type", kind, filepath.Base(d))
		}
	}
}

func TestProvider_ImplementsInterface(t *testing.T) {
	var _ fwprovider.Provider = &ArtifactAPIProvider{}
}
