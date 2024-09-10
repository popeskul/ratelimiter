package ratelimiter

import (
	"sync/atomic"
	"time"
)

// Metrics contains the metrics for the rate limiter
type Metrics struct {
	TotalRequests   int64
	AllowedRequests int64
	DeniedRequests  int64
	CurrentRate     int64
	LastResetTime   int64
	TotalWaitTime   int64
	MaxWaitTime     int64
	WindowDuration  time.Duration
	InnerRate       int64
	InnerWindow     time.Duration
}

// MetricsCollector collects metrics for the rate limiter
type MetricsCollector interface {
	IncrementTotalRequests()
	IncrementAllowedRequests()
	IncrementDeniedRequests()
	UpdateCurrentRate(rate int64)
	RecordWaitTime(waitTime time.Duration)
	GetMetrics() Metrics
	Reset()
}

// DefaultMetricsCollector implements the MetricsCollector interface
type DefaultMetricsCollector struct {
	metrics Metrics
}

// NewMetricsCollector creates a new DefaultMetricsCollector
func NewMetricsCollector() *DefaultMetricsCollector {
	return &DefaultMetricsCollector{
		metrics: Metrics{LastResetTime: time.Now().UnixNano()},
	}
}

func (mc *DefaultMetricsCollector) IncrementTotalRequests() {
	atomic.AddInt64(&mc.metrics.TotalRequests, 1)
}

func (mc *DefaultMetricsCollector) IncrementAllowedRequests() {
	atomic.AddInt64(&mc.metrics.AllowedRequests, 1)
}

func (mc *DefaultMetricsCollector) IncrementDeniedRequests() {
	atomic.AddInt64(&mc.metrics.DeniedRequests, 1)
}

func (mc *DefaultMetricsCollector) UpdateCurrentRate(rate int64) {
	atomic.StoreInt64(&mc.metrics.CurrentRate, rate)
}

func (mc *DefaultMetricsCollector) RecordWaitTime(waitTime time.Duration) {
	atomic.AddInt64(&mc.metrics.TotalWaitTime, int64(waitTime))
	for {
		oldMax := atomic.LoadInt64(&mc.metrics.MaxWaitTime)
		newMax := int64(waitTime)
		if newMax <= oldMax {
			break
		}
		if atomic.CompareAndSwapInt64(&mc.metrics.MaxWaitTime, oldMax, newMax) {
			break
		}
	}
}

func (mc *DefaultMetricsCollector) GetMetrics() Metrics {
	return Metrics{
		TotalRequests:   atomic.LoadInt64(&mc.metrics.TotalRequests),
		AllowedRequests: atomic.LoadInt64(&mc.metrics.AllowedRequests),
		DeniedRequests:  atomic.LoadInt64(&mc.metrics.DeniedRequests),
		CurrentRate:     atomic.LoadInt64(&mc.metrics.CurrentRate),
		LastResetTime:   atomic.LoadInt64(&mc.metrics.LastResetTime),
		TotalWaitTime:   atomic.LoadInt64(&mc.metrics.TotalWaitTime),
		MaxWaitTime:     atomic.LoadInt64(&mc.metrics.MaxWaitTime),
	}
}

func (mc *DefaultMetricsCollector) Reset() {
	atomic.StoreInt64(&mc.metrics.TotalRequests, 0)
	atomic.StoreInt64(&mc.metrics.AllowedRequests, 0)
	atomic.StoreInt64(&mc.metrics.DeniedRequests, 0)
	atomic.StoreInt64(&mc.metrics.CurrentRate, 0)
	atomic.StoreInt64(&mc.metrics.LastResetTime, time.Now().UnixNano())
	atomic.StoreInt64(&mc.metrics.TotalWaitTime, 0)
	atomic.StoreInt64(&mc.metrics.MaxWaitTime, 0)
}
