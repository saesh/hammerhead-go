package hammerhead_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	hammerhead "github.com/saesh/hammerhead-go"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) (*hammerhead.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client, err := hammerhead.NewClient("test-token",
		hammerhead.WithBaseURL(srv.URL),
		hammerhead.WithAuthURL(srv.URL),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client, srv
}

func TestNewClient_injectsBearer(t *testing.T) {
	var gotAuth string
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	})

	_ = client.DeleteRoute(context.Background(), "any-id")

	if gotAuth != "Bearer test-token" {
		t.Errorf("expected 'Bearer test-token', got %q", gotAuth)
	}
}

func TestAPIError_format(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})

	err := client.DeleteRoute(context.Background(), "missing")

	var apiErr *hammerhead.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
}
