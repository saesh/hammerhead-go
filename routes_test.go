package hammerhead_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	hammerhead "github.com/saesh/hammerhead-go"
)

func TestListRoutes_success(t *testing.T) {
	payload := map[string]any{
		"totalItems":  5,
		"totalPages":  1,
		"perPage":     10,
		"currentPage": 1,
		"data": []map[string]any{
			{"id": "1000.route.abcd", "name": "The Usual Loop", "createdAt": "2025-01-25T12:10:09.409Z", "distance": 2500, "gain": 50},
		},
	}
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/routes" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(payload)
	})

	list, err := client.ListRoutes(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if list.TotalItems != 5 {
		t.Errorf("expected TotalItems=5, got %d", list.TotalItems)
	}
	if len(list.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(list.Routes))
	}
	if list.Routes[0].ID != "1000.route.abcd" {
		t.Errorf("unexpected route ID: %s", list.Routes[0].ID)
	}
}

func TestListRoutes_sendsQueryParams(t *testing.T) {
	var gotQuery string
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	})

	_, _ = client.ListRoutes(context.Background(), &hammerhead.RouteListOptions{
		Page:    3,
		PerPage: 25,
	})

	for _, want := range []string{"page=3", "perPage=25"} {
		if !containsParam(gotQuery, want) {
			t.Errorf("query %q missing param %q", gotQuery, want)
		}
	}
}

func TestCreateRoute_success(t *testing.T) {
	routePayload := map[string]any{
		"id":        "1000.route.abcd",
		"name":      "Uploaded Route",
		"createdAt": "2025-01-25T12:10:09.409Z",
		"updatedAt": "2025-01-25T12:10:09.409Z",
		"distance":  1000,
		"gain":      30,
		"polyline":  "}xq~FhlzvO",
	}
	var gotFilename string
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/routes/file" {
			http.NotFound(w, r)
			return
		}
		_, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err != nil {
				break
			}
			if part.FormName() == "file" {
				gotFilename = part.FileName()
			}
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(routePayload)
	})

	route, err := client.CreateRoute(context.Background(), "loop.gpx", strings.NewReader("<gpx/>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotFilename != "loop.gpx" {
		t.Errorf("expected filename 'loop.gpx', got %q", gotFilename)
	}
	if route.ID != "1000.route.abcd" {
		t.Errorf("unexpected route ID: %s", route.ID)
	}
	if route.Polyline != "}xq~FhlzvO" {
		t.Errorf("unexpected polyline: %s", route.Polyline)
	}
}

func TestUpdateRoute_success(t *testing.T) {
	routePayload := map[string]any{
		"id": "1000.route.abcd", "name": "Updated Route",
		"createdAt": "2025-01-25T12:10:09.409Z", "updatedAt": "2025-01-25T13:00:00.000Z",
		"distance": 2000, "gain": 40, "polyline": "abc",
	}
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/routes/1000.route.abcd/file" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(routePayload)
	})

	route, err := client.UpdateRoute(context.Background(), "1000.route.abcd", "updated.fit", bytes.NewReader([]byte("fit data")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if route.ID != "1000.route.abcd" {
		t.Errorf("unexpected route ID: %s", route.ID)
	}
}

func TestDeleteRoute_success(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/routes/1000.route.abcd" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	if err := client.DeleteRoute(context.Background(), "1000.route.abcd"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteRoute_error(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	})

	err := client.DeleteRoute(context.Background(), "other.route.xyz")

	var apiErr *hammerhead.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", apiErr.StatusCode)
	}
}
