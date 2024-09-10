package ratelimiter

import (
	"context"
	"sync/atomic"
	"time"
)

type FixedWindow struct {
	rate         int64
	window       time.Duration
	count        int64
	windowStart  int64
	allowedCount int64
	deniedCount  int64
}

func NewFixedWindow(config *Config) *FixedWindow {
	return &FixedWindow{
		rate:        int64(config.Rate),
		window:      config.Window,
		count:       0,
		windowStart: time.Now().UnixNano(),
	}
}

func (fw *FixedWindow) Allow() bool {
	return fw.AllowN(1)
}

func (fw *FixedWindow) AllowN(n int) bool {
	now := time.Now().UnixNano()
	windowStart := atomic.LoadInt64(&fw.windowStart)

	if now-windowStart >= fw.window.Nanoseconds() {
		atomic.StoreInt64(&fw.count, 0)
		atomic.StoreInt64(&fw.windowStart, now)
		// Удалено неэффективное присваивание windowStart = now
	}

	count := atomic.AddInt64(&fw.count, int64(n))
	if count <= fw.rate {
		atomic.AddInt64(&fw.allowedCount, 1)
		return true
	}
	atomic.AddInt64(&fw.deniedCount, 1)
	atomic.AddInt64(&fw.count, -int64(n)) // Rollback the count increase
	return false
}

func (fw *FixedWindow) Wait(ctx context.Context) error {
	return fw.WaitN(ctx, 1)
}

func (fw *FixedWindow) WaitN(ctx context.Context, n int) error {
	for {
		if fw.AllowN(n) {
			return nil
		}

		select {
		case <-time.After(fw.timeToNextWindow()):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (fw *FixedWindow) Reset() {
	atomic.StoreInt64(&fw.count, 0)
	atomic.StoreInt64(&fw.windowStart, time.Now().UnixNano())
	atomic.StoreInt64(&fw.allowedCount, 0)
	atomic.StoreInt64(&fw.deniedCount, 0)
}

func (fw *FixedWindow) GetMetrics() Metrics {
	return Metrics{
		TotalRequests:   atomic.LoadInt64(&fw.allowedCount) + atomic.LoadInt64(&fw.deniedCount),
		AllowedRequests: atomic.LoadInt64(&fw.allowedCount),
		DeniedRequests:  atomic.LoadInt64(&fw.deniedCount),
		CurrentRate:     atomic.LoadInt64(&fw.rate),
		LastResetTime:   atomic.LoadInt64(&fw.windowStart),
		TotalWaitTime:   0, // FixedWindow doesn't track wait time
		MaxWaitTime:     0, // FixedWindow doesn't track max wait time
		WindowDuration:  fw.window,
		InnerRate:       0, // Not applicable for FixedWindow
		InnerWindow:     0, // Not applicable for FixedWindow
	}
}

func (fw *FixedWindow) timeToNextWindow() time.Duration {
	now := time.Now().UnixNano()
	windowStart := atomic.LoadInt64(&fw.windowStart)
	elapsed := time.Duration(now - windowStart)
	if elapsed >= fw.window {
		return 0
	}
	return fw.window - elapsed
}
