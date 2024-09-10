package ratelimiter_test

import (
	"context"
	"github.com/popeskul/ratelimiter"
	"testing"
	"time"
)

func TestMetrics(t *testing.T) {
	limiter, err := ratelimiter.New(
		ratelimiter.WithAlgorithm("token_bucket"),
		ratelimiter.WithRate(10),
		ratelimiter.WithCapacity(10),
		ratelimiter.WithMetrics(true),
	)
	if err != nil {
		t.Fatalf("Failed to create limiter: %v", err)
	}

	// Allow 5 requests
	for i := 0; i < 5; i++ {
		limiter.Allow()
	}

	// Try to allow 10 more requests (5 should be allowed, 5 should be denied)
	for i := 0; i < 10; i++ {
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

	// Test Reset
	limiter.Reset()
	metrics = limiter.GetMetrics()

	if metrics.TotalRequests != 0 {
		t.Errorf("Expected 0 total requests after reset, got %d", metrics.TotalRequests)
	}

	// Test wait time metrics
	limiter, _ = ratelimiter.New(
		ratelimiter.WithAlgorithm("token_bucket"),
		ratelimiter.WithRate(1),
		ratelimiter.WithCapacity(1),
		ratelimiter.WithMetrics(true),
	)

	limiter.Allow() // Use the only available token

	start := time.Now()
	if err = limiter.Wait(context.Background()); err != nil {
		t.Fatalf("Failed to wait: %v", err)
	}
	waitTime := time.Since(start)

	metrics = limiter.GetMetrics()

	tolerance := 1 * time.Millisecond // 1ms tolerance for time measurements

	if time.Duration(metrics.TotalWaitTime) < waitTime-tolerance {
		t.Errorf("Expected total wait time to be at least %v, got %v", waitTime, time.Duration(metrics.TotalWaitTime))
	}

	if time.Duration(metrics.MaxWaitTime) < waitTime-tolerance {
		t.Errorf("Expected max wait time to be at least %v, got %v", waitTime, time.Duration(metrics.MaxWaitTime))
	}
}
