package client

import (
	"time"

	"github.com/segmentio/ksuid"
)

// token factory function creates a new token
type tokenFactory func() *Token

// Token represents a Rate Limit Token
type Token struct {
	// The unique token ID
	ID string

	// The time at which the token was created
	CreatedAt time.Time
}

// NewToken creates a new token
func NewToken() *Token {
	return &Token{
		ID:        ksuid.New().String(),
		CreatedAt: time.Now().UTC(),
	}
}

type RateLimiter interface {
	Acquire() (*Token, error)
}

type Config struct {
	// Throttle is the min time between requests for a Throttle Rate Limiter
	Throttle time.Duration
}

type Manager struct {
	errorChan chan error
	outChan   chan *Token
	inChan    chan struct{}
	makeToken tokenFactory
}

func NewManager(conf *Config) *Manager {
	m := &Manager{
		errorChan: make(chan error),
		outChan:   make(chan *Token),
		inChan:    make(chan struct{}),
		makeToken: NewToken,
	}
	return m
}

func (m *Manager) Acquire() (*Token, error) {
	go func() {
		m.inChan <- struct{}{}
	}()

	// Await rate limit token
	select {
	case t := <-m.outChan:
		return t, nil
	case err := <-m.errorChan:
		return nil, err
	}
}

func (m *Manager) tryGenerateToken() {
	// panic if token factory is not defined
	if m.makeToken == nil {
		panic("ErrTokenFactoryNotDefined")
	}

	token := m.makeToken()

	// send token to outChan
	go func() {
		m.outChan <- token
	}()
}
