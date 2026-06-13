package hammerhead_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"testing"

	hammerhead "github.com/saesh/hammerhead-go"
)

func TestListActivities_success(t *testing.T) {
	payload := map[string]any{
		"totalItems":  34,
		"totalPages":  4,
		"perPage":     10,
		"currentPage": 1,
		"data": []map[string]any{
			{"id": "1000.activity.abcd", "name": "My Epic Ride", "createdAt": "2025-01-25T12:10:09.409Z", "duration": 76765, "distance": 123.45},
		},
	}
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/activities" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(payload)
	})

	list, err := client.ListActivities(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if list.TotalItems != 34 {
		t.Errorf("expected TotalItems=34, got %d", list.TotalItems)
	}
	if len(list.Activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(list.Activities))
	}
	if list.Activities[0].ID != "1000.activity.abcd" {
		t.Errorf("unexpected activity ID: %s", list.Activities[0].ID)
	}
}

func TestListActivities_sendsQueryParams(t *testing.T) {
	var gotQuery string
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	})

	_, _ = client.ListActivities(context.Background(), &hammerhead.ActivityListOptions{
		Page:      2,
		PerPage:   50,
		StartDate: "2025-01-01",
	})

	for _, want := range []string{"page=2", "perPage=50", "startDate=2025-01-01"} {
		if !containsParam(gotQuery, want) {
			t.Errorf("query %q missing param %q", gotQuery, want)
		}
	}
}

func TestListActivities_error(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})

	_, err := client.ListActivities(context.Background(), nil)

	var apiErr *hammerhead.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", apiErr.StatusCode)
	}
}

func TestGetActivity_success(t *testing.T) {
	payload := map[string]any{
		"id":           "1000.activity.abcd",
		"name":         "My Epic Ride",
		"createdAt":    "2025-01-25T12:10:09.409Z",
		"updatedAt":    "2025-01-25T12:10:09.409Z",
		"duration":     76765,
		"distance":     123.45,
		"activityType": "RIDE",
		"description":  "I fell clipping out",
		"polyline":     "}xq~FhlzvO",
	}
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/activities/1000.activity.abcd" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(payload)
	})

	activity, err := client.GetActivity(context.Background(), "1000.activity.abcd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activity.ID != "1000.activity.abcd" {
		t.Errorf("unexpected ID: %s", activity.ID)
	}
	if activity.ActivityType != hammerhead.ActivityTypeRide {
		t.Errorf("unexpected type: %s", activity.ActivityType)
	}
	if activity.Polyline != "}xq~FhlzvO" {
		t.Errorf("unexpected polyline: %s", activity.Polyline)
	}
}

func TestGetActivity_notFound(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	_, err := client.GetActivity(context.Background(), "missing")

	var apiErr *hammerhead.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", apiErr.StatusCode)
	}
}

func TestGetActivityFile_success(t *testing.T) {
	fitData := []byte{0x0E, 0x10, 0xD9, 0x07} // FIT file magic bytes
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/activities/1000.activity.abcd/file" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/vnd.ant.fit")
		w.Write(fitData)
	})

	got, err := client.GetActivityFile(context.Background(), "1000.activity.abcd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(fitData) {
		t.Errorf("unexpected file bytes")
	}
}

func containsParam(query, param string) bool {
	return slices.Contains(splitParams(query), param)
}

func splitParams(query string) []string {
	if query == "" {
		return nil
	}
	var params []string
	start := 0
	for i := 0; i <= len(query); i++ {
		if i == len(query) || query[i] == '&' {
			params = append(params, query[start:i])
			start = i + 1
		}
	}
	return params
}
