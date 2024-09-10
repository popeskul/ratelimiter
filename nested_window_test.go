package ratelimiter_test

import (
	"context"
	"github.com/popeskul/ratelimiter"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNestedWindow(t *testing.T) {
	t.Run("Basic Allow", func(t *testing.T) {
		limiter, _ := ratelimiter.New(
			ratelimiter.WithAlgorithm("nested_window"),
			ratelimiter.WithRate(10),
			ratelimiter.WithBurst(5),
			ratelimiter.WithWindow(time.Second),
		)

		for i := 0; i < 5; i++ {
			if !limiter.Allow() {
				t.Errorf("Request %d should be allowed", i+1)
			}
		}

		if limiter.Allow() {
			t.Error("Request should be denied")
		}
	})

	t.Run("AllowN", func(t *testing.T) {
		limiter, _ := ratelimiter.New(
			ratelimiter.WithAlgorithm("nested_window"),
			ratelimiter.WithRate(10),
			ratelimiter.WithBurst(5),
			ratelimiter.WithWindow(time.Second),
		)

		if !limiter.AllowN(3) {
			t.Error("AllowN(3) should be allowed")
		}

		if !limiter.AllowN(2) {
			t.Error("AllowN(2) should be allowed")
		}

		if limiter.AllowN(1) {
			t.Error("AllowN(1) should be denied")
		}
	})

	t.Run("Concurrent", func(t *testing.T) {
		limiter, _ := ratelimiter.New(
			ratelimiter.WithAlgorithm("nested_window"),
			ratelimiter.WithRate(10),
			ratelimiter.WithBurst(5),
			ratelimiter.WithWindow(time.Second),
		)

		var wg sync.WaitGroup
		var allowed, denied int64
		totalRequests := int64(100)

		for i := int64(0); i < totalRequests; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
				defer cancel()
				err := limiter.Wait(ctx)
				if err == nil {
					atomic.AddInt64(&allowed, 1)
				} else {
					atomic.AddInt64(&denied, 1)
				}
			}()
		}

		wg.Wait()

		metrics := limiter.GetMetrics()
		t.Logf("Metrics after concurrent test: %+v", metrics)
		t.Logf("Allowed: %d, Denied: %d, Total: %d", allowed, denied, allowed+denied)

		if allowed+denied != totalRequests {
			t.Errorf("Expected total of %d requests, got %d", totalRequests, allowed+denied)
		}

		tolerance := int64(5) // Allow for small variations due to timing

		if metrics.AllowedRequests < allowed-tolerance || metrics.AllowedRequests > allowed+tolerance {
			t.Errorf("Mismatch in allowed requests: test counted %d, metrics show %d", allowed, metrics.AllowedRequests)
		}

		if metrics.DeniedRequests < denied-tolerance || metrics.DeniedRequests > denied+tolerance {
			t.Errorf("Mismatch in denied requests: test counted %d, metrics show %d", denied, metrics.DeniedRequests)
		}

		// Check if the number of allowed requests is close to the rate limit
		expectedAllowed := int64(10) // 10 requests per second for 200ms
		if allowed < expectedAllowed-tolerance || allowed > expectedAllowed+tolerance {
			t.Errorf("Expected about %d allowed requests, got %d", expectedAllowed, allowed)
		}
	})

	t.Run("Reset", func(t *testing.T) {
		limiter, _ := ratelimiter.New(
			ratelimiter.WithAlgorithm("nested_window"),
			ratelimiter.WithRate(10),
			ratelimiter.WithBurst(5),
			ratelimiter.WithWindow(time.Second),
		)

		for i := 0; i < 5; i++ {
			limiter.Allow()
		}

		limiter.Reset()

		for i := 0; i < 5; i++ {
			if !limiter.Allow() {
				t.Errorf("After reset, request %d should be allowed", i+1)
			}
		}
	})

	t.Run("GetMetrics", func(t *testing.T) {
		limiter, _ := ratelimiter.New(
			ratelimiter.WithAlgorithm("nested_window"),
			ratelimiter.WithRate(10),
			ratelimiter.WithBurst(5),
			ratelimiter.WithWindow(time.Second),
		)

		for i := 0; i < 7; i++ {
			allowed := limiter.Allow()
			if i < 5 && !allowed {
				t.Errorf("Request %d should be allowed", i+1)
			} else if i >= 5 && allowed {
				t.Errorf("Request %d should be denied", i+1)
			}
		}

		metrics := limiter.GetMetrics()

		if metrics.TotalRequests != 7 {
			t.Errorf("Expected 7 total requests, got %d", metrics.TotalRequests)
		}

		if metrics.AllowedRequests != 5 {
			t.Errorf("Expected 5 allowed requests, got %d", metrics.AllowedRequests)
		}

		if metrics.DeniedRequests != 2 {
			t.Errorf("Expected 2 denied requests, got %d", metrics.DeniedRequests)
		}

		if metrics.CurrentRate != 10 {
			t.Errorf("Expected current rate 10, got %d", metrics.CurrentRate)
		}

		if metrics.WindowDuration != time.Second {
			t.Errorf("Expected window duration 1s, got %v", metrics.WindowDuration)
		}

		if metrics.InnerRate != 5 {
			t.Errorf("Expected inner rate 5, got %d", metrics.InnerRate)
		}

		if metrics.InnerWindow != 100*time.Millisecond {
			t.Errorf("Expected inner window 100ms, got %v", metrics.InnerWindow)
		}
	})
}
