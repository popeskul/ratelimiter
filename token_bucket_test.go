package ratelimiter_test

import (
	"context"
	"github.com/popeskul/ratelimiter"
	"testing"
	"time"
)

func TestTokenBucket(t *testing.T) {
	t.Run("Basic Allow", func(t *testing.T) {
		tb := ratelimiter.NewTokenBucket(&ratelimiter.Config{
			Rate:     10,
			Capacity: 10,
		})

		for i := 0; i < 10; i++ {
			if !tb.Allow() {
				t.Errorf("Request %d should be allowed", i+1)
			}
		}

		if tb.Allow() {
			t.Error("11th request should be denied")
		}
	})

	t.Run("AllowN", func(t *testing.T) {
		tb := ratelimiter.NewTokenBucket(&ratelimiter.Config{
			Rate:     10,
			Capacity: 10,
		})

		if !tb.AllowN(5) {
			t.Error("AllowN(5) should be allowed")
		}

		if !tb.AllowN(5) {
			t.Error("Second AllowN(5) should be allowed")
		}

		if tb.AllowN(1) {
			t.Error("AllowN(1) should be denied after consuming all tokens")
		}
	})

	t.Run("Wait", func(t *testing.T) {
		tb := ratelimiter.NewTokenBucket(&ratelimiter.Config{
			Rate:     10,
			Capacity: 1,
		})

		if err := tb.Wait(context.Background()); err != nil {
			t.Errorf("First Wait should not error: %v", err)
		}

		start := time.Now()
		if err := tb.Wait(context.Background()); err != nil {
			t.Errorf("Second Wait should not error: %v", err)
		}
		duration := time.Since(start)

		if duration < 90*time.Millisecond {
			t.Errorf("Expected to wait at least 90ms, but waited %v", duration)
		}
	})

	t.Run("WaitN", func(t *testing.T) {
		tb := ratelimiter.NewTokenBucket(&ratelimiter.Config{
			Rate:     10,
			Capacity: 10,
		})

		if err := tb.WaitN(context.Background(), 10); err != nil {
			t.Errorf("WaitN(10) should not error: %v", err)
		}

		start := time.Now()
		if err := tb.WaitN(context.Background(), 5); err != nil {
			t.Errorf("WaitN(5) should not error: %v", err)
		}
		duration := time.Since(start)

		if duration < 450*time.Millisecond {
			t.Errorf("Expected to wait at least 450ms, but waited %v", duration)
		}
	})

	t.Run("Reset", func(t *testing.T) {
		tb := ratelimiter.NewTokenBucket(&ratelimiter.Config{
			Rate:     10,
			Capacity: 10,
		})

		for i := 0; i < 10; i++ {
			tb.Allow()
		}

		if tb.Allow() {
			t.Error("Request should be denied before reset")
		}

		tb.Reset()

		if !tb.Allow() {
			t.Error("Request should be allowed after reset")
		}
	})

	t.Run("GetMetrics", func(t *testing.T) {
		tb := ratelimiter.NewTokenBucket(&ratelimiter.Config{
			Rate:     10,
			Capacity: 10,
		})

		for i := 0; i < 15; i++ {
			tb.Allow()
		}

		metrics := tb.GetMetrics()

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
	})

	t.Run("Refill", func(t *testing.T) {
		tb := ratelimiter.NewTokenBucket(&ratelimiter.Config{
			Rate:     10,
			Capacity: 10,
		})

		for i := 0; i < 10; i++ {
			tb.Allow()
		}

		if tb.Allow() {
			t.Error("Request should be denied immediately after consuming all tokens")
		}

		time.Sleep(110 * time.Millisecond)

		if !tb.Allow() {
			t.Error("Request should be allowed after waiting for token refill")
		}
	})
}
