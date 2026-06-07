package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAPIClient(t *testing.T) {
	c := newAPIClient("http://example.com")
	if c.baseURL != "http://example.com" {
		t.Errorf("expected baseURL http://example.com, got %s", c.baseURL)
	}
	if c.httpClient == nil {
		t.Error("expected non-nil httpClient")
	}
}

func TestGet_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v2/remotes/test" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"name": "test"})
	}))
	defer srv.Close()

	c := newAPIClient(srv.URL)
	var out map[string]string
	err := c.get(context.Background(), "/api/v2/remotes/test", &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["name"] != "test" {
		t.Errorf("expected name=test, got %s", out["name"])
	}
}

func TestGet_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newAPIClient(srv.URL)
	var out map[string]string
	err := c.get(context.Background(), "/api/v2/remotes/missing", &out)
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !isNotFound(err) {
		t.Errorf("expected notFoundError, got %T: %v", err, err)
	}
}

func TestGet_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "internal server error")
	}))
	defer srv.Close()

	c := newAPIClient(srv.URL)
	var out map[string]string
	err := c.get(context.Background(), "/api/v2/remotes/fail", &out)
	if err == nil {
		t.Fatal("expected error for 500")
	}
	if isNotFound(err) {
		t.Error("500 should not be notFoundError")
	}
	expected := "api error 500: internal server error"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func TestPost_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("expected Content-Type application/json")
		}
		// Decode the request body and echo it back
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(body)
	}))
	defer srv.Close()

	c := newAPIClient(srv.URL)
	input := map[string]string{"name": "new-remote"}
	var out map[string]string
	err := c.post(context.Background(), "/api/v2/remotes", input, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["name"] != "new-remote" {
		t.Errorf("expected name=new-remote, got %s", out["name"])
	}
}

func TestPut_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(body)
	}))
	defer srv.Close()

	c := newAPIClient(srv.URL)
	input := map[string]string{"name": "updated"}
	var out map[string]string
	err := c.put(context.Background(), "/api/v2/remotes/updated", input, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["name"] != "updated" {
		t.Errorf("expected name=updated, got %s", out["name"])
	}
}

func TestDel_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newAPIClient(srv.URL)
	err := c.del(context.Background(), "/api/v2/remotes/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDel_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newAPIClient(srv.URL)
	err := c.del(context.Background(), "/api/v2/remotes/missing")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !isNotFound(err) {
		t.Errorf("expected notFoundError, got %T", err)
	}
}

func TestDo_BadRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "invalid input")
	}))
	defer srv.Close()

	c := newAPIClient(srv.URL)
	err := c.do(context.Background(), http.MethodPost, "/api/v2/remotes", map[string]string{"bad": "data"}, nil)
	if err == nil {
		t.Fatal("expected error for 400")
	}
	expected := "api error 400: invalid input"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestDo_NoContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newAPIClient(srv.URL)
	var out map[string]string
	// Should not attempt to decode 204 even with out pointer
	err := c.do(context.Background(), http.MethodDelete, "/test", nil, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDo_JSONEncoding(t *testing.T) {
	type payload struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var p payload
		if err := json.Unmarshal(body, &p); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if p.Name != "test" || p.Count != 42 {
			t.Errorf("unexpected payload: %+v", p)
		}
		// Return a different payload
		resp := payload{Name: "response", Count: 99}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newAPIClient(srv.URL)
	input := payload{Name: "test", Count: 42}
	var out payload
	err := c.post(context.Background(), "/test", input, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Name != "response" || out.Count != 99 {
		t.Errorf("unexpected response: %+v", out)
	}
}

func TestNotFoundError_Error(t *testing.T) {
	e := &notFoundError{path: "/api/v2/remotes/test"}
	expected := "not found: /api/v2/remotes/test"
	if e.Error() != expected {
		t.Errorf("expected %q, got %q", expected, e.Error())
	}
}

func TestIsNotFound_True(t *testing.T) {
	err := &notFoundError{path: "/test"}
	if !isNotFound(err) {
		t.Error("expected isNotFound to return true for *notFoundError")
	}
}

func TestIsNotFound_False(t *testing.T) {
	err := fmt.Errorf("some other error")
	if isNotFound(err) {
		t.Error("expected isNotFound to return false for non-notFoundError")
	}
}

func TestIsNotFound_Nil(t *testing.T) {
	if isNotFound(nil) {
		t.Error("expected isNotFound to return false for nil")
	}
}

func TestIsNotFound_WrappedError(t *testing.T) {
	// isNotFound uses type assertion, not errors.As, so a wrapped notFoundError should NOT match
	inner := &notFoundError{path: "/test"}
	wrapped := fmt.Errorf("wrapped: %w", inner)
	// The current implementation uses type assertion, so wrapped should be false
	if isNotFound(wrapped) {
		t.Error("expected isNotFound to return false for wrapped notFoundError (uses type assertion)")
	}
	// But errors.As would find it
	var nfe *notFoundError
	if !errors.As(wrapped, &nfe) {
		t.Error("errors.As should find wrapped notFoundError")
	}
}

func TestGet_DecodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "not valid json")
	}))
	defer srv.Close()

	c := newAPIClient(srv.URL)
	var out map[string]string
	err := c.get(context.Background(), "/test", &out)
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

func TestDo_NilBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "" {
			t.Error("Content-Type should not be set for nil body")
		}
		if r.Body != nil {
			body, _ := io.ReadAll(r.Body)
			if len(body) > 0 {
				t.Errorf("expected empty body, got %s", string(body))
			}
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "{}")
	}))
	defer srv.Close()

	c := newAPIClient(srv.URL)
	var out map[string]string
	err := c.get(context.Background(), "/test", &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
