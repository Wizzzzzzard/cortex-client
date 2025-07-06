package client

import (
	"testing"
	"time"
)

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
}

func TestNewManagerWithConfig(t *testing.T) {
	conf := &Config{
		Limit:            5,
		Throttle:         10 * time.Millisecond,
		TokenResetsAfter: 0, // No reset for this test
	}

	m := NewManager(conf)
	if m == nil {
		t.Fatal("expected a new Manager instance, got nil")
	}
	if m.limit != 5 {
		t.Fatalf("expected limit %d, got %d", 5, m.limit)
	}
	if m.errorChan == nil {
		t.Fatal("expected error channel to be initialized, got nil")
	}
	if m.outChan == nil {
		t.Fatal("expected out channel to be initialized, got nil")
	}
	if m.inChan == nil {
		t.Fatal("expected in channel to be initialized, got nil")
	}
	if m.releaseChan == nil {
		t.Fatal("expected release channel to be initialized, got nil")
	}
	if m.activeTokens == nil {
		t.Fatal("expected active tokens map to be initialized, got nil")
	}
	if m.makeToken == nil {
		t.Fatal("expected token factory to be initialized, got nil")
	}
	if m.needToken != 0 {
		t.Fatalf("expected needToken to be 0, got %d", m.needToken)
	}
	if len(m.activeTokens) != 0 {
		t.Fatalf("expected active tokens to be empty, got %d tokens", len(m.activeTokens))
	}
}

func TestManagerAcquiresToken(t *testing.T) {
	conf := &Config{
		Limit:            1,
		Throttle:         10 * time.Millisecond,
		TokenResetsAfter: 0, // No reset for this test
	}

	m := NewManager(conf)
	if m == nil {
		t.Fatal("expected a new Manager instance, got nil")
	}

	token, err := m.Acquire()
	if err != nil {
		t.Fatalf("expected no error when acquiring token, got %v", err)
	}
	if token == nil {
		t.Fatal("expected a valid token, got nil")
	}

}
