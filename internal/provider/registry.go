package provider

import (
	"context"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	resourceRegistry   []func() resource.Resource
	dataSourceRegistry []func() datasource.DataSource
)

func registerResource(f func() resource.Resource) {
	resourceRegistry = append(resourceRegistry, f)
}

func registerDataSource(f func() datasource.DataSource) {
	dataSourceRegistry = append(dataSourceRegistry, f)
}

func resourceTypeName(f func() resource.Resource) string {
	var resp resource.MetadataResponse
	f().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "artifactapi"}, &resp)
	return resp.TypeName
}

func dataSourceTypeName(f func() datasource.DataSource) string {
	var resp datasource.MetadataResponse
	f().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "artifactapi"}, &resp)
	return resp.TypeName
}

func sortedByTypeName[T any](fs []T, name func(T) string) []T {
	out := slices.Clone(fs)
	slices.SortFunc(out, func(a, b T) int { return strings.Compare(name(a), name(b)) })
	return out
}
