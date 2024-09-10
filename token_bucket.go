package ratelimiter

import (
	"context"
	"sync/atomic"
	"time"
)

type TokenBucket struct {
	rate           float64
	capacity       float64
	tokens         int64
	lastRefillTime int64
	refillInterval time.Duration
	allowedCount   int64
	deniedCount    int64
}

func NewTokenBucket(config *Config) *TokenBucket {
	refillInterval := time.Second
	if config.Rate > 0 {
		refillInterval = time.Duration(float64(time.Second) / float64(config.Rate))
	}
	return &TokenBucket{
		rate:           float64(config.Rate),
		capacity:       float64(config.Capacity),
		tokens:         int64(config.Capacity),
		lastRefillTime: time.Now().UnixNano(),
		refillInterval: refillInterval,
	}
}

func (tb *TokenBucket) Allow() bool {
	return tb.AllowN(1)
}

func (tb *TokenBucket) AllowN(n int) bool {
	now := time.Now().UnixNano()
	tb.refill(now)

	available := atomic.LoadInt64(&tb.tokens)
	if available >= int64(n) {
		if atomic.AddInt64(&tb.tokens, -int64(n)) >= 0 {
			atomic.AddInt64(&tb.allowedCount, 1)
			return true
		}
	}
	atomic.AddInt64(&tb.deniedCount, 1)
	return false
}

func (tb *TokenBucket) Wait(ctx context.Context) error {
	return tb.WaitN(ctx, 1)
}

func (tb *TokenBucket) WaitN(ctx context.Context, n int) error {
	if tb.AllowN(n) {
		return nil
	}

	timer := time.NewTimer(tb.timeToToken(float64(n)))
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			if tb.AllowN(n) {
				return nil
			}
			timer.Reset(tb.timeToToken(float64(n)))
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (tb *TokenBucket) Reset() {
	atomic.StoreInt64(&tb.tokens, int64(tb.capacity))
	atomic.StoreInt64(&tb.lastRefillTime, time.Now().UnixNano())
	atomic.StoreInt64(&tb.allowedCount, 0)
	atomic.StoreInt64(&tb.deniedCount, 0)
}

func (tb *TokenBucket) GetMetrics() Metrics {
	return Metrics{
		TotalRequests:   atomic.LoadInt64(&tb.allowedCount) + atomic.LoadInt64(&tb.deniedCount),
		AllowedRequests: atomic.LoadInt64(&tb.allowedCount),
		DeniedRequests:  atomic.LoadInt64(&tb.deniedCount),
		CurrentRate:     int64(tb.rate),
		LastResetTime:   atomic.LoadInt64(&tb.lastRefillTime),
		TotalWaitTime:   0, // TokenBucket doesn't track total wait time
		MaxWaitTime:     0, // TokenBucket doesn't track max wait time
		WindowDuration:  tb.refillInterval,
		InnerRate:       0, // Not applicable for TokenBucket
		InnerWindow:     0, // Not applicable for TokenBucket
	}
}

func (tb *TokenBucket) refill(now int64) {
	last := atomic.LoadInt64(&tb.lastRefillTime)
	elapsed := time.Duration(now - last)
	tokensToAdd := int64(tb.rate * elapsed.Seconds())
	if tokensToAdd > 0 {
		newTokens := atomic.AddInt64(&tb.tokens, tokensToAdd)
		if newTokens > int64(tb.capacity) {
			atomic.StoreInt64(&tb.tokens, int64(tb.capacity))
		}
		atomic.StoreInt64(&tb.lastRefillTime, now)
	}
}

func (tb *TokenBucket) timeToToken(tokens float64) time.Duration {
	available := float64(atomic.LoadInt64(&tb.tokens))
	if available >= tokens {
		return 0
	}
	missingTokens := tokens - available
	return time.Duration(missingTokens / tb.rate * float64(time.Second))
}
