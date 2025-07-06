package ratelimiter

import (
	"time"
)

type RateLimiter interface {
	Acquire() (*Token, error)
	Release(*Token)
}

// Config represents a rate limiter config object
type Config struct {
	// Limit determines how many rate limit tokens can be active at a time
	Limit int

	// FixedInterval sets the fixed time window for a Fixed Window Rate Limiter
	FixedInterval time.Duration

	// Throttle is the min time between requests for a Throttle Rate Limiter
	Throttle time.Duration

	// TokenResetsAfter is the maximum amount of time a token can live before being
	// forcefully released - if set to zero time then the token may live forever
	TokenResetsAfter time.Duration
}

// FixedWindowInterval represents a fixed window of time with a start / end time
type FixedWindowInterval struct {
	startTime time.Time
	endTime   time.Time
	interval  time.Duration
}

func (w *FixedWindowInterval) setWindowTime() {
	w.startTime = time.Now().UTC()
	w.endTime = time.Now().UTC().Add(w.interval)
}

func (w *FixedWindowInterval) run(cb func()) {
	go func() {
		ticker := time.NewTicker(w.interval)
		w.setWindowTime()
		for range ticker.C {
			cb()
			w.setWindowTime()
		}
	}()
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

// NewThrottleRateLimiter returns a throttle rate limiter
func NewThrottleRateLimiter(conf *Config) (RateLimiter, error) {
	if conf.Throttle == 0 {
		return nil, ErrInvalidThrottleDuration
	}

	m := NewManager(conf)

	// Throttle Await Function
	await := func(throttle time.Duration) {
		ticker := time.NewTicker(throttle)
		go func() {
			<-m.inChan
			m.tryGenerateToken()
			for {
				select {
				case <-m.inChan:
					<-ticker.C
					m.tryGenerateToken()
				case t := <-m.releaseChan:
					m.releaseToken(t)
				}
			}
		}()
	}

	// Call await to start
	await(conf.Throttle)
	return m, nil
}

// NewFixedWindowRateLimiter returns a fixed window rate limiter
func NewFixedWindowRateLimiter(conf *Config) (RateLimiter, error) {
	if conf.FixedInterval == 0 {
		return nil, ErrInvalidInterval
	}

	if conf.Limit == 0 {
		return nil, ErrInvalidLimit
	}

	m := NewManager(conf)
	w := &FixedWindowInterval{interval: conf.FixedInterval}

	// override the manager makeToken function
	m.makeToken = func() *Token {
		t := NewToken()
		t.ExpiresAt = w.endTime
		return t
	}

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

	w.run(m.releaseExpiredTokens)
	await()
	return m, nil
}
