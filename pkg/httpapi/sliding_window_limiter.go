package httpapi

import (
	"context"
	"sync"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/logger"
)

// SlidingWindowLimiter implements a sliding window rate limiting algorithm
// This matches server-side rate limiters like Istio/Envoy which use Redis-backed sliding windows
type SlidingWindowLimiter struct {
	requestsPerMinute int
	windowSize        time.Duration
	requests          []time.Time
	mu                sync.Mutex
	name              string

	// Stats
	requestCount  int64
	lastResetTime time.Time
	totalWaitTime time.Duration
}

// NewSlidingWindowLimiter creates a sliding window rate limiter
func NewSlidingWindowLimiter(requestsPerMinute int, name string) *SlidingWindowLimiter {
	windowSize := time.Minute

	logger.GetLogger().With(
		"rate_limiter", name,
		"type", "sliding_window",
		"requests_per_minute", requestsPerMinute,
		"window_size", windowSize.String(),
	).Info("sliding window rate limiter initialized")

	return &SlidingWindowLimiter{
		requestsPerMinute: requestsPerMinute,
		windowSize:        windowSize,
		requests:          make([]time.Time, 0, requestsPerMinute*2), // Pre-allocate with buffer
		name:              name,
		lastResetTime:     time.Now(),
	}
}

// Wait blocks until a request can proceed without exceeding the rate limit
func (sw *SlidingWindowLimiter) Wait(ctx context.Context) error {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-sw.windowSize)

	// Remove requests outside the sliding window
	validRequests := 0
	for _, reqTime := range sw.requests {
		if reqTime.After(windowStart) {
			sw.requests[validRequests] = reqTime
			validRequests++
		}
	}
	sw.requests = sw.requests[:validRequests]

	// Check if we're at the limit
	for len(sw.requests) >= sw.requestsPerMinute {
		// We're at the limit - need to wait until the oldest request falls out of the window
		oldestRequest := sw.requests[0]
		waitUntil := oldestRequest.Add(sw.windowSize)
		waitDuration := time.Until(waitUntil)

		if waitDuration <= 0 {
			// Oldest request has fallen out of window, remove it
			sw.requests = sw.requests[1:]
			continue
		}

		// Log if significant wait
		if waitDuration > 100*time.Millisecond {
			logger.GetLogger().With(
				"rate_limiter", sw.name,
				"wait_ms", waitDuration.Milliseconds(),
				"requests_in_window", len(sw.requests),
			).Debug("sliding window rate limiter delaying request")
		}

		// Unlock while waiting to allow other goroutines to check
		sw.mu.Unlock()

		select {
		case <-ctx.Done():
			sw.mu.Lock()
			return ctx.Err()
		case <-time.After(waitDuration):
			sw.mu.Lock()
			// Re-check after waiting (window may have moved)
			now = time.Now()
			windowStart = now.Add(-sw.windowSize)

			// Clean up old requests again
			validRequests := 0
			for _, reqTime := range sw.requests {
				if reqTime.After(windowStart) {
					sw.requests[validRequests] = reqTime
					validRequests++
				}
			}
			sw.requests = sw.requests[:validRequests]
		}
	}

	// Record this request
	sw.requests = append(sw.requests, now)

	// Update stats
	sw.requestCount++
	if now.Sub(sw.lastResetTime) >= time.Minute {
		logger.GetLogger().With(
			"rate_limiter", sw.name,
			"requests_last_minute", sw.requestCount,
			"requests_in_window", len(sw.requests),
		).Info("sliding window rate limiter stats")

		sw.requestCount = 0
		sw.lastResetTime = now
	}

	return nil
}

// GetStats returns current statistics
func (sw *SlidingWindowLimiter) GetStats() (requestsInWindow int, limit int) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-sw.windowSize)

	// Count requests in current window
	count := 0
	for _, reqTime := range sw.requests {
		if reqTime.After(windowStart) {
			count++
		}
	}

	return count, sw.requestsPerMinute
}
