package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

// envelopePrefix tags encrypted at-rest values. The "v1" segment is the
// scheme version; future key-rotation or algorithm changes can introduce
// "v2" while Decrypt keeps accepting older envelopes.
const envelopePrefix = "enc:v1:"

// ErrInvalidKey means the supplied key is not 32 bytes (AES-256).
var ErrInvalidKey = errors.New("encryption key must be 32 bytes for AES-256")

// ErrMalformed means a stored value carries the encryption envelope prefix
// but cannot be parsed (bad base64, truncated, or tampered ciphertext).
var ErrMalformed = errors.New("malformed encrypted value")

// KeyCipher encrypts and decrypts provider API keys at rest using
// AES-256-GCM. GCM provides both confidentiality and integrity: a
// tampered or truncated ciphertext fails authentication instead of
// returning garbage.
type KeyCipher struct {
	gcm cipher.AEAD
}

// NewKeyCipher builds a KeyCipher from a raw 32-byte AES-256 key.
func NewKeyCipher(key []byte) (*KeyCipher, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("%w: got %d bytes", ErrInvalidKey, len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}
	return &KeyCipher{gcm: gcm}, nil
}

// NewKeyCipherFromBase64 builds a KeyCipher from a standard-base64
// encoded 32-byte key, as supplied via the PROVIDER_ENCRYPTION_KEY
// environment variable.
func NewKeyCipherFromBase64(encoded string) (*KeyCipher, error) {
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return nil, fmt.Errorf("decode base64 encryption key: %w", err)
	}
	return NewKeyCipher(key)
}

// Encrypt seals plaintext and returns the versioned envelope string.
// A fresh random nonce is used for every call, so encrypting the same
// plaintext twice yields different ciphertexts. An empty plaintext
// maps to an empty string so empty/nullable columns stay empty.
func (c *KeyCipher) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	// Seal appends ciphertext+tag to nonce, storing them contiguously.
	sealed := c.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return envelopePrefix + base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt opens a value produced by Encrypt. Values without the envelope
// prefix are treated as legacy plaintext (rows written before encryption
// was enabled) and returned unchanged; they get encrypted lazily on their
// next write. An empty input returns an empty string.
func (c *KeyCipher) Decrypt(stored string) (string, error) {
	if stored == "" {
		return "", nil
	}
	if !strings.HasPrefix(stored, envelopePrefix) {
		return stored, nil
	}

	raw, err := base64.StdEncoding.DecodeString(stored[len(envelopePrefix):])
	if err != nil {
		return "", fmt.Errorf("%w: decode ciphertext: %v", ErrMalformed, err)
	}
	nonceSize := c.gcm.NonceSize()
	if len(raw) < nonceSize {
		return "", fmt.Errorf("%w: ciphertext shorter than nonce", ErrMalformed)
	}

	nonce, ciphertext := raw[:nonceSize], raw[nonceSize:]
	plaintext, err := c.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt (wrong key or tampered data): %w", err)
	}
	return string(plaintext), nil
}
