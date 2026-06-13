package hammerhead

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Scope represents an OAuth permission scope.
type Scope string

const (
	ScopeActivityRead Scope = "activity:read"
	ScopeRouteRead    Scope = "route:read"
	ScopeRouteWrite   Scope = "route:write"
	ScopeWorkoutWrite Scope = "workout:write"
)

// Token holds an OAuth token response.
type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	UserID       string `json:"user_id"`
}

// AuthorizeURL builds the OAuth authorization redirect URL.
// Direct users to this URL to begin the authorization code flow.
// The URL is constructed from the client's configured auth base URL so that
// WithAuthURL overrides work correctly for staging or test environments.
func (c *Client) AuthorizeURL(clientID, redirectURI string, scopes []Scope, state string) string {
	scopeStrs := make([]string, len(scopes))
	for i, s := range scopes {
		scopeStrs[i] = string(s)
	}
	params := url.Values{
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
		"response_type": {"code"},
		"scope":         {strings.Join(scopeStrs, " ")},
		"state":         {state},
	}
	return fmt.Sprintf("%s/oauth/authorize?%s", c.authURL, params.Encode())
}

// ExchangeToken exchanges an authorization code for an access token.
func (c *Client) ExchangeToken(ctx context.Context, code, clientID, clientSecret, redirectURI string) (*Token, error) {
	return c.postTokenForm(ctx, url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	})
}

// RefreshToken exchanges a refresh token for a new access token.
func (c *Client) RefreshToken(ctx context.Context, refreshToken, clientID, clientSecret string) (*Token, error) {
	return c.postTokenForm(ctx, url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"refresh_token": {refreshToken},
	})
}

// Deauthorize revokes user authorization, removes all imported data, and invalidates tokens.
func (c *Client) Deauthorize(ctx context.Context, clientID, clientSecret, accessToken string) error {
	form := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"token":         {accessToken},
	}
	req, err := http.NewRequest(http.MethodPost, c.authURL+"/oauth/deauthorize", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.do(ctx, req, nil)
}

func (c *Client) postTokenForm(ctx context.Context, form url.Values) (*Token, error) {
	req, err := http.NewRequest(http.MethodPost, c.authURL+"/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var token Token
	if err := c.do(ctx, req, &token); err != nil {
		return nil, err
	}
	return &token, nil
}
