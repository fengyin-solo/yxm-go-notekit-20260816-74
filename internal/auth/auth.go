package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"
)

// Authenticator validates bearer tokens and manages a simple token registry.
type Authenticator struct {
	tokens map[string]tokenInfo
}

type tokenInfo struct {
	created  time.Time
	expires  time.Time
	username string
}

// New creates a new Authenticator.
func New() *Authenticator {
	return &Authenticator{tokens: make(map[string]tokenInfo)}
}

// Validate checks the provided token string. Returns the associated username on
// success, or ErrUnauthorized on failure.
func (a *Authenticator) Validate(token string) (string, error) {
	a.cleanExpired()
	info, ok := a.tokens[token]
	if !ok {
		return "", errors.New("invalid token")
	}
	if !info.expires.IsZero() && time.Now().After(info.expires) {
		delete(a.tokens, token)
		return "", errors.New("token expired")
	}
	return info.username, nil
}

// Register issues a new bearer token for the given username, valid for the
// specified duration. Returns the raw token (shown only once).
func (a *Authenticator) Register(username string, ttl time.Duration) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw)
	expires := time.Time{}
	if ttl > 0 {
		expires = time.Now().Add(ttl)
	}
	a.tokens[token] = tokenInfo{created: time.Now(), expires: expires, username: username}
	return token, nil
}

// Revoke deletes a token from the registry.
func (a *Authenticator) Revoke(token string) {
	delete(a.tokens, token)
}

// ParseAuthorizationHeader extracts the bearer token from a request's Authorization
// header.
func ParseAuthorizationHeader(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(h, "Bearer ")
}

// RequireAuth is middleware that rejects requests without a valid token.
func RequireAuth(a *Authenticator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ParseAuthorizationHeader(r)
		if token == "" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		_, err := a.Validate(token)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ValidateToken performs a constant-time comparison against the configured token.
func ValidateToken(token, expected string) bool {
	return subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1
}

// cleanExpired removes all expired tokens from the registry.
func (a *Authenticator) cleanExpired() {
	now := time.Now()
	for token, info := range a.tokens {
		if !info.expires.IsZero() && now.After(info.expires) {
			delete(a.tokens, token)
		}
	}
}
