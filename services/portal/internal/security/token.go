package security

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// TokenStore holds the short-lived access tokens handed to portal clients.
// Each token is valid for 5 minutes from issue.
type TokenStore struct {
	mu     sync.RWMutex
	tokens map[string]time.Time // token -> expiry time
}

// NewTokenStore creates an empty TokenStore.
func NewTokenStore() *TokenStore {
	return &TokenStore{
		tokens: make(map[string]time.Time),
	}
}

// Issue generates a new access token, stores it, and returns it. The caller
// has 5 minutes to present the token before it expires.
func (ts *TokenStore) Issue() (string, error) {
	// crypto/rand, so the token cannot be predicted from earlier ones.
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	ts.mu.Lock()
	ts.tokens[token] = time.Now().Add(5 * time.Minute)
	ts.mu.Unlock()

	return token, nil
}

// Validate reports whether token exists and has not expired. An expired token
// is evicted as a side effect.
func (ts *TokenStore) Validate(token string) bool {
	ts.mu.RLock()
	expire, ok := ts.tokens[token]
	ts.mu.RUnlock()

	if !ok {
		return false
	}

	if time.Now().After(expire) {
		ts.mu.Lock()
		delete(ts.tokens, token)
		ts.mu.Unlock()
		return false
	}
	return true
}

// CleanExpired drops every expired token and returns how many were removed.
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
