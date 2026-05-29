package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
)

// EncryptResponse encrypts data using AES-GCM encryption
// Parameters:
//   - data: The data to encrypt (will be JSON marshaled)
//   - key: The AES key (must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256)
//
// Returns the encrypted ciphertext with the nonce prepended, or an error if encryption fails
func EncryptResponse(data interface{}, key []byte) ([]byte, error) {
	// Convert data to JSON
	plain, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Encrypt and prepend nonce to ciphertext
	ciphertext := gcm.Seal(nonce, nonce, plain, nil)
	return ciphertext, nil
}
