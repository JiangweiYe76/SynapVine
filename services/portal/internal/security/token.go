package security

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// TokenStore manages temporary access tokens for API authentication
type TokenStore struct {
	mu     sync.RWMutex              // Mutex for thread-safe access
	tokens map[string]time.Time     // Map of token to expiration time
}

// NewTokenStore creates a new TokenStore instance
func NewTokenStore() *TokenStore {
	return &TokenStore{
		tokens: make(map[string]time.Time),
	}
}

// Issue generates and stores a new temporary access token.
// Returns the generated token string (valid for 5 minutes).
func (ts *TokenStore) Issue() (string, error) {
	// Generate a cryptographically secure random token (32 bytes = 64 hex chars)
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	// Store token with 5-minute expiration
	ts.mu.Lock()
	ts.tokens[token] = time.Now().Add(5 * time.Minute)
	ts.mu.Unlock()

	return token, nil
}

// Validate checks if a token is valid (exists and not expired)
// Returns true if the token is valid, false otherwise
func (ts *TokenStore) Validate(token string) bool {
	ts.mu.RLock()
	expire, ok := ts.tokens[token]
	ts.mu.RUnlock()

	// Check if token exists
	if !ok {
		return false
	}

	// Check if token is expired
	if time.Now().After(expire) {
		// Clean up expired token
		ts.mu.Lock()
		delete(ts.tokens, token)
		ts.mu.Unlock()
		return false
	}
	return true
}

// CleanExpired removes all expired tokens from the store
// Returns the number of tokens that were cleaned.
func (ts *TokenStore) CleanExpired() int {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	now := time.Now()
	count := 0
	for token, expire := range ts.tokens {
		if now.After(expire) {
			delete(ts.tokens, token)
			count++
		}
	}
	return count
}
