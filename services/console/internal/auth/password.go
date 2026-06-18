package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters chosen for an interactive login flow on a small
// dev box: ~0.5s per hash on a modern CPU. Tune upward in production
// once the workload is profiled.
const (
	argonTime    = 1
	argonMemory  = 64 * 1024 // KiB -> 64 MiB
	argonThreads = 4
	argonKeyLen  = 32
	argonSaltLen = 16
)

// HashPassword generates an argon2id hash of the given password. The
// returned string is self-describing: it embeds the algorithm
// parameters and salt so future parameter changes don't break
// previously stored hashes.
//
// Format: $argon2id$v=19$m=<mem>,t=<time>,p=<threads>$<saltB64>$<hashB64>
func HashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("read salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return encoded, nil
}

// CheckPassword verifies a plain-text password against a stored hash.
// Returns true on match. Uses constant-time comparison to avoid
// leaking match progress via timing.
func CheckPassword(password, stored string) bool {
	salt, hash, params, err := decodeArgonHash(stored)
	if err != nil {
		return false
	}

	computed := argon2.IDKey([]byte(password), salt, params.time, params.memory, params.threads, uint32(len(hash)))
	return subtle.ConstantTimeCompare(hash, computed) == 1
}

type argonParams struct {
	memory  uint32
	time    uint32
	threads uint8
}

func decodeArgonHash(s string) (salt, hash []byte, p argonParams, err error) {
	parts := strings.Split(s, "$")
	// parts[0] is empty (leading "$"), parts[1] is the algorithm.
	if len(parts) != 6 {
		return nil, nil, p, errors.New("invalid hash format")
	}
	if parts[1] != "argon2id" {
		return nil, nil, p, errors.New("unsupported algorithm")
	}
	if _, perr := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.memory, &p.time, &p.threads); perr != nil {
		return nil, nil, p, fmt.Errorf("parse params: %w", perr)
	}

	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, p, fmt.Errorf("decode salt: %w", err)
	}
	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, p, fmt.Errorf("decode hash: %w", err)
	}
	return salt, hash, p, nil
}
