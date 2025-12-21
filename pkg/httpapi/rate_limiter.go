package httpapi

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/logger"
	"golang.org/x/time/rate"
)

// RateLimiterAlgorithm specifies which rate limiting algorithm to use
type RateLimiterAlgorithm string

const (
	// AlgorithmSlidingWindow uses a sliding window algorithm (matches Istio/Envoy behavior)
	AlgorithmSlidingWindow RateLimiterAlgorithm = "sliding_window"
	// AlgorithmTokenBucket uses a token bucket algorithm (allows bursts)
	AlgorithmTokenBucket RateLimiterAlgorithm = "token_bucket"
)

// RateLimiterInterface defines the interface for rate limiters
type RateLimiterInterface interface {
	Wait(ctx context.Context) error
}

// RateLimiter manages API request rate limiting
// It can use either token bucket or sliding window algorithm
type RateLimiter struct {
	// Token bucket implementation
	limiter       *rate.Limiter
	mu            sync.RWMutex
	enabled       bool
	requestCount  atomic.Int64
	lastResetTime atomic.Int64
	totalWaitTime atomic.Int64 // Total time spent waiting in nanoseconds
	name          string       // Name for logging (e.g., "inspections", "inspection_items")

	// Sliding window implementation (preferred for matching server-side behavior)
	slidingWindow *SlidingWindowLimiter
	algorithm     RateLimiterAlgorithm
}

// RateLimiterConfig configures rate limiting behavior
type RateLimiterConfig struct {
	// RequestsPerMinute is the target rate (e.g., 180)
	RequestsPerMinute int
	// BurstSize allows burst requests (only used for token bucket algorithm)
	BurstSize int
	// Enabled controls whether rate limiting is active
	Enabled bool
	// Name is used for logging and debugging
	Name string
	// Algorithm specifies which algorithm to use (defaults to sliding window)
	Algorithm RateLimiterAlgorithm
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	if !config.Enabled || config.RequestsPerMinute <= 0 {
		return &RateLimiter{enabled: false, name: config.Name}
	}

	// Default to sliding window if not specified (matches server-side rate limiters)
	algorithm := config.Algorithm
	if algorithm == "" {
		algorithm = AlgorithmSlidingWindow
	}

	rl := &RateLimiter{
		enabled:   true,
		name:      config.Name,
		algorithm: algorithm,
	}
	rl.lastResetTime.Store(time.Now().Unix())

	if algorithm == AlgorithmSlidingWindow {
		// Use sliding window algorithm (matches Istio/Envoy/Redis rate limiters)
		rl.slidingWindow = NewSlidingWindowLimiter(config.RequestsPerMinute, config.Name)
	} else {
		// Use token bucket algorithm (allows bursts)
		requestsPerSecond := float64(config.RequestsPerMinute) / 60.0

		// Default burst size to same as rate if not specified
		burstSize := config.BurstSize
		if burstSize <= 0 {
			burstSize = config.RequestsPerMinute
		}

		rl.limiter = rate.NewLimiter(rate.Limit(requestsPerSecond), burstSize)

		logger.GetLogger().With(
			"rate_limiter", config.Name,
			"type", "token_bucket",
			"requests_per_minute", config.RequestsPerMinute,
			"burst_size", burstSize,
		).Info("rate limiter initialized")
	}

	return rl
}

// Wait blocks until a token is available or context is cancelled
func (rl *RateLimiter) Wait(ctx context.Context) error {
	if !rl.enabled {
		return nil
	}

	// Delegate to appropriate implementation
	if rl.algorithm == AlgorithmSlidingWindow {
		return rl.slidingWindow.Wait(ctx)
	}

	// Token bucket implementation (legacy)
	rl.mu.RLock()
	limiter := rl.limiter
	rl.mu.RUnlock()

	// Track request count and reset counter every minute
	now := time.Now()
	currentMinute := now.Unix() / 60
	lastResetMinute := rl.lastResetTime.Load() / 60

	if currentMinute > lastResetMinute {
		oldCount := rl.requestCount.Swap(0)
		rl.lastResetTime.Store(now.Unix())
		avgWaitMs := float64(0)
		totalWait := rl.totalWaitTime.Swap(0)
		if oldCount > 0 {
			avgWaitMs = float64(totalWait) / float64(oldCount) / 1e6 // Convert to milliseconds
		}

		logger.GetLogger().With(
			"rate_limiter", rl.name,
			"type", "token_bucket",
			"requests_last_minute", oldCount,
			"avg_wait_ms", avgWaitMs,
		).Info("rate limiter stats")
	}

	// Wait for token and measure wait time
	start := time.Now()
	err := limiter.Wait(ctx)
	waitDuration := time.Since(start)

	// Track metrics
	rl.requestCount.Add(1)
	rl.totalWaitTime.Add(int64(waitDuration))

	// Log if we had to wait significantly
	if waitDuration > 100*time.Millisecond {
		logger.GetLogger().With(
			"rate_limiter", rl.name,
			"type", "token_bucket",
			"wait_ms", waitDuration.Milliseconds(),
		).Debug("rate limiter delayed request")
	}

	return err
}

// UpdateRate dynamically updates the rate limit
func (rl *RateLimiter) UpdateRate(requestsPerMinute int) {
	if !rl.enabled {
		return
	}

	if rl.algorithm == AlgorithmSlidingWindow {
		rl.slidingWindow.mu.Lock()
		rl.slidingWindow.requestsPerMinute = requestsPerMinute
		rl.slidingWindow.mu.Unlock()
		return
	}

	// Token bucket implementation
	requestsPerSecond := float64(requestsPerMinute) / 60.0

	rl.mu.Lock()
	rl.limiter.SetLimit(rate.Limit(requestsPerSecond))
	rl.mu.Unlock()
}

// GetCurrentRate returns the current rate in requests per minute
func (rl *RateLimiter) GetCurrentRate() int {
	if !rl.enabled {
		return 0
	}

	if rl.algorithm == AlgorithmSlidingWindow {
		rl.slidingWindow.mu.Lock()
		defer rl.slidingWindow.mu.Unlock()
		return rl.slidingWindow.requestsPerMinute
	}

	// Token bucket implementation
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return int(float64(rl.limiter.Limit()) * 60.0)
}
