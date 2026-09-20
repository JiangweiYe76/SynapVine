package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"strconv"
	"time"
)

// VerifySignature reports whether an HMAC-SHA256 request signature is valid.
// A request is rejected when its timestamp is malformed or more than 30
// seconds old, when its nonce was already used (replay), or when the MAC does
// not match.
func VerifySignature(path, token, timestamp, nonce, signature string, nonceStore *NonceStore) bool {
	// The 30-second window bounds how long a captured request stays replayable.
	if ts, err := strconv.ParseInt(timestamp, 10, 64); err != nil || time.Now().Unix()-ts > 30 {
		slog.Warn("signature_timestamp_invalid",
			slog.String("path", path),
			slog.String("timestamp", timestamp),
			slog.String("reason", "expired_or_malformed"),
		)
		return false
	}

	// Mark consumes the nonce, so a second sighting of the same nonce is a
	// replay rather than a legitimate retry.
	if !nonceStore.Mark(nonce) {
		slog.Warn("signature_replay_detected",
			slog.String("path", path),
			slog.String("nonce", nonce),
		)
		return false
	}

	// The client must build the identical path:timestamp:nonce string.
	payload := path + ":" + timestamp + ":" + nonce

	mac := hmac.New(sha256.New, []byte(token))
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))

	// hmac.Equal compares in constant time, so a near-miss signature does not
	// leak the length of the matching prefix through timing.
	if !hmac.Equal([]byte(signature), []byte(expected)) {
		slog.Warn("signature_mismatch",
			slog.String("path", path),
		)
		return false
	}

	return true
}
