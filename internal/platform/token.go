package platform

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// TokenStore defines the interface for persisting auth tokens.
type TokenStore interface {
	// Get retrieves a token for the given account ID.
	Get(ctx context.Context, accountID string) (*AuthToken, error)
	// Save stores a token for the given account ID.
	Save(ctx context.Context, accountID string, token *AuthToken) error
	// Delete removes a token for the given account ID.
	Delete(ctx context.Context, accountID string) error
	// List returns all account IDs that have stored tokens.
	List(ctx context.Context) ([]string, error)
}

// TokenManager handles automatic token refresh, expiration warnings, and persistence.
// It wraps a PlatformClient's token operations with caching and auto-refresh logic.
type TokenManager struct {
	store       TokenStore
	client      PlatformClient
	cache       map[string]*cachedToken
	mu          sync.RWMutex
	alertBefore time.Duration // how long before expiration to trigger warning
}

type cachedToken struct {
	token     *AuthToken
	expiresAt time.Time
}

// NewTokenManager creates a new TokenManager for the given platform client.
func NewTokenManager(store TokenStore, client PlatformClient) *TokenManager {
	return &TokenManager{
		store:       store,
		client:      client,
		cache:       make(map[string]*cachedToken),
		alertBefore: 24 * time.Hour, // warn 24 hours before expiration by default
	}
}

// SetAlertThreshold sets how far before expiration to trigger a warning.
func (tm *TokenManager) SetAlertThreshold(d time.Duration) {
	tm.alertBefore = d
}

// GetToken retrieves a valid token for the given account.
// It checks the in-memory cache first, then the persistent store.
// If the token is expired, it attempts auto-refresh.
func (tm *TokenManager) GetToken(ctx context.Context, accountID string) (*AuthToken, error) {
	tm.mu.RLock()
	cached, exists := tm.cache[accountID]
	tm.mu.RUnlock()

	if exists && !cached.token.IsExpired() {
		tm.checkExpirationWarning(cached.token, accountID)
		return cached.token, nil
	}

	// Try loading from persistent store
	stored, err := tm.store.Get(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("get token for %s: %w", accountID, err)
	}
	if stored == nil {
		return nil, fmt.Errorf("get token for %s: no token found", accountID)
	}

	// Auto-refresh if expired
	if stored.IsExpired() {
		log.Printf("[TokenManager] Token expired for account %s, attempting refresh", accountID)
		newToken, err := tm.client.RefreshToken(ctx, stored)
		if err != nil {
			return nil, fmt.Errorf("auto-refresh token for %s: %w", accountID, err)
		}

		if err := tm.store.Save(ctx, accountID, newToken); err != nil {
			return nil, fmt.Errorf("save refreshed token for %s: %w", accountID, err)
		}

		tm.updateCache(accountID, newToken)
		return newToken, nil
	}

	tm.updateCache(accountID, stored)
	tm.checkExpirationWarning(stored, accountID)
	return stored, nil
}

// SaveToken saves a token for the given account to both cache and persistent store.
func (tm *TokenManager) SaveToken(ctx context.Context, accountID string, token *AuthToken) error {
	if err := tm.store.Save(ctx, accountID, token); err != nil {
		return fmt.Errorf("save token for %s: %w", accountID, err)
	}
	tm.updateCache(accountID, token)
	return nil
}

// InvalidateToken removes a token from cache and store (e.g., on revocation).
func (tm *TokenManager) InvalidateToken(ctx context.Context, accountID string) error {
	tm.mu.Lock()
	delete(tm.cache, accountID)
	tm.mu.Unlock()

	if err := tm.store.Delete(ctx, accountID); err != nil {
		return fmt.Errorf("invalidate token for %s: %w", accountID, err)
	}
	return nil
}

// RefreshAll attempts to refresh all stored tokens.
// Returns a list of accounts that failed to refresh.
func (tm *TokenManager) RefreshAll(ctx context.Context) []string {
	accountIDs, err := tm.store.List(ctx)
	if err != nil {
		log.Printf("[TokenManager] Failed to list accounts for refresh: %v", err)
		return nil
	}

	var failed []string
	for _, id := range accountIDs {
		if _, err := tm.GetToken(ctx, id); err != nil {
			log.Printf("[TokenManager] Auto-refresh failed for account %s: %v", id, err)
			failed = append(failed, id)
		}
	}
	return failed
}

func (tm *TokenManager) updateCache(accountID string, token *AuthToken) {
	tm.mu.Lock()
	tm.cache[accountID] = &cachedToken{
		token:     token,
		expiresAt: token.ExpiresAt,
	}
	tm.mu.Unlock()
}

func (tm *TokenManager) checkExpirationWarning(token *AuthToken, accountID string) {
	if token.ExpiresAt.IsZero() {
		return
	}
	timeUntilExpiry := time.Until(token.ExpiresAt)
	if timeUntilExpiry < tm.alertBefore {
		log.Printf("[TokenManager] WARNING: Token for account %s will expire in %s",
			accountID, timeUntilExpiry.Round(time.Minute))
	}
}
