package client

import (
	"testing"
	"time"
)

func TestRateLimiterAcquire(t *testing.T) {
	conf := &Config{
		Throttle: 100 * time.Millisecond,
	}
	manager := NewManager(conf)

	t.Run("Acquire token", func(t *testing.T) {
		token, err := manager.Acquire()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if token == nil {
			t.Fatal("expected a token, got nil")
		}
		if token.ID == "" {
			t.Fatal("expected a non-empty token ID")
		}
	})
}
