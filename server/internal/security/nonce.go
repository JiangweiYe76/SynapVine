package security

import (
	"log/slog"
	"sync"
	"time"
)

// NonceStore manages nonces to prevent replay attacks
type NonceStore struct {
	mu     sync.Mutex           // Mutex for thread-safe access
	nonces map[string]time.Time // Map of nonce to expiration time
}

// NewNonceStore creates a new NonceStore instance
func NewNonceStore() *NonceStore {
	return &NonceStore{
		nonces: make(map[string]time.Time),
	}
}

// Mark attempts to mark a nonce as used
// Returns true if the nonce was successfully marked (first use)
// Returns false if the nonce was already used (replay attack detected)
func (ns *NonceStore) Mark(nonce string) bool {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	// Check if nonce already exists
	if _, exists := ns.nonces[nonce]; exists {
		slog.Warn("nonce_replay_rejected", slog.String("nonce", nonce))
		return false
	}

	// Mark nonce with 5-minute expiration
	ns.nonces[nonce] = time.Now().Add(5 * time.Minute)
	return true
}

// CleanExpired removes all expired nonces from the store
func (ns *NonceStore) CleanExpired() {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	now := time.Now()
	for nonce, expire := range ns.nonces {
		if now.After(expire) {
			delete(ns.nonces, nonce)
		}
	}
}
