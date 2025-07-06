package client

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/segmentio/ksuid"
)

var ErrTokenFactoryNotDefined = fmt.Errorf("token factory not defined")
var ErrInvalidLimit = fmt.Errorf("invalid limit, must be greater than zero")

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
	Release(*Token)
}

type Config struct {
	// Limit determines how many rate limit tokens can be active at a time
	Limit int
	// Throttle is the min time between requests for a Throttle Rate Limiter
	Throttle time.Duration
	// TokenResetsAfter is the maximum amount of time a token can live before being
	// forcefully released - if set to zero time then the token may live forever
	TokenResetsAfter time.Duration
}

type Manager struct {
	errorChan    chan error
	releaseChan  chan *Token
	outChan      chan *Token
	inChan       chan struct{}
	needToken    int64
	activeTokens map[string]*Token
	limit        int
	makeToken    tokenFactory
}

func NewManager(conf *Config) *Manager {
	m := &Manager{
		errorChan:    make(chan error),
		outChan:      make(chan *Token),
		inChan:       make(chan struct{}),
		activeTokens: make(map[string]*Token),
		releaseChan:  make(chan *Token),
		needToken:    0,
		limit:        conf.Limit,
		makeToken:    NewToken,
	}
	return m
}

func (m *Manager) Acquire() (*Token, error) {
	go func() {
		m.inChan <- struct{}{}
	}()

	// Await rate limit token
	select {
	case token := <-m.outChan:
		return token, nil
	case err := <-m.errorChan:
		return nil, err
	}
}

func (m *Manager) Release(token *Token) {
	// send token to releaseChan
	go func() {
		m.releaseChan <- token
	}()
}

func (m *Manager) isLimitExceeded() bool {
	return len(m.activeTokens) >= m.limit
}

func (m *Manager) incNeedToken() {
	atomic.AddInt64(&m.needToken, 1)
}

func (m *Manager) decNeedToken() {
	atomic.AddInt64(&m.needToken, -1)
}

func (m *Manager) awaitingToken() bool {
	return atomic.LoadInt64(&m.needToken) > 0
}

func (m *Manager) tryGenerateToken() {
	// panic if token factory is not defined
	if m.makeToken == nil {
		panic(ErrTokenFactoryNotDefined)
	}

	// cannot continue if limit has been reached
	if m.isLimitExceeded() {
		m.incNeedToken()
		return
	}

	token := m.makeToken()

	// Add token to active map
	m.activeTokens[token.ID] = token

	// send token to outChan
	go func() {
		m.outChan <- token
	}()
}

func (m *Manager) releaseToken(token *Token) {
	if token == nil {
		log.Print("unable to relase nil token")
		return
	}

	if _, ok := m.activeTokens[token.ID]; !ok {
		log.Printf("unable to relase token %s - not in use", token)
		return
	}

	// Delete from map
	delete(m.activeTokens, token.ID)

	// process anything waiting for a rate limit
	if m.awaitingToken() {
		m.decNeedToken()
		go m.tryGenerateToken()
	}
}

// NewMaxConcurrencyRateLimiter returns a max concurrency rate limiter
func NewMaxConcurrencyRateLimiter(conf *Config) (RateLimiter, error) {
	if conf.Limit <= 0 {
		return nil, ErrInvalidLimit
	}

	m := NewManager(conf)
	// max concurrency await function
	await := func() {
		go func() {
			for {
				select {
				case <-m.inChan:
					m.tryGenerateToken()
				case token := <-m.releaseChan:
					m.releaseToken(token)
				}
			}
		}()
	}

	await()
	return m, nil
}
