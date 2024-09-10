package ratelimiter

import "time"

type Algorithm string

const (
	TokenBucketAlgorithm   Algorithm = "token_bucket"
	FixedWindowAlgorithm   Algorithm = "fixed_window"
	SlidingWindowAlgorithm Algorithm = "sliding_window"
	NestedWindowAlgorithm  Algorithm = "nested_window"
)

// Option func is a function that takes a pointer to Config and modifies it
type Option func(*Config)

// Config struct contains the configuration for the rate limiter
type Config struct {
	Rate           int           // Number of allowed requests per unit of time
	Burst          int           // The maximum number of requests that can be executed at once
	Capacity       int           // Maximum number of tokens in the bucket (for Token Bucket)
	Window         time.Duration // Time window for window-based algorithms
	InnerWindow    time.Duration // Inner time window for Nested Window
	Algorithm      Algorithm     // Algorithm to use for rate limiting
	MetricsEnabled bool          // Enable metrics for the rate limiter
}

// WithRate sets the Rate for Config
func WithRate(rate int) Option {
	return func(c *Config) {
		c.Rate = rate
	}
}

// WithBurst sets the Burst for Config
func WithBurst(burst int) Option {
	return func(c *Config) {
		c.Burst = burst
	}
}

// WithCapacity sets the Capacity for Config
func WithCapacity(capacity int) Option {
	return func(c *Config) {
		c.Capacity = capacity
	}
}

// WithWindow sets the Window for Config
func WithWindow(window time.Duration) Option {
	return func(c *Config) {
		c.Window = window
	}
}

// WithInnerWindow sets the InnerWindow for Config (for Nested Window)
func WithInnerWindow(innerWindow time.Duration) Option {
	return func(c *Config) {
		c.InnerWindow = innerWindow
	}
}

// WithAlgorithm sets the Algorithm for Config
func WithAlgorithm(algo Algorithm) Option {
	return func(c *Config) {
		c.Algorithm = algo
	}
}

// WithMetrics sets the MetricsEnabled for Config
func WithMetrics(enabled bool) Option {
	return func(c *Config) {
		c.MetricsEnabled = enabled
	}
}

// DefaultConfig returns the default configuration for the rate limiter
func DefaultConfig() *Config {
	return &Config{
		Rate:           100,
		Burst:          1,
		Capacity:       100,
		Window:         time.Minute,
		InnerWindow:    time.Second * 6,
		Algorithm:      "token_bucket",
		MetricsEnabled: false,
	}
}
