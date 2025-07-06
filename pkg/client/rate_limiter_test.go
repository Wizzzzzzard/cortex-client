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

func TestManagerReturnsLimitExceeded(t *testing.T) {
	conf := &Config{
		Limit:            3,
		Throttle:         10 * time.Millisecond,
		TokenResetsAfter: 0, // No reset for this test
	}
	m := NewManager(conf)
	if m == nil {
		t.Fatal("expected a new Manager instance, got nil")
	}

	m.activeTokens["token1"] = NewToken()
	m.activeTokens["token2"] = NewToken()

	if m.isLimitExceeded() {
		t.Fatal("expected limit to not be exceeded, but it is")
	}

	m.activeTokens["token3"] = NewToken()

	if !m.isLimitExceeded() {
		t.Fatal("expected limit to be exceeded, but it is not")
	}
}

func TestManagerIncrementsNeedToken(t *testing.T) {
	conf := &Config{
		Limit:            1,
		Throttle:         10 * time.Millisecond,
		TokenResetsAfter: 0, // No reset for this test
	}

	m := NewManager(conf)
	if m == nil {
		t.Fatal("expected a new Manager instance, got nil")
	}

	m.incNeedToken()

	if m.needToken != 1 {
		t.Fatalf("expected needToken to be 1, got %d", m.needToken)
	}
}

func TestManagerDecrementsNeedToken(t *testing.T) {
	conf := &Config{
		Limit:            1,
		Throttle:         10 * time.Millisecond,
		TokenResetsAfter: 0, // No reset for this test
	}

	m := NewManager(conf)
	if m == nil {
		t.Fatal("expected a new Manager instance, got nil")
	}

	m.needToken = 5 // Set needToken to 1 for this test
	m.decNeedToken()

	if m.needToken != 4 {
		t.Fatalf("expected needToken to be 4, got %d", m.needToken)
	}
}

func TestManagerAwaitsToken(t *testing.T) {
	conf := &Config{
		Limit:            1,
		Throttle:         10 * time.Millisecond,
		TokenResetsAfter: 0, // No reset for this test
	}

	m := NewManager(conf)
	if m == nil {
		t.Fatal("expected a new Manager instance, got nil")
	}

	m.needToken = 1 // Set needToken to 1 for this test

	if !m.awaitingToken() {
		t.Fatal("expected awaitingToken to return true, but it returned false")
	}
}

func TestManagerMakesToken(t *testing.T) {
	conf := &Config{
		Limit:            1,
		Throttle:         10 * time.Millisecond,
		TokenResetsAfter: 0, // No reset for this test
	}

	m := NewManager(conf)
	if m == nil {
		t.Fatal("expected a new Manager instance, got nil")
	}

	token := m.makeToken()

	if token == nil {
		t.Fatal("expected a valid token, got nil")
	}
	if token.ID == "" {
		t.Fatal("expected a non-empty token ID")
	}
	if token.CreatedAt.IsZero() {
		t.Fatal("expected a valid creation time, got zero value")
	}

}

func TestManagerGeneratesToken(t *testing.T) {
	conf := &Config{
		Limit:            2,
		Throttle:         10 * time.Millisecond,
		TokenResetsAfter: 0, // No reset for this test
	}

	m := NewManager(conf)
	if m == nil {
		t.Fatal("expected a new Manager instance, got nil")
	}

	m.tryGenerateToken()

	if m.isLimitExceeded() {
		t.Fatal("expected limit not to be exceeded, but it is")
	}

	token := <-m.outChan

	if token == nil {
		t.Fatal("expected a valid token, got nil")
	}
	if token.ID == "" {
		t.Fatal("expected a non-empty token ID")
	}
	if token.CreatedAt.IsZero() {
		t.Fatal("expected a valid creation time, got zero value")
	}
	if len(m.activeTokens) != 1 {
		t.Fatalf("expected 1 active token, got %d", len(m.activeTokens))
	}
	if _, exists := m.activeTokens[token.ID]; !exists {
		t.Fatalf("expected token with ID %s to be in active tokens, but it is not", token.ID)
	}

}

func TestManagerDoesntGenerateTokenWhenLimitExceeded(t *testing.T) {
	conf := &Config{
		Limit:            2,
		Throttle:         10 * time.Millisecond,
		TokenResetsAfter: 0, // No reset for this test
	}

	m := NewManager(conf)
	if m == nil {
		t.Fatal("expected a new Manager instance, got nil")
	}

	m.tryGenerateToken()

	if m.isLimitExceeded() {
		t.Fatal("expected limit not to be exceeded, but it is")
	}

	m.tryGenerateToken()

	if !m.isLimitExceeded() {
		t.Fatal("expected limit to be exceeded after adding a token, but it is not")
	}
}

func TestManagerReleasesToken(t *testing.T) {
	conf := &Config{
		Limit:            2,
		Throttle:         10 * time.Millisecond,
		TokenResetsAfter: 0, // No reset for this test
	}

	m := NewManager(conf)
	if m == nil {
		t.Fatal("expected a new Manager instance, got nil")
	}

	m.tryGenerateToken()
	token := <-m.outChan

	if token == nil {
		t.Fatal("expected a valid token, got nil")
	}
	if m.isLimitExceeded() {
		t.Fatal("expected limit not to be exceeded, but it is")
	}
	if _, exists := m.activeTokens[token.ID]; !exists {
		t.Fatalf("expected token with ID %s to be in active tokens, but it is not", token.ID)
	}
	if len(m.activeTokens) != 1 {
		t.Fatalf("expected 1 active token, got %d", len(m.activeTokens))
	}

	m.releaseToken(token)

	if len(m.activeTokens) != 0 {
		t.Fatalf("expected no active tokens after release, got %d", len(m.activeTokens))
	}
	if _, exists := m.activeTokens[token.ID]; exists {
		t.Fatalf("expected token with ID %s to be removed from active tokens, but it still exists", token.ID)
	}
}

func TestGenerateNewMaxConcurrencyRateLimiter(t *testing.T) {
	conf := &Config{
		Limit:            3,
		Throttle:         10 * time.Millisecond,
		TokenResetsAfter: 0, // No reset for this test
	}

	rl, err := NewMaxConcurrencyRateLimiter(conf)
	if rl == nil {
		t.Fatal("expected a new Manager instance, got nil")
	}
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMaxConcurrencyRateLimiterAcquiresTokens(t *testing.T) {
	conf := &Config{
		Limit:            2,
		Throttle:         10 * time.Millisecond,
		TokenResetsAfter: 0, // No reset for this test
	}

	rl, err := NewMaxConcurrencyRateLimiter(conf)
	if rl == nil {
		t.Fatal("expected a new Manager instance, got nil")
	}
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

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

func TestMaxConcurrencyRateLimiterAcquiresTokensAfterRelease(t *testing.T) {
	conf := &Config{
		Limit:            2,
		Throttle:         10 * time.Millisecond,
		TokenResetsAfter: 0, // No reset for this test
	}

	rl, err := NewMaxConcurrencyRateLimiter(conf)
	if rl == nil {
		t.Fatal("expected a new Manager instance, got nil")
	}
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

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
