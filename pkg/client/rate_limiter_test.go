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

	t.Run("Acquire multiple tokens", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			token, err := manager.Acquire()
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if token == nil || token.ID == "" {
				t.Fatal("expected a valid token, got nil or empty ID")
			}
			time.Sleep(50 * time.Millisecond) // Ensure we respect the throttle
		}
	})
}

func TestRateLimiterTokenCreation(t *testing.T) {
	conf := &Config{
		Throttle: 100 * time.Millisecond,
	}
	manager := NewManager(conf)

	t.Run("Token creation", func(t *testing.T) {
		token := manager.makeToken()
		if token == nil {
			t.Fatal("expected a token, got nil")
		}
		if token.ID == "" {
			t.Fatal("expected a non-empty token ID")
		}
		if token.CreatedAt.IsZero() {
			t.Fatal("expected a valid creation time, got zero value")
		}
	})
}

func TestRateLimiterErrorHandling(t *testing.T) {
	conf := &Config{
		Throttle: 100 * time.Millisecond,
	}
	manager := NewManager(conf)

	t.Run("Error channel", func(t *testing.T) {
		// Simulate an error by closing the inChan without sending a token
		close(manager.inChan)
		_, err := manager.Acquire()
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
	})
}

// TestRateLimiterConcurrency tests the rate limiter's ability to handle concurrent requests.
func TestRateLimiterConcurrency(t *testing.T) {
	conf := &Config{
		Throttle: 100 * time.Millisecond,
	}
	manager := NewManager(conf)

	t.Run("Concurrent Acquire", func(t *testing.T) {
		numGoroutines := 10
		done := make(chan struct{}, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer func() { done <- struct{}{} }()
				token, err := manager.Acquire()
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if token == nil || token.ID == "" {
					t.Error("expected a valid token, got nil or empty ID")
				}
			}()
		}

		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})
}
