package ratelimiter

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type NestedWindow struct {
	outerRate        int64
	innerRate        int64
	outerWindow      time.Duration
	innerWindow      time.Duration
	outerCount       int64
	innerCount       int64
	outerWindowStart int64
	innerWindowStart int64
	totalRequests    int64
	allowedCount     int64
	deniedCount      int64
	mu               sync.Mutex
}

func NewNestedWindow(config *Config) *NestedWindow {
	now := time.Now().UnixNano()
	return &NestedWindow{
		outerRate:        int64(config.Rate),
		innerRate:        int64(config.Burst),
		outerWindow:      config.Window,
		innerWindow:      config.Window / 10, // Inner window is 1/10th of the outer window
		outerCount:       0,
		innerCount:       0,
		outerWindowStart: now,
		innerWindowStart: now,
	}
}

func (nw *NestedWindow) Allow() bool {
	return nw.AllowN(1)
}

func (nw *NestedWindow) AllowN(n int) bool {
	nw.mu.Lock()
	defer nw.mu.Unlock()

	atomic.AddInt64(&nw.totalRequests, 1)

	now := time.Now().UnixNano()
	nw.updateWindows(now)

	if nw.outerCount+int64(n) > nw.outerRate || nw.innerCount+int64(n) > nw.innerRate {
		atomic.AddInt64(&nw.deniedCount, 1)
		return false
	}

	nw.outerCount += int64(n)
	nw.innerCount += int64(n)
	atomic.AddInt64(&nw.allowedCount, 1)
	return true
}

func (nw *NestedWindow) Wait(ctx context.Context) error {
	nw.mu.Lock()
	defer nw.mu.Unlock()

	atomic.AddInt64(&nw.totalRequests, 1)

	timer := time.NewTimer(nw.innerWindow)
	defer timer.Stop()

	for {
		if nw.allow() {
			atomic.AddInt64(&nw.allowedCount, 1)
			return nil
		}

		atomic.AddInt64(&nw.deniedCount, 1)

		select {
		case <-timer.C:
			timer.Reset(nw.innerWindow)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (nw *NestedWindow) allow() bool {
	now := time.Now().UnixNano()
	nw.updateWindows(now)

	if nw.outerCount+1 > nw.outerRate || nw.innerCount+1 > nw.innerRate {
		return false
	}

	nw.outerCount++
	nw.innerCount++
	return true
}

func (nw *NestedWindow) WaitN(ctx context.Context, n int) error {
	timer := time.NewTimer(nw.innerWindow)
	defer timer.Stop()

	for {
		if nw.AllowN(n) {
			return nil
		}

		select {
		case <-timer.C:
			timer.Reset(nw.innerWindow)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (nw *NestedWindow) Reset() {
	nw.mu.Lock()
	defer nw.mu.Unlock()

	now := time.Now().UnixNano()
	nw.outerCount = 0
	nw.innerCount = 0
	nw.outerWindowStart = now
	nw.innerWindowStart = now
	atomic.StoreInt64(&nw.totalRequests, 0)
	atomic.StoreInt64(&nw.allowedCount, 0)
	atomic.StoreInt64(&nw.deniedCount, 0)
}

func (nw *NestedWindow) GetMetrics() Metrics {
	return Metrics{
		TotalRequests:   atomic.LoadInt64(&nw.totalRequests),
		AllowedRequests: atomic.LoadInt64(&nw.allowedCount),
		DeniedRequests:  atomic.LoadInt64(&nw.deniedCount),
		CurrentRate:     atomic.LoadInt64(&nw.outerRate),
		LastResetTime:   atomic.LoadInt64(&nw.outerWindowStart),
		TotalWaitTime:   0, // NestedWindow doesn't track wait time
		MaxWaitTime:     0, // NestedWindow doesn't track max wait time
		WindowDuration:  nw.outerWindow,
		InnerRate:       atomic.LoadInt64(&nw.innerRate),
		InnerWindow:     nw.innerWindow,
	}
}

func (nw *NestedWindow) updateWindows(now int64) {
	if now-nw.outerWindowStart >= nw.outerWindow.Nanoseconds() {
		nw.outerCount = 0
		nw.outerWindowStart = now
	}

	if now-nw.innerWindowStart >= nw.innerWindow.Nanoseconds() {
		nw.innerCount = 0
		nw.innerWindowStart = now
	}
}
