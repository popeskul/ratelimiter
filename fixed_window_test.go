package ratelimiter_test

import (
	"context"
	"github.com/popeskul/ratelimiter"
	"sync"
	"testing"
	"time"
)

func TestFixedWindow(t *testing.T) {
	t.Run("Basic Allow", func(t *testing.T) {
		limiter := ratelimiter.NewFixedWindow(&ratelimiter.Config{
			Rate:   10,
			Window: time.Second,
		})

		for i := 0; i < 10; i++ {
			if !limiter.Allow() {
				t.Errorf("Request %d should be allowed", i+1)
			}
		}

		if limiter.Allow() {
			t.Error("Request should be denied")
		}
	})

	t.Run("AllowN", func(t *testing.T) {
		limiter := ratelimiter.NewFixedWindow(&ratelimiter.Config{
			Rate:   10,
			Window: time.Second,
		})

		if !limiter.AllowN(5) {
			t.Error("AllowN(5) should be allowed")
		}

		if !limiter.AllowN(5) {
			t.Error("AllowN(5) should be allowed")
		}

		if limiter.AllowN(1) {
			t.Error("AllowN(1) should be denied")
		}
	})

	t.Run("Wait", func(t *testing.T) {
		limiter := ratelimiter.NewFixedWindow(&ratelimiter.Config{
			Rate:   2,
			Window: 100 * time.Millisecond,
		})

		if err := limiter.Wait(context.Background()); err != nil {
			t.Errorf("First wait should not error: %v", err)
		}

		if err := limiter.Wait(context.Background()); err != nil {
			t.Errorf("Second wait should not error: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		if err := limiter.Wait(ctx); err == nil {
			t.Error("Third wait should timeout")
		}
	})

	t.Run("Reset", func(t *testing.T) {
		limiter := ratelimiter.NewFixedWindow(&ratelimiter.Config{
			Rate:   2,
			Window: time.Second,
		})

		if !limiter.Allow() {
			t.Error("First request should be allowed")
		}
		if !limiter.Allow() {
			t.Error("Second request should be allowed")
		}
		if limiter.Allow() {
			t.Error("Third request should be denied")
		}

		limiter.Reset()

		if !limiter.Allow() {
			t.Error("After reset, first request should be allowed")
		}
		if !limiter.Allow() {
			t.Error("After reset, second request should be allowed")
		}
	})

	t.Run("Concurrent", func(t *testing.T) {
		limiter := ratelimiter.NewFixedWindow(&ratelimiter.Config{
			Rate:   100,
			Window: time.Second,
		})

		var wg sync.WaitGroup
		concurrent := 1000

		for i := 0; i < concurrent; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if limiter.Allow() {
					time.Sleep(time.Millisecond)
				}
			}()
		}

		wg.Wait()

		metrics := limiter.GetMetrics()
		if metrics.AllowedRequests != 100 {
			t.Errorf("Expected 100 allowed requests, got %d", metrics.AllowedRequests)
		}
		if metrics.DeniedRequests != 900 {
			t.Errorf("Expected 900 denied requests, got %d", metrics.DeniedRequests)
		}
	})

	t.Run("GetMetrics", func(t *testing.T) {
		limiter := ratelimiter.NewFixedWindow(&ratelimiter.Config{
			Rate:   10,
			Window: time.Second,
		})

		for i := 0; i < 15; i++ {
			limiter.Allow()
		}

		metrics := limiter.GetMetrics()

		if metrics.TotalRequests != 15 {
			t.Errorf("Expected 15 total requests, got %d", metrics.TotalRequests)
		}
		if metrics.AllowedRequests != 10 {
			t.Errorf("Expected 10 allowed requests, got %d", metrics.AllowedRequests)
		}
		if metrics.DeniedRequests != 5 {
			t.Errorf("Expected 5 denied requests, got %d", metrics.DeniedRequests)
		}
		if metrics.CurrentRate != 10 {
			t.Errorf("Expected current rate 10, got %d", metrics.CurrentRate)
		}
		if metrics.WindowDuration != time.Second {
			t.Errorf("Expected window duration 1s, got %v", metrics.WindowDuration)
		}
	})
}
