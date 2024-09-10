package ratelimiter

import (
	"context"
)

type Limiter interface {
	// Allow checks if a request is allowed
	Allow() bool

	// AllowN checks if N requests are allowed
	AllowN(n int) bool

	// Wait blocks until a request is allowed
	Wait(ctx context.Context) error

	// WaitN blocks until N requests are allowed
	WaitN(ctx context.Context, n int) error

	// Reset resets the limiter
	Reset()

	// GetMetrics returns the metrics for the limiter
	GetMetrics() Metrics
}

func New(opts ...Option) (Limiter, error) {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	var limiter Limiter
	var err error

	switch config.Algorithm {
	case "token_bucket":
		limiter = NewTokenBucket(config)
	case "fixed_window":
		limiter = NewFixedWindow(config)
	case "sliding_window":
		limiter = NewSlidingWindow(config)
	case "nested_window":
		limiter = NewNestedWindow(config)
	default:
		return nil, ErrUnsupportedAlgorithm
	}

	if config.MetricsEnabled {
		limiter = NewMetricsWrapper(limiter, NewMetricsCollector())
	}

	return limiter, err
}
