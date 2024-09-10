package ratelimiter_test

import (
	"github.com/popeskul/ratelimiter"
	"sync"
	"testing"
	"time"
)

func TestSlidingWindow(t *testing.T) {
	t.Run("Basic Allow", func(t *testing.T) {
		limiter := ratelimiter.NewSlidingWindow(&ratelimiter.Config{
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
		limiter := ratelimiter.NewSlidingWindow(&ratelimiter.Config{
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

	t.Run("Reset", func(t *testing.T) {
		limiter := ratelimiter.NewSlidingWindow(&ratelimiter.Config{
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
		limiter := ratelimiter.NewSlidingWindow(&ratelimiter.Config{
			Rate:   100,
			Window: time.Second,
		})

		var wg sync.WaitGroup
		concurrent := 1000
		var allowed int64

		for i := 0; i < concurrent; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if limiter.Allow() {
					time.Sleep(time.Millisecond)
					allowed++
				}
			}()
		}

		wg.Wait()

		if allowed < 95 || allowed > 105 {
			t.Errorf("Expected about 100 allowed requests, got %d", allowed)
		}
	})

	t.Run("GetMetrics", func(t *testing.T) {
		limiter := ratelimiter.NewSlidingWindow(&ratelimiter.Config{
			Rate:   10,
			Window: time.Second,
		})

		for i := 0; i < 15; i++ {
			limiter.Allow()
		}

		metrics := limiter.GetMetrics()

		if metrics.AllowedRequests != 10 {
			t.Errorf("Expected 10 allowed requests, got %d", metrics.AllowedRequests)
		}
		if metrics.CurrentRate != 10 {
			t.Errorf("Expected current rate 10, got %d", metrics.CurrentRate)
		}
		if metrics.WindowDuration != time.Second {
			t.Errorf("Expected window duration 1s, got %v", metrics.WindowDuration)
		}
	})
}
