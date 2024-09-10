package ratelimiter

import (
	"context"
	"time"
)

// MetricsWrapper wraps a Limiter and collects metrics
type MetricsWrapper struct {
	limiter   Limiter
	collector MetricsCollector
}

func NewMetricsWrapper(limiter Limiter, collector MetricsCollector) *MetricsWrapper {
	return &MetricsWrapper{
		limiter:   limiter,
		collector: collector,
	}
}

func (mw *MetricsWrapper) Allow() bool {
	mw.collector.IncrementTotalRequests()
	allowed := mw.limiter.Allow()
	if allowed {
		mw.collector.IncrementAllowedRequests()
	} else {
		mw.collector.IncrementDeniedRequests()
	}
	return allowed
}

func (mw *MetricsWrapper) AllowN(n int) bool {
	mw.collector.IncrementTotalRequests()
	allowed := mw.limiter.AllowN(n)
	if allowed {
		mw.collector.IncrementAllowedRequests()
	} else {
		mw.collector.IncrementDeniedRequests()
	}
	return allowed
}

func (mw *MetricsWrapper) Wait(ctx context.Context) error {
	mw.collector.IncrementTotalRequests()
	start := time.Now()
	err := mw.limiter.Wait(ctx)
	if err == nil {
		mw.collector.IncrementAllowedRequests()
		mw.collector.RecordWaitTime(time.Since(start))
	} else {
		mw.collector.IncrementDeniedRequests()
	}
	return err
}

func (mw *MetricsWrapper) WaitN(ctx context.Context, n int) error {
	mw.collector.IncrementTotalRequests()
	start := time.Now()
	err := mw.limiter.WaitN(ctx, n)
	if err == nil {
		mw.collector.IncrementAllowedRequests()
		mw.collector.RecordWaitTime(time.Since(start))
	} else {
		mw.collector.IncrementDeniedRequests()
	}
	return err
}

func (mw *MetricsWrapper) Reset() {
	mw.limiter.Reset()
	mw.collector.Reset()
}

func (mw *MetricsWrapper) GetMetrics() Metrics {
	return mw.collector.GetMetrics()
}
