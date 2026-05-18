package platform

import (
	"context"
	"fmt"
)

// APIKeyAuth implements authentication for platforms that use API Key-based auth.
type APIKeyAuth struct {
	APIKey     string
	APISecret  string
	HeaderName string // custom header name for the API key (default: "X-API-Key")
}

// NewAPIKeyAuth creates a new API Key authentication adapter.
func NewAPIKeyAuth(apiKey, apiSecret string) *APIKeyAuth {
	return &APIKeyAuth{
		APIKey:     apiKey,
		APISecret:  apiSecret,
		HeaderName: "X-API-Key",
	}
}

// GetHeaders returns the HTTP headers for API Key authentication.
func (a *APIKeyAuth) GetHeaders() map[string]string {
	return map[string]string{
		a.HeaderName: a.APIKey,
	}
}

// Authenticate performs an API Key authentication check.
func (a *APIKeyAuth) Authenticate(ctx context.Context) error {
	if a.APIKey == "" {
		return fmt.Errorf("api key auth: api key is required")
	}
	return nil
}

// CookieAuth implements authentication for platforms that use Cookie-based auth.
// This is typically used for platforms without official APIs (scraping approach).
type CookieAuth struct {
	Cookies      map[string]string
	UserAgent    string
	ExtraHeaders map[string]string
}

// NewCookieAuth creates a new Cookie authentication adapter.
func NewCookieAuth(cookies map[string]string) *CookieAuth {
	return &CookieAuth{
		Cookies:      cookies,
		UserAgent:    "OPC-Agent/1.0",
		ExtraHeaders: make(map[string]string),
	}
}

// GetCookieString returns the cookies as a single HTTP Cookie header string.
func (a *CookieAuth) GetCookieString() string {
	var cookieStr string
	for name, value := range a.Cookies {
		if cookieStr != "" {
			cookieStr += "; "
		}
		cookieStr += fmt.Sprintf("%s=%s", name, value)
	}
	return cookieStr
}

// GetHeaders returns all HTTP headers including cookies.
func (a *CookieAuth) GetHeaders() map[string]string {
	headers := make(map[string]string)
	for k, v := range a.ExtraHeaders {
		headers[k] = v
	}
	headers["Cookie"] = a.GetCookieString()
	headers["User-Agent"] = a.UserAgent
	return headers
}

// Validate checks that cookies are not empty.
func (a *CookieAuth) Validate() error {
	if len(a.Cookies) == 0 {
		return fmt.Errorf("cookie auth: at least one cookie is required")
	}
	return nil
}
