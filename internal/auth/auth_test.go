package auth

import (
	"testing"
	"time"
)

func TestAuthenticatorValidate(t *testing.T) {
	a := New()
	token, err := a.Register("alice", 1*time.Hour)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	user, err := a.Validate(token)
	if err != nil {
		t.Errorf("Validate: %v", err)
	}
	if user != "alice" {
		t.Errorf("user = %q, want alice", user)
	}
}

func TestAuthenticatorInvalid(t *testing.T) {
	a := New()
	if _, err := a.Validate("bad-token"); err == nil {
		t.Error("invalid token should error")
	}
}

func TestAuthenticatorRevoke(t *testing.T) {
	a := New()
	token, _ := a.Register("alice", 0)
	a.Revoke(token)
	if _, err := a.Validate(token); err == nil {
		t.Error("revoked token should error")
	}
}

func TestValidateToken(t *testing.T) {
	if !ValidateToken("secret", "secret") {
		t.Error("same tokens should match")
	}
	if ValidateToken("a", "b") {
		t.Error("different tokens should not match")
	}
}
