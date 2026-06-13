// Package hammerhead provides a Go client for the Hammerhead Karoo API.
// See https://api.hammerhead.io/v1/docs for API documentation.
package hammerhead

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://api.hammerhead.io/v1/api"
	defaultAuthURL = "https://api.hammerhead.io/v1/auth"
	defaultTimeout = 30 * time.Second
)

// Page holds pagination metadata returned by list endpoints.
type Page struct {
	TotalItems  int `json:"totalItems"`
	TotalPages  int `json:"totalPages"`
	PerPage     int `json:"perPage"`
	CurrentPage int `json:"currentPage"`
}

// Client is the Hammerhead API client. Use NewClient to create one.
type Client struct {
	httpClient *http.Client
	baseURL    string
	authURL    string
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithAuthURL overrides the OAuth base URL.
func WithAuthURL(url string) Option {
	return func(c *Client) { c.authURL = url }
}

// WithHTTPClient provides a custom *http.Client whose Transport is wrapped
// with bearer-token injection. The client's Timeout, CheckRedirect, and Jar
// are preserved; its Transport (or http.DefaultTransport if nil) becomes the
// inner transport of the bearer wrapper.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// NewClient creates a Client that authenticates every request with accessToken.
// Returns an error if accessToken is empty.
func NewClient(accessToken string, opts ...Option) (*Client, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("hammerhead: accessToken must not be empty")
	}
	c := &Client{
		baseURL: defaultBaseURL,
		authURL: defaultAuthURL,
	}
	for _, opt := range opts {
		opt(c)
	}

	// Resolve base transport and client settings, preserving any WithHTTPClient config.
	var base http.RoundTripper = http.DefaultTransport
	timeout := defaultTimeout
	var redirect func(*http.Request, []*http.Request) error
	var jar http.CookieJar

	if c.httpClient != nil {
		if c.httpClient.Transport != nil {
			base = c.httpClient.Transport
		}
		if c.httpClient.Timeout > 0 {
			timeout = c.httpClient.Timeout
		}
		redirect = c.httpClient.CheckRedirect
		jar = c.httpClient.Jar
	}

	c.httpClient = &http.Client{
		Transport:     &bearerTransport{token: accessToken, base: base},
		Timeout:       timeout,
		CheckRedirect: redirect,
		Jar:           jar,
	}
	return c, nil
}

// bearerTransport injects an Authorization: Bearer header into every request.
type bearerTransport struct {
	token string
	base  http.RoundTripper
}

func (t *bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(clone)
}

// APIError is returned when the Hammerhead API responds with a non-2xx status.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("hammerhead: %d %s", e.StatusCode, e.Message)
}

// do executes req and decodes a JSON response into v (if non-nil).
func (c *Client) do(ctx context.Context, req *http.Request, v any) error {
	req = req.WithContext(ctx)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return &APIError{StatusCode: resp.StatusCode, Message: string(body)}
	}
	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
	}
	return nil
}

// doRaw executes req and returns the raw response body bytes.
func (c *Client) doRaw(ctx context.Context, req *http.Request) ([]byte, error) {
	req = req.WithContext(ctx)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{StatusCode: resp.StatusCode, Message: string(body)}
	}
	return io.ReadAll(resp.Body)
}
