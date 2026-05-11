package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

// VerifySignature verifies an HMAC signature for request integrity
// Parameters:
//   - path: The request path
//   - token: The secret token
//   - timestamp: Unix timestamp of the request
//   - nonce: Unique nonce for this request
//   - signature: The HMAC signature to verify
//   - nonceStore: Store to track used nonces (prevents replay attacks)
//
// Returns true if the signature is valid and the request is not expired/replayed
func VerifySignature(path, token, timestamp, nonce, signature string, nonceStore *NonceStore) bool {
	// Parse and validate timestamp (must be within 30 seconds)
	if ts, err := strconv.ParseInt(timestamp, 10, 64); err != nil || time.Now().Unix()-ts > 30 {
		return false
	}

	// Check nonce to prevent replay attacks
	if !nonceStore.Mark(nonce) {
		return false
	}

	// Build the payload to sign: path:timestamp:nonce
	payload := path + ":" + timestamp + ":" + nonce

	// Calculate expected HMAC-SHA256 signature
	mac := hmac.New(sha256.New, []byte(token))
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures using constant-time comparison
	return hmac.Equal([]byte(signature), []byte(expected))
}
