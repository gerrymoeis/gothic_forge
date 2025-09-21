package security

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("secret")
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}
	ok, err := VerifyPassword("secret", hash)
	if err != nil {
		t.Fatalf("verify error: %v", err)
	}
	if !ok {
		t.Fatalf("expected password to verify")
	}
	ok, err = VerifyPassword("wrong", hash)
	if err != nil {
		t.Fatalf("verify error: %v", err)
	}
	if ok {
		t.Fatalf("expected wrong password to fail verify")
	}
}

func TestHashPasswordEmpty(t *testing.T) {
	if _, err := HashPassword(""); err == nil {
		t.Fatalf("expected error for empty password")
	}
}
