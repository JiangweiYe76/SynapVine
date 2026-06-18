package auth

import "testing"

func TestHashAndCheckPassword_RoundTrip(t *testing.T) {
	const password = "admin123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if hash == password {
		t.Fatal("hash must not equal plaintext")
	}

	if !CheckPassword(password, hash) {
		t.Error("CheckPassword returned false for the correct password")
	}
	if CheckPassword("wrong-password", hash) {
		t.Error("CheckPassword returned true for an incorrect password")
	}
}

func TestHashPassword_UniqueSalts(t *testing.T) {
	const password = "same-input"

	first, err := HashPassword(password)
	if err != nil {
		t.Fatalf("first HashPassword failed: %v", err)
	}
	second, err := HashPassword(password)
	if err != nil {
		t.Fatalf("second HashPassword failed: %v", err)
	}

	if first == second {
		t.Error("two hashes of the same password must differ (random salt)")
	}
}

func TestCheckPassword_InvalidHash(t *testing.T) {
	if CheckPassword("anything", "not-a-valid-hash") {
		t.Error("CheckPassword must return false for a malformed hash")
	}
	if CheckPassword("anything", "$argon2id$v=19$m=65536,t=1,p=4$bad$bad") {
		t.Error("CheckPassword must return false when base64 decoding fails")
	}
}
