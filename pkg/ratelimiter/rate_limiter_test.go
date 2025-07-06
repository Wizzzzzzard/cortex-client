package ratelimiter

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func DoWork(r RateLimiter, workerCount int) {
	var wg sync.WaitGroup
	rand.Seed(time.Now().UnixNano())

	doWork := func(id int) {
		// Acquire a rate limit token
		token, err := r.Acquire()
		fmt.Printf("Rate Limit Token %s acquired at %s...\n", token.ID, time.Now().UTC())
		if err != nil {
			panic(err)
		}
		// Simulate some work
		n := rand.Intn(5)
		fmt.Printf("Worker %d Sleeping %d seconds...\n", id, n)
		time.Sleep(time.Duration(n) * time.Second)
		fmt.Printf("Worker %d Done\n", id)
		r.Release(token)
		wg.Done()
	}

	// Spin up a 10 workers that need a rate limit resource
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go doWork(i)
	}

	wg.Wait()
}

func TestGenerateNewMaxConcurrencyRateLimiter(t *testing.T) {
	conf := &Config{
		Limit:            3,
		Throttle:         10 * time.Millisecond,
		TokenResetsAfter: 0, // No reset for this test
	}

	rl, err := NewMaxConcurrencyRateLimiter(conf)
	if rl == nil {
		t.Fatal("expected a new MaxConcurrencyRateLimiter instance, got nil")
	}
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestGenerateNewMaxConcurrencyRateLimiterFailsWithLimitZero(t *testing.T) {
	conf := &Config{
		Limit:            0,
		Throttle:         10 * time.Millisecond,
		TokenResetsAfter: 0, // No reset for this test
	}

	rl, err := NewMaxConcurrencyRateLimiter(conf)
	if rl != nil {
		t.Fatal("expected new MaxConcurrencyRateLimiter instance to fail")
	}
	if err != ErrInvalidLimit {
		t.Fatalf("expected %v, got another error %v", ErrInvalidLimit, err)
	}
}

func TestGenerateNewThrottleRateLimiter(t *testing.T) {
	conf := &Config{
		Limit:            3,
		Throttle:         10 * time.Millisecond,
		TokenResetsAfter: 0, // No reset for this test
	}

	rl, err := NewThrottleRateLimiter(conf)
	if rl == nil {
		t.Fatal("expected a new ThrottleRateLimiter instance, got nil")
	}
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestGenerateNewThrottleRateLimiterFailsWithThrottleZero(t *testing.T) {
	conf := &Config{
		Limit:            3,
		Throttle:         0 * time.Millisecond,
		TokenResetsAfter: 0, // No reset for this test
	}

	rl, err := NewThrottleRateLimiter(conf)
	if rl != nil {
		t.Fatal("expected new NewThrottleRateLimiter instance to fail")
	}
	if err != ErrInvalidThrottleDuration {
		t.Fatalf("expected %v, got another error %v", ErrInvalidThrottleDuration, err)
	}
}

func TestGenerateNewFixedWindowRateLimiter(t *testing.T) {
	conf := &Config{
		Limit:            3,
		Throttle:         10 * time.Millisecond,
		FixedInterval:    15 * time.Second,
		TokenResetsAfter: 0, // No reset for this test
	}

	rl, err := NewThrottleRateLimiter(conf)
	if rl == nil {
		t.Fatal("expected a new NewFixedWindowRateLimiter instance, got nil")
	}
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestGenerateNewFixedWindowRateLimiterFailsWithLimitZero(t *testing.T) {
	conf := &Config{
		Limit:            0,
		Throttle:         10 * time.Millisecond,
		FixedInterval:    15 * time.Second,
		TokenResetsAfter: 0, // No reset for this test
	}

	rl, err := NewFixedWindowRateLimiter(conf)
	if rl != nil {
		t.Fatal("expected new NewFixedWindowRateLimiter instance to fail")
	}
	if err != ErrInvalidLimit {
		t.Fatalf("expected %v, got another error %v", ErrInvalidLimit, err)
	}
}

func TestGenerateNewFixedWindowRateLimiterFailsWithFixedIntervalZero(t *testing.T) {
	conf := &Config{
		Limit:            2,
		Throttle:         10 * time.Millisecond,
		FixedInterval:    0 * time.Second,
		TokenResetsAfter: 0, // No reset for this test
	}

	rl, err := NewFixedWindowRateLimiter(conf)
	if rl != nil {
		t.Fatal("expected new NewFixedWindowRateLimiter instance to fail")
	}
	if err != ErrInvalidInterval {
		t.Fatalf("expected %v, got another error %v", ErrInvalidInterval, err)
	}
}

func TestRateLimitersAcquiresTokens(t *testing.T) {
	conf := &Config{
		Limit:            2,
		Throttle:         10 * time.Millisecond,
		FixedInterval:    15 * time.Second,
		TokenResetsAfter: 0, // No reset for this test
	}

	rl1, err := NewMaxConcurrencyRateLimiter(conf)
	if rl1 == nil {
		t.Fatal("expected a new Manager instance, got nil")
	}
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	rl2, err := NewThrottleRateLimiter(conf)
	if rl2 == nil {
		t.Fatal("expected a new Manager instance, got nil")
	}
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	rl3, err := NewFixedWindowRateLimiter(conf)
	if rl3 == nil {
		t.Fatal("expected a new Manager instance, got nil")
	}
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	rateLimiters := []RateLimiter{rl1, rl2, rl3}

	for _, rl := range rateLimiters {
		token1, err := rl.Acquire()
		if err != nil {
			t.Fatalf("expected to acquire token, got error: %v", err)
		}
		if token1 == nil {
			t.Fatal("expected a valid token, got nil")
		}

		token2, err := rl.Acquire()
		if err != nil {
			t.Fatalf("expected to acquire second token, got error: %v", err)
		}
		if token2 == nil {
			t.Fatal("expected a valid second token, got nil")
		}
	}
}

func TestRateLimitersAcquireTokensAfterRelease(t *testing.T) {
	conf := &Config{
		Limit:            2,
		Throttle:         10 * time.Millisecond,
		FixedInterval:    15 * time.Second,
		TokenResetsAfter: 0, // No reset for this test
	}

	rl1, err := NewMaxConcurrencyRateLimiter(conf)
	if rl1 == nil {
		t.Fatal("expected a new Manager instance, got nil")
	}
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	rl2, err := NewThrottleRateLimiter(conf)
	if rl2 == nil {
		t.Fatal("expected a new Manager instance, got nil")
	}
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	rl3, err := NewFixedWindowRateLimiter(conf)
	if rl3 == nil {
		t.Fatal("expected a new Manager instance, got nil")
	}
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	rateLimiters := []RateLimiter{rl1, rl2, rl3}

	for _, rl := range rateLimiters {
		token1, err := rl.Acquire()
		if err != nil {
			t.Fatalf("expected to acquire token, got error: %v", err)
		}
		if token1 == nil {
			t.Fatal("expected a valid token, got nil")
		}

		token2, err := rl.Acquire()
		if err != nil {
			t.Fatalf("expected to acquire second token, got error: %v", err)
		}
		if token2 == nil {
			t.Fatal("expected a valid second token, got nil")
		}

		rl.Release(token2)

		token3, err := rl.Acquire()
		if err != nil {
			t.Fatalf("expected to acquire third token, got error: %v", err)
		}
		if token3 == nil {
			t.Fatal("expected a valid third token, got nil")
		}
	}
}
