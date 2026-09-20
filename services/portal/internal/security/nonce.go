package security

import (
	"log/slog"
	"sync"
	"time"
)

// NonceStore remembers recently seen nonces so that a replayed request can be
// rejected. Entries are retained for 5 minutes, which comfortably exceeds the
// 30-second signature window.
type NonceStore struct {
	mu     sync.Mutex
	nonces map[string]time.Time // nonce -> expiry time
}

// NewNonceStore creates an empty NonceStore.
func NewNonceStore() *NonceStore {
	return &NonceStore{
		nonces: make(map[string]time.Time),
	}
}

// Mark records nonce as used and reports whether it was previously unseen.
// A false result means the nonce was already present, i.e. a replay.
func (ns *NonceStore) Mark(nonce string) bool {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, exists := ns.nonces[nonce]; exists {
		slog.Warn("nonce_replay_rejected", slog.String("nonce", nonce))
		return false
	}

	ns.nonces[nonce] = time.Now().Add(5 * time.Minute)
	return true
}

// CleanExpired drops every nonce whose expiry has passed.
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
