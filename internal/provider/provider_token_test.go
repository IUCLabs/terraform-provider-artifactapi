package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func configuredClient(t *testing.T, endpoint string, token *string) *apiClient {
	t.Helper()
	p := &ArtifactAPIProvider{}
	var schemaResp fwprovider.SchemaResponse
	p.Schema(context.Background(), fwprovider.SchemaRequest{}, &schemaResp)

	tokenVal := tftypes.NewValue(tftypes.String, nil)
	if token != nil {
		tokenVal = tftypes.NewValue(tftypes.String, *token)
	}
	raw := tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"endpoint": tftypes.String,
		"token":    tftypes.String,
	}}, map[string]tftypes.Value{
		"endpoint": tftypes.NewValue(tftypes.String, endpoint),
		"token":    tokenVal,
	})

	var resp fwprovider.ConfigureResponse
	p.Configure(context.Background(), fwprovider.ConfigureRequest{
		Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: raw},
	}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("configure: %v", resp.Diagnostics)
	}
	return resp.ResourceData.(*apiClient)
}

func TestProvider_BearerToken(t *testing.T) {
	strp := func(s string) *string { return &s }
	cases := []struct {
		name string
		attr *string
		env  string
		want string
	}{
		{"attribute", strp("attr-tok"), "", "Bearer attr-tok"},
		{"env fallback", nil, "env-tok", "Bearer env-tok"},
		{"attribute wins over env", strp("attr-tok"), "env-tok", "Bearer attr-tok"},
		{"unset", nil, "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("ARTIFACTAPI_TOKEN", tc.env)
			var got []string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = append(got, r.Header.Get("Authorization"))
				_, present := r.Header["Authorization"]
				if tc.want == "" && present {
					t.Errorf("unexpected Authorization header on %s", r.Method)
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer srv.Close()

			c := configuredClient(t, srv.URL, tc.attr)
			ctx := context.Background()
			if err := c.get(ctx, "/x", nil); err != nil {
				t.Fatal(err)
			}
			if err := c.put(ctx, "/x", map[string]string{"a": "b"}, nil); err != nil {
				t.Fatal(err)
			}
			if err := c.del(ctx, "/x"); err != nil {
				t.Fatal(err)
			}
			for _, h := range got {
				if h != tc.want {
					t.Errorf("Authorization = %q, want %q", h, tc.want)
				}
			}
		})
	}
}
