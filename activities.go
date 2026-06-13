package hammerhead

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// ActivityType identifies the kind of activity recorded.
type ActivityType string

const (
	ActivityTypeRide          ActivityType = "RIDE"
	ActivityTypeEBike         ActivityType = "EBIKE"
	ActivityTypeMountainBike  ActivityType = "MOUNTAIN_BIKE"
	ActivityTypeGravel        ActivityType = "GRAVEL"
	ActivityTypeEMountainBike ActivityType = "EMOUNTAIN_BIKE"
	ActivityTypeVelomobile    ActivityType = "VELOMOBILE"
)

// ActivitySummary is a lightweight activity record returned by list endpoints.
type ActivitySummary struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	Duration  float64   `json:"duration"`
	Distance  float64   `json:"distance"`
}

// Activity is a full activity record including GPS and metadata.
type Activity struct {
	ActivitySummary
	ActivityType ActivityType `json:"activityType"`
	Description  string       `json:"description"`
	Polyline     string       `json:"polyline"`
	UpdatedAt    time.Time    `json:"updatedAt"`
}

// ActivityList is a paginated list of activity summaries.
type ActivityList struct {
	Page
	Activities []ActivitySummary `json:"data"`
}

// ActivityListOptions filters and paginates a ListActivities request.
// Zero values are omitted and the API defaults apply.
type ActivityListOptions struct {
	Page      int
	PerPage   int    // max 100
	StartDate string // YYYY-MM-DD
}

// ListActivities returns a paginated list of activity summaries.
// opts may be nil to use API defaults (page 1, 10 per page).
func (c *Client) ListActivities(ctx context.Context, opts *ActivityListOptions) (*ActivityList, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PerPage > 0 {
			params.Set("perPage", strconv.Itoa(opts.PerPage))
		}
		if opts.StartDate != "" {
			params.Set("startDate", opts.StartDate)
		}
	}

	u := c.baseURL + "/activities"
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	var result ActivityList
	if err := c.do(ctx, req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetActivity returns a single activity by ID.
func (c *Client) GetActivity(ctx context.Context, activityID string) (*Activity, error) {
	u := fmt.Sprintf("%s/activities/%s", c.baseURL, url.PathEscape(activityID))
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	var result Activity
	if err := c.do(ctx, req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetActivityFile downloads the FIT file for an activity and returns the raw bytes.
func (c *Client) GetActivityFile(ctx context.Context, activityID string) ([]byte, error) {
	u := fmt.Sprintf("%s/activities/%s/file", c.baseURL, url.PathEscape(activityID))
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	return c.doRaw(ctx, req)
}
