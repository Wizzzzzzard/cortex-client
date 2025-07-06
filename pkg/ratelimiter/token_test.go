package ratelimiter

import "testing"

func TestNewTokenGeneratesNewToken(t *testing.T) {
	token := NewToken()
	if token == nil {
		t.Fatal("expected a new token, got nil")
	}
	if token.ID == "" {
		t.Fatal("expected a non-empty token ID")
	}
	if token.CreatedAt.IsZero() {
		t.Fatal("expected a valid creation time, got zero value")
	}
	if !token.ExpiresAt.IsZero() {
		t.Fatal("expected a default expiration time of 0")
	}
}
