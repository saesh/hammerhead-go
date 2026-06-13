package hammerhead

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// RouteSummary is a lightweight route record returned by list endpoints.
type RouteSummary struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	Distance  float64   `json:"distance"`
	Gain      float64   `json:"gain"`
}

// Route is a full route record.
type Route struct {
	RouteSummary
	UpdatedAt time.Time `json:"updatedAt"`
	Polyline  string    `json:"polyline"`
}

// RouteList is a paginated list of route summaries.
type RouteList struct {
	Page
	Routes []RouteSummary `json:"data"`
}

// RouteListOptions paginates a ListRoutes request.
// Zero values are omitted and the API defaults apply.
type RouteListOptions struct {
	Page    int
	PerPage int // max 100
}

// ListRoutes returns a paginated list of route summaries.
// opts may be nil to use API defaults (page 1, 10 per page).
func (c *Client) ListRoutes(ctx context.Context, opts *RouteListOptions) (*RouteList, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PerPage > 0 {
			params.Set("perPage", strconv.Itoa(opts.PerPage))
		}
	}

	u := c.baseURL + "/routes"
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	var result RouteList
	if err := c.do(ctx, req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateRoute creates a new route from the given file.
// filename must have a supported extension: .gpx, .fit, .tcx, .kml, .kmz.
func (c *Client) CreateRoute(ctx context.Context, filename string, r io.Reader) (*Route, error) {
	body, contentType, err := buildFileUpload(filename, r)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/routes/file", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	var result Route
	if err := c.do(ctx, req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateRoute replaces an existing route's file content.
// filename must have a supported extension: .gpx, .fit, .tcx, .kml, .kmz.
func (c *Client) UpdateRoute(ctx context.Context, routeID, filename string, r io.Reader) (*Route, error) {
	body, contentType, err := buildFileUpload(filename, r)
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf("%s/routes/%s/file", c.baseURL, url.PathEscape(routeID))
	req, err := http.NewRequest(http.MethodPut, u, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	var result Route
	if err := c.do(ctx, req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteRoute deletes a route by ID. Only routes created by your client can be deleted.
func (c *Client) DeleteRoute(ctx context.Context, routeID string) error {
	u := fmt.Sprintf("%s/routes/%s", c.baseURL, url.PathEscape(routeID))
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	return c.do(ctx, req, nil)
}

// buildFileUpload encodes r as a multipart/form-data body with the field name "file".
func buildFileUpload(filename string, r io.Reader) (*bytes.Buffer, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		return nil, "", err
	}
	if _, err := io.Copy(fw, r); err != nil {
		return nil, "", err
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return &buf, w.FormDataContentType(), nil
}
