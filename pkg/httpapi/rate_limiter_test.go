package httpapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimiter_Enabled(t *testing.T) {
	config := httpapi.RateLimiterConfig{
		RequestsPerMinute: 180,
		BurstSize:         180,
		Enabled:           true,
	}

	limiter := httpapi.NewRateLimiter(config)
	require.NotNil(t, limiter)
}

func TestNewRateLimiter_Disabled(t *testing.T) {
	config := httpapi.RateLimiterConfig{
		RequestsPerMinute: 0,
		BurstSize:         0,
		Enabled:           false,
	}

	limiter := httpapi.NewRateLimiter(config)
	require.NotNil(t, limiter)

	// Disabled limiter should not block
	ctx := context.Background()
	err := limiter.Wait(ctx)
	assert.NoError(t, err)
}

func TestNewRateLimiter_DisabledWhenRateZero(t *testing.T) {
	config := httpapi.RateLimiterConfig{
		RequestsPerMinute: 0,
		BurstSize:         10,
		Enabled:           true,
	}

	limiter := httpapi.NewRateLimiter(config)
	require.NotNil(t, limiter)

	// Should be disabled when RequestsPerMinute is 0
	ctx := context.Background()
	err := limiter.Wait(ctx)
	assert.NoError(t, err)
}

func TestRateLimiter_Wait_AllowsInitialBurst(t *testing.T) {
	config := httpapi.RateLimiterConfig{
		RequestsPerMinute: 60, // 1 per second
		BurstSize:         5,  // Allow burst of 5
		Enabled:           true,
		Algorithm:         httpapi.AlgorithmTokenBucket, // Use token bucket for this test
	}

	limiter := httpapi.NewRateLimiter(config)
	ctx := context.Background()

	// First 5 requests should complete immediately (burst)
	start := time.Now()
	for i := 0; i < 5; i++ {
		err := limiter.Wait(ctx)
		assert.NoError(t, err)
	}
	elapsed := time.Since(start)

	// Should complete very quickly (under 100ms for 5 requests)
	assert.Less(t, elapsed, 100*time.Millisecond, "Burst requests should complete quickly")
}

func TestRateLimiter_Wait_EnforcesRate(t *testing.T) {
	config := httpapi.RateLimiterConfig{
		RequestsPerMinute: 120, // 2 per second
		BurstSize:         1,   // No burst
		Enabled:           true,
		Algorithm:         httpapi.AlgorithmTokenBucket, // Use token bucket for this test
	}

	limiter := httpapi.NewRateLimiter(config)
	ctx := context.Background()

	// First request should be immediate
	err := limiter.Wait(ctx)
	assert.NoError(t, err)

	// Second request should wait ~500ms
	start := time.Now()
	err = limiter.Wait(ctx)
	assert.NoError(t, err)
	elapsed := time.Since(start)

	// Should wait approximately 500ms (with some tolerance)
	assert.GreaterOrEqual(t, elapsed, 400*time.Millisecond, "Should enforce rate limit")
	assert.Less(t, elapsed, 700*time.Millisecond, "Should not wait too long")
}

func TestRateLimiter_Wait_RespectsContext(t *testing.T) {
	config := httpapi.RateLimiterConfig{
		RequestsPerMinute: 60, // 1 per second
		BurstSize:         1,
		Enabled:           true,
		Algorithm:         httpapi.AlgorithmTokenBucket, // Use token bucket for this test
	}

	limiter := httpapi.NewRateLimiter(config)

	// Create a context that will be canceled
	ctx, cancel := context.WithCancel(context.Background())

	// Use up the burst
	err := limiter.Wait(ctx)
	assert.NoError(t, err)

	// Cancel the context before waiting
	cancel()

	// Wait should return error immediately
	err = limiter.Wait(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestRateLimiter_Wait_RespectsTimeout(t *testing.T) {
	config := httpapi.RateLimiterConfig{
		RequestsPerMinute: 60, // 1 per second
		BurstSize:         1,
		Enabled:           true,
		Algorithm:         httpapi.AlgorithmTokenBucket, // Use token bucket for this test
	}

	limiter := httpapi.NewRateLimiter(config)

	// Use up the burst
	ctx := context.Background()
	err := limiter.Wait(ctx)
	assert.NoError(t, err)

	// Create context with short timeout
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Wait should timeout before rate limit allows
	start := time.Now()
	err = limiter.Wait(ctxWithTimeout)
	elapsed := time.Since(start)

	assert.Error(t, err)
	// Error message can vary but should indicate deadline/timeout
	assert.True(t,
		err.Error() == "context deadline exceeded" ||
			err.Error() == "rate: Wait(n=1) would exceed context deadline",
		"Expected deadline error, got: %v", err)
	assert.Less(t, elapsed, 200*time.Millisecond, "Should timeout quickly")
}

func TestRateLimiter_UpdateRate(t *testing.T) {
	// Test both algorithms
	algorithms := []httpapi.RateLimiterAlgorithm{
		httpapi.AlgorithmTokenBucket,
		httpapi.AlgorithmSlidingWindow,
	}

	for _, algo := range algorithms {
		t.Run(string(algo), func(t *testing.T) {
			config := httpapi.RateLimiterConfig{
				RequestsPerMinute: 60, // 1 per second
				BurstSize:         60,
				Enabled:           true,
				Algorithm:         algo,
			}

			limiter := httpapi.NewRateLimiter(config)

			// Verify initial rate
			assert.Equal(t, 60, limiter.GetCurrentRate())

			// Update to higher rate
			limiter.UpdateRate(120) // 2 per second

			// Verify new rate
			assert.Equal(t, 120, limiter.GetCurrentRate())

			// Verify rate limiter is still functional
			ctx := context.Background()
			err := limiter.Wait(ctx)
			assert.NoError(t, err)
		})
	}
}

func TestRateLimiter_DefaultBurstSize(t *testing.T) {
	config := httpapi.RateLimiterConfig{
		RequestsPerMinute: 180,
		BurstSize:         0, // Should default to RequestsPerMinute
		Enabled:           true,
		Algorithm:         httpapi.AlgorithmTokenBucket, // Use token bucket for this test
	}

	limiter := httpapi.NewRateLimiter(config)
	ctx := context.Background()

	// Should allow a burst equal to RequestsPerMinute
	start := time.Now()
	for i := 0; i < 180; i++ {
		err := limiter.Wait(ctx)
		assert.NoError(t, err)
	}
	elapsed := time.Since(start)

	// Should complete quickly (burst)
	assert.Less(t, elapsed, 500*time.Millisecond, "Should allow burst of 180 requests")
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	config := httpapi.RateLimiterConfig{
		RequestsPerMinute: 600, // 10 per second
		BurstSize:         100,
		Enabled:           true,
		Algorithm:         httpapi.AlgorithmTokenBucket, // Use token bucket for this test
	}

	limiter := httpapi.NewRateLimiter(config)
	ctx := context.Background()

	// Simulate concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			err := limiter.Wait(ctx)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestRateLimiter_HighRateLimit(t *testing.T) {
	config := httpapi.RateLimiterConfig{
		RequestsPerMinute: 1000, // Very high rate
		BurstSize:         1000,
		Enabled:           true,
		Algorithm:         httpapi.AlgorithmTokenBucket, // Use token bucket for this test
	}

	limiter := httpapi.NewRateLimiter(config)
	ctx := context.Background()

	// Should handle high rate limits without issues
	start := time.Now()
	for i := 0; i < 100; i++ {
		err := limiter.Wait(ctx)
		assert.NoError(t, err)
	}
	elapsed := time.Since(start)

	// Should complete very quickly
	assert.Less(t, elapsed, 500*time.Millisecond, "High rate limit should allow fast requests")
}

func TestRateLimiter_ZeroAndNegativeValues(t *testing.T) {
	tests := []struct {
		name   string
		config httpapi.RateLimiterConfig
	}{
		{
			name: "zero_rate",
			config: httpapi.RateLimiterConfig{
				RequestsPerMinute: 0,
				BurstSize:         10,
				Enabled:           true,
			},
		},
		{
			name: "negative_rate",
			config: httpapi.RateLimiterConfig{
				RequestsPerMinute: -100,
				BurstSize:         10,
				Enabled:           true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := httpapi.NewRateLimiter(tt.config)
			ctx := context.Background()

			// Should not block (treated as disabled)
			err := limiter.Wait(ctx)
			assert.NoError(t, err)
		})
	}
}
