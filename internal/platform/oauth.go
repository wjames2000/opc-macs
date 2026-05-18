package platform

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OAuth2Config holds the configuration for OAuth2.0 authentication flow.
type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	RedirectURI  string
	Scopes       []string
}

// OAuth2Client implements the standard OAuth2.0 Authorization Code flow.
// It handles AuthURL generation, code exchange, token refresh, and token validation.
type OAuth2Client struct {
	config     OAuth2Config
	httpClient *http.Client
}

// NewOAuth2Client creates a new OAuth2.0 client with the given configuration.
func NewOAuth2Client(config OAuth2Config) *OAuth2Client {
	return &OAuth2Client{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// GenerateState creates a cryptographically random state string for CSRF protection.
func GenerateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate state: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// GetAuthURL generates the authorization URL with optional state parameter.
func (c *OAuth2Client) GetAuthURL(state string) string {
	v := url.Values{}
	v.Set("client_id", c.config.ClientID)
	v.Set("response_type", "code")
	v.Set("redirect_uri", c.config.RedirectURI)
	v.Set("state", state)
	if len(c.config.Scopes) > 0 {
		v.Set("scope", strings.Join(c.config.Scopes, " "))
	}
	return c.config.AuthURL + "?" + v.Encode()
}

// ExchangeCode exchanges an authorization code for an access token.
func (c *OAuth2Client) ExchangeCode(ctx context.Context, code string) (*AuthToken, error) {
	v := url.Values{}
	v.Set("client_id", c.config.ClientID)
	v.Set("client_secret", c.config.ClientSecret)
	v.Set("code", code)
	v.Set("grant_type", "authorization_code")
	v.Set("redirect_uri", c.config.RedirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.TokenURL, strings.NewReader(v.Encode()))
	if err != nil {
		return nil, fmt.Errorf("exchange code: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("exchange code: read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("exchange code: server returned %d: %s", resp.StatusCode, string(body))
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("exchange code: parse response: %w", err)
	}

	return parseTokenResponse(raw), nil
}

// RefreshToken refreshes an expired access token using the refresh token.
func (c *OAuth2Client) RefreshToken(ctx context.Context, token *AuthToken) (*AuthToken, error) {
	if token.RefreshToken == "" {
		return nil, fmt.Errorf("refresh token: no refresh token available")
	}

	v := url.Values{}
	v.Set("client_id", c.config.ClientID)
	v.Set("client_secret", c.config.ClientSecret)
	v.Set("refresh_token", token.RefreshToken)
	v.Set("grant_type", "refresh_token")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.TokenURL, strings.NewReader(v.Encode()))
	if err != nil {
		return nil, fmt.Errorf("refresh token: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("refresh token: read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh token: server returned %d: %s", resp.StatusCode, string(body))
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("refresh token: parse response: %w", err)
	}

	return parseTokenResponse(raw), nil
}

// ValidateToken checks if the token is still valid by making a lightweight API call.
// Returns true if valid, false otherwise. The checkURL should be a lightweight
// endpoint (e.g., user info) that validates the token.
func (c *OAuth2Client) ValidateToken(ctx context.Context, token *AuthToken, checkURL string) (bool, error) {
	if token.IsExpired() {
		return false, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, checkURL, nil)
	if err != nil {
		return false, fmt.Errorf("validate token: create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", token.TokenType, token.AccessToken))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("validate token: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// parseTokenResponse parses an OAuth2.0 token response into an AuthToken.
func parseTokenResponse(raw map[string]any) *AuthToken {
	token := &AuthToken{
		TokenType: "Bearer",
		Raw:       raw,
	}

	if at, ok := raw["access_token"].(string); ok {
		token.AccessToken = at
	}
	if rt, ok := raw["refresh_token"].(string); ok {
		token.RefreshToken = rt
	}
	if tt, ok := raw["token_type"].(string); ok {
		token.TokenType = tt
	}
	if scope, ok := raw["scope"].(string); ok {
		token.Scope = scope
	}
	if expiresIn, ok := raw["expires_in"].(float64); ok {
		token.ExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
	}

	return token
}
