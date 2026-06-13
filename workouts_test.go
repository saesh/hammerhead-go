package hammerhead_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestCreateWorkout_success(t *testing.T) {
	workoutPayload := map[string]any{
		"id":          "1000.workout.abcd",
		"name":        "Leg Crusher 2025",
		"description": "Actually crushed my legs",
		"plannedDate": "2025-01-26",
		"createdAt":   "2025-01-25T12:10:09.409Z",
		"updatedAt":   "2025-01-25T12:10:09.409Z",
	}
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/workouts/file" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(workoutPayload)
	})

	workout, err := client.CreateWorkout(context.Background(), "workout.zwo", strings.NewReader("<workout/>"), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if workout.ID != "1000.workout.abcd" {
		t.Errorf("unexpected workout ID: %s", workout.ID)
	}
	if workout.Name != "Leg Crusher 2025" {
		t.Errorf("unexpected workout name: %s", workout.Name)
	}
}

func TestCreateWorkout_withPlannedDate(t *testing.T) {
	var gotQuery string
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"id": "1000.workout.abcd", "name": "Test",
			"createdAt": "2025-01-25T12:10:09.409Z", "updatedAt": "2025-01-25T12:10:09.409Z",
		})
	})

	_, _ = client.CreateWorkout(context.Background(), "workout.fit", strings.NewReader("fit"), "2025-01-26")

	if !containsParam(gotQuery, "plannedDate=2025-01-26") {
		t.Errorf("query %q missing plannedDate param", gotQuery)
	}
}

func TestUpdateWorkout_success(t *testing.T) {
	workoutPayload := map[string]any{
		"id": "1000.workout.abcd", "name": "Updated",
		"createdAt": "2025-01-25T12:10:09.409Z", "updatedAt": "2025-01-26T08:00:00.000Z",
	}
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/workouts/1000.workout.abcd/file" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(workoutPayload)
	})

	workout, err := client.UpdateWorkout(context.Background(), "1000.workout.abcd", "updated.zwo", strings.NewReader("<workout/>"), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if workout.ID != "1000.workout.abcd" {
		t.Errorf("unexpected ID: %s", workout.ID)
	}
}

func TestDeleteWorkout_success(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/workouts/1000.workout.abcd" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	if err := client.DeleteWorkout(context.Background(), "1000.workout.abcd"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
