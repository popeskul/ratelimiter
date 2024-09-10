package ratelimiter

import (
	"context"
	"sync"
	"time"
)

type SlidingWindow struct {
	rate     int
	window   time.Duration
	requests []time.Time
	mu       sync.Mutex
}

func NewSlidingWindow(config *Config) *SlidingWindow {
	return &SlidingWindow{
		rate:     config.Rate,
		window:   config.Window,
		requests: make([]time.Time, 0, config.Rate),
	}
}

func (sw *SlidingWindow) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	sw.clearExpired(now)

	if len(sw.requests) < sw.rate {
		sw.requests = append(sw.requests, now)
		return true
	}
	return false
}

func (sw *SlidingWindow) AllowN(n int) bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	sw.clearExpired(now)

	if len(sw.requests)+n <= sw.rate {
		for i := 0; i < n; i++ {
			sw.requests = append(sw.requests, now)
		}
		return true
	}
	return false
}

func (sw *SlidingWindow) Wait(ctx context.Context) error {
	for {
		sw.mu.Lock()
		now := time.Now()
		sw.clearExpired(now)

		if len(sw.requests) < sw.rate {
			sw.requests = append(sw.requests, now)
			sw.mu.Unlock()
			return nil
		}

		nextExpiry := sw.requests[0].Add(sw.window)
		sw.mu.Unlock()

		select {
		case <-time.After(nextExpiry.Sub(now)):
			// Continue and try again
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (sw *SlidingWindow) WaitN(ctx context.Context, n int) error {
	for {
		sw.mu.Lock()
		now := time.Now()
		sw.clearExpired(now)

		if len(sw.requests)+n <= sw.rate {
			for i := 0; i < n; i++ {
				sw.requests = append(sw.requests, now)
			}
			sw.mu.Unlock()
			return nil
		}

		nextExpiry := sw.requests[0].Add(sw.window)
		sw.mu.Unlock()

		select {
		case <-time.After(nextExpiry.Sub(now)):
			// Continue and try again
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (sw *SlidingWindow) clearExpired(now time.Time) {
	cutoff := now.Add(-sw.window)
	i := 0
	for ; i < len(sw.requests) && sw.requests[i].Before(cutoff); i++ {
	}
	if i > 0 {
		sw.requests = sw.requests[i:]
	}
}

func (sw *SlidingWindow) Reset() {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.requests = sw.requests[:0]
}

func (sw *SlidingWindow) GetMetrics() Metrics {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	sw.clearExpired(now)

	return Metrics{
		TotalRequests:   int64(len(sw.requests)),
		AllowedRequests: int64(len(sw.requests)),
		DeniedRequests:  0,
		CurrentRate:     int64(sw.rate),
		LastResetTime:   0,
		TotalWaitTime:   0,
		MaxWaitTime:     0,
		WindowDuration:  sw.window,
		InnerRate:       0,
		InnerWindow:     0,
	}
}
