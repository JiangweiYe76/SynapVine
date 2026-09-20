package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
)

// EncryptResponse marshals data to JSON and encrypts it with AES-GCM. The key
// must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256. The returned
// ciphertext has the randomly generated nonce prepended.
func EncryptResponse(data interface{}, key []byte) ([]byte, error) {
	plain, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Seal appends the ciphertext to dst, so passing the nonce as dst emits
	// nonce||ciphertext without a second allocation or copy.
	ciphertext := gcm.Seal(nonce, nonce, plain, nil)
	return ciphertext, nil
}
