package hammerhead

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Workout represents a structured training workout.
type Workout struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PlannedDate string    `json:"plannedDate"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CreateWorkout creates a new workout from the given file.
// filename must have a supported extension: .fit, .zwo.
// plannedDate is optional; pass "" to omit. Format: YYYY-MM-DD.
func (c *Client) CreateWorkout(ctx context.Context, filename string, r io.Reader, plannedDate string) (*Workout, error) {
	body, contentType, err := buildFileUpload(filename, r)
	if err != nil {
		return nil, err
	}

	u := c.baseURL + "/workouts/file"
	if plannedDate != "" {
		u += "?" + url.Values{"plannedDate": {plannedDate}}.Encode()
	}

	req, err := http.NewRequest(http.MethodPost, u, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	var result Workout
	if err := c.do(ctx, req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateWorkout replaces an existing workout's file content.
// filename must have a supported extension: .fit, .zwo.
// plannedDate is optional; pass "" to omit. Format: YYYY-MM-DD.
func (c *Client) UpdateWorkout(ctx context.Context, workoutID, filename string, r io.Reader, plannedDate string) (*Workout, error) {
	body, contentType, err := buildFileUpload(filename, r)
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf("%s/workouts/%s/file", c.baseURL, url.PathEscape(workoutID))
	if plannedDate != "" {
		u += "?" + url.Values{"plannedDate": {plannedDate}}.Encode()
	}

	req, err := http.NewRequest(http.MethodPut, u, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	var result Workout
	if err := c.do(ctx, req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteWorkout deletes a workout by ID. Only workouts created by your client can be deleted.
func (c *Client) DeleteWorkout(ctx context.Context, workoutID string) error {
	u := fmt.Sprintf("%s/workouts/%s", c.baseURL, url.PathEscape(workoutID))
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	return c.do(ctx, req, nil)
}
