package hammerhead_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	hammerhead "github.com/saesh/hammerhead-go"
)

func TestAuthorizeURL_buildsCorrectURL(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})
	got := client.AuthorizeURL(
		"my-client-id",
		"https://example.com/callback",
		[]hammerhead.Scope{hammerhead.ScopeActivityRead, hammerhead.ScopeRouteRead},
		"random-state",
	)

	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("invalid URL: %v", err)
	}
	q := u.Query()

	if q.Get("client_id") != "my-client-id" {
		t.Errorf("unexpected client_id: %s", q.Get("client_id"))
	}
	if q.Get("redirect_uri") != "https://example.com/callback" {
		t.Errorf("unexpected redirect_uri: %s", q.Get("redirect_uri"))
	}
	if q.Get("response_type") != "code" {
		t.Errorf("unexpected response_type: %s", q.Get("response_type"))
	}
	if q.Get("state") != "random-state" {
		t.Errorf("unexpected state: %s", q.Get("state"))
	}
	scope := q.Get("scope")
	if !strings.Contains(scope, "activity:read") || !strings.Contains(scope, "route:read") {
		t.Errorf("unexpected scope: %s", scope)
	}
}

func TestExchangeToken_success(t *testing.T) {
	var gotForm url.Values
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/token" {
			http.NotFound(w, r)
			return
		}
		body, _ := io.ReadAll(r.Body)
		gotForm, _ = url.ParseQuery(string(body))
		json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "new-access-token",
			"refresh_token": "new-refresh-token",
			"token_type":    "Bearer",
			"expires_in":    52000,
			"user_id":       "1000",
		})
	})

	token, err := client.ExchangeToken(context.Background(), "auth-code", "client-id", "client-secret", "https://example.com/cb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "new-access-token" {
		t.Errorf("unexpected access token: %s", token.AccessToken)
	}
	if token.UserID != "1000" {
		t.Errorf("unexpected user ID: %s", token.UserID)
	}
	if gotForm.Get("grant_type") != "authorization_code" {
		t.Errorf("unexpected grant_type: %s", gotForm.Get("grant_type"))
	}
	if gotForm.Get("code") != "auth-code" {
		t.Errorf("unexpected code: %s", gotForm.Get("code"))
	}
	if gotForm.Get("client_id") != "client-id" {
		t.Errorf("unexpected client_id: %s", gotForm.Get("client_id"))
	}
}

func TestRefreshToken_success(t *testing.T) {
	var gotForm url.Values
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotForm, _ = url.ParseQuery(string(body))
		json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "refreshed-token",
			"refresh_token": "new-refresh",
			"token_type":    "Bearer",
			"expires_in":    52000,
			"user_id":       "1000",
		})
	})

	token, err := client.RefreshToken(context.Background(), "old-refresh", "client-id", "client-secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "refreshed-token" {
		t.Errorf("unexpected access token: %s", token.AccessToken)
	}
	if gotForm.Get("grant_type") != "refresh_token" {
		t.Errorf("unexpected grant_type: %s", gotForm.Get("grant_type"))
	}
	if gotForm.Get("refresh_token") != "old-refresh" {
		t.Errorf("unexpected refresh_token: %s", gotForm.Get("refresh_token"))
	}
}

func TestDeauthorize_success(t *testing.T) {
	var gotForm url.Values
	client, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/deauthorize" {
			http.NotFound(w, r)
			return
		}
		body, _ := io.ReadAll(r.Body)
		gotForm, _ = url.ParseQuery(string(body))
		w.WriteHeader(http.StatusOK)
	})

	err := client.Deauthorize(context.Background(), "client-id", "client-secret", "access-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotForm.Get("client_id") != "client-id" {
		t.Errorf("unexpected client_id: %s", gotForm.Get("client_id"))
	}
	if gotForm.Get("token") != "access-token" {
		t.Errorf("unexpected token: %s", gotForm.Get("token"))
	}
}
