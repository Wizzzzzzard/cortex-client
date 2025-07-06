package ratelimiter

import (
	"log"
	"sync/atomic"
	"time"
)

// MaxUint holds the maximum unsigned int value
const MaxUint = ^uint(0)

// MaxInt holds the maximum int value
const MaxInt = int(MaxUint >> 1)

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

	// If limit is not defined, then default to max value
	if m.limit <= 0 {
		m.limit = MaxInt
	}

	// If the config TokenResetsAfter value exists, then run the reset task
	if conf.TokenResetsAfter > 0 {
		m.runResetTokenTask(conf.TokenResetsAfter)
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

	if !token.IsExpired() {
		log.Printf("unable to relase token %s - has not expired", token)
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

func (m *Manager) releaseExpiredTokens() {
	for _, token := range m.activeTokens {
		if token.IsExpired() {
			go func(t *Token) {
				m.releaseChan <- t
			}(token)
		}
	}
}

func (m *Manager) runResetTokenTask(resetAfter time.Duration) {
	go func() {
		ticker := time.NewTicker(resetAfter)
		for range ticker.C {
			for _, token := range m.activeTokens {
				if token.NeedReset(resetAfter) {
					go func(t *Token) {
						m.releaseChan <- t
					}(token)
				}
			}
		}
	}()
}
