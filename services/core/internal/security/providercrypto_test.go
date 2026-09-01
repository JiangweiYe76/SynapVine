package security

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"
)

func testKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return key
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	c, err := NewKeyCipher(testKey(t))
	if err != nil {
		t.Fatalf("NewKeyCipher: %v", err)
	}

	plaintext := "sk-secret-abc123-XYZ"
	sealed, err := c.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if sealed == plaintext {
		t.Fatal("ciphertext equals plaintext")
	}
	if !strings.HasPrefix(sealed, envelopePrefix) {
		t.Fatalf("ciphertext missing envelope prefix: %q", sealed)
	}

	got, err := c.Decrypt(sealed)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if got != plaintext {
		t.Fatalf("roundtrip mismatch: got %q want %q", got, plaintext)
	}
}

func TestEncryptUsesRandomNonce(t *testing.T) {
	c, err := NewKeyCipher(testKey(t))
	if err != nil {
		t.Fatalf("NewKeyCipher: %v", err)
	}

	a, err := c.Encrypt("same-plaintext")
	if err != nil {
		t.Fatalf("Encrypt a: %v", err)
	}
	b, err := c.Encrypt("same-plaintext")
	if err != nil {
		t.Fatalf("Encrypt b: %v", err)
	}
	if a == b {
		t.Fatal("expected different ciphertexts for the same plaintext (nonce reuse)")
	}
}

func TestDecryptWrongKeyFails(t *testing.T) {
	enc, err := NewKeyCipher(testKey(t))
	if err != nil {
		t.Fatalf("NewKeyCipher enc: %v", err)
	}
	dec, err := NewKeyCipher(testKey(t))
	if err != nil {
		t.Fatalf("NewKeyCipher dec: %v", err)
	}

	sealed, err := enc.Encrypt("sk-secret")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if _, err := dec.Decrypt(sealed); err == nil {
		t.Fatal("Decrypt with wrong key succeeded, want authentication failure")
	}
}

func TestDecryptTamperedCiphertextFails(t *testing.T) {
	c, err := NewKeyCipher(testKey(t))
	if err != nil {
		t.Fatalf("NewKeyCipher: %v", err)
	}

	sealed, err := c.Encrypt("sk-secret")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	// Flip the last character of the base64 payload (inside the GCM tag).
	tampered := sealed[:len(sealed)-1]
	if sealed[len(sealed)-1] == 'A' {
		tampered += "B"
	} else {
		tampered += "A"
	}
	if _, err := c.Decrypt(tampered); err == nil {
		t.Fatal("Decrypt of tampered ciphertext succeeded, want authentication failure")
	}
}

func TestDecryptLegacyPlaintextPassthrough(t *testing.T) {
	c, err := NewKeyCipher(testKey(t))
	if err != nil {
		t.Fatalf("NewKeyCipher: %v", err)
	}

	legacy := "sk-plaintext-from-old-row"
	got, err := c.Decrypt(legacy)
	if err != nil {
		t.Fatalf("Decrypt legacy: %v", err)
	}
	if got != legacy {
		t.Fatalf("legacy passthrough mismatch: got %q want %q", got, legacy)
	}
}

func TestEmptyStringRoundtrip(t *testing.T) {
	c, err := NewKeyCipher(testKey(t))
	if err != nil {
		t.Fatalf("NewKeyCipher: %v", err)
	}

	sealed, err := c.Encrypt("")
	if err != nil {
		t.Fatalf("Encrypt empty: %v", err)
	}
	if sealed != "" {
		t.Fatalf("Encrypt empty = %q, want empty string", sealed)
	}
	got, err := c.Decrypt("")
	if err != nil {
		t.Fatalf("Decrypt empty: %v", err)
	}
	if got != "" {
		t.Fatalf("Decrypt empty = %q, want empty string", got)
	}
}

func TestDecryptMalformedEnvelopeFails(t *testing.T) {
	c, err := NewKeyCipher(testKey(t))
	if err != nil {
		t.Fatalf("NewKeyCipher: %v", err)
	}

	cases := map[string]string{
		"not base64":      envelopePrefix + "!!!not-base64!!!",
		"too short":       envelopePrefix + base64.StdEncoding.EncodeToString([]byte("short")),
		"empty after tag": envelopePrefix,
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := c.Decrypt(input); err == nil {
				t.Fatal("Decrypt malformed input succeeded, want error")
			}
		})
	}
}

func TestNewKeyCipherRejectsBadKeyLength(t *testing.T) {
	for _, n := range []int{0, 16, 24, 31, 33, 64} {
		if _, err := NewKeyCipher(make([]byte, n)); err == nil {
			t.Fatalf("NewKeyCipher with %d-byte key succeeded, want error", n)
		}
	}
}

func TestNewKeyCipherFromBase64(t *testing.T) {
	key := testKey(t)
	encoded := base64.StdEncoding.EncodeToString(key)

	c, err := NewKeyCipherFromBase64(encoded)
	if err != nil {
		t.Fatalf("NewKeyCipherFromBase64: %v", err)
	}

	sealed, err := c.Encrypt("roundtrip-via-base64-key")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	got, err := c.Decrypt(sealed)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if got != "roundtrip-via-base64-key" {
		t.Fatalf("mismatch: got %q", got)
	}

	if _, err := NewKeyCipherFromBase64("!!!not-base64!!!"); err == nil {
		t.Fatal("NewKeyCipherFromBase64 with invalid base64 succeeded, want error")
	}
	if _, err := NewKeyCipherFromBase64(base64.StdEncoding.EncodeToString(make([]byte, 16))); err == nil {
		t.Fatal("NewKeyCipherFromBase64 with 16-byte key succeeded, want error")
	}
}
