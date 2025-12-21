package util_test

import (
	"testing"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestCalculateOptimalConcurrency(t *testing.T) {
	tests := map[string]struct {
		ratePerMinute   int
		totalBlocks     int
		maxWorkers      int
		expectedWorkers int
	}{
		"normal case": {
			ratePerMinute:   180,
			totalBlocks:     100,
			maxWorkers:      50,
			expectedWorkers: 30, // 180 / 6 = 30
		},
		"rate limited by max workers": {
			ratePerMinute:   600,
			totalBlocks:     100,
			maxWorkers:      50,
			expectedWorkers: 50, // 600 / 6 = 100, but capped at 50
		},
		"limited by total blocks": {
			ratePerMinute:   600,
			totalBlocks:     20,
			maxWorkers:      50,
			expectedWorkers: 20, // 600 / 6 = 100, but only 20 blocks
		},
		"very low rate": {
			ratePerMinute:   30,
			totalBlocks:     100,
			maxWorkers:      50,
			expectedWorkers: 5, // 30 / 6 = 5
		},
		"zero rate per minute": {
			ratePerMinute:   0,
			totalBlocks:     100,
			maxWorkers:      50,
			expectedWorkers: 1, // Minimum is 1
		},
		"negative rate per minute": {
			ratePerMinute:   -100,
			totalBlocks:     100,
			maxWorkers:      50,
			expectedWorkers: 1, // Minimum is 1
		},
		"one block": {
			ratePerMinute:   180,
			totalBlocks:     1,
			maxWorkers:      50,
			expectedWorkers: 1, // Only 1 block, so 1 worker
		},
		"zero blocks": {
			ratePerMinute:   180,
			totalBlocks:     0,
			maxWorkers:      50,
			expectedWorkers: 0, // No blocks, so 0 workers
		},
		"very high rate": {
			ratePerMinute:   1000,
			totalBlocks:     1000,
			maxWorkers:      50,
			expectedWorkers: 50, // 1000 / 6 = 166, but capped at 50
		},
		"exact division": {
			ratePerMinute:   120, // 120 / 6 = 20
			totalBlocks:     100,
			maxWorkers:      50,
			expectedWorkers: 20,
		},
		"rate results in less than 1 worker": {
			ratePerMinute:   5, // 5 / 6 = 0.83
			totalBlocks:     100,
			maxWorkers:      50,
			expectedWorkers: 1, // Minimum is 1
		},
		"all equal": {
			ratePerMinute:   60, // 60 / 6 = 10
			totalBlocks:     10,
			maxWorkers:      10,
			expectedWorkers: 10,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := util.CalculateOptimalConcurrency(tt.ratePerMinute, tt.totalBlocks, tt.maxWorkers)
			assert.Equal(t, tt.expectedWorkers, result,
				"ratePerMinute=%d, totalBlocks=%d, maxWorkers=%d",
				tt.ratePerMinute, tt.totalBlocks, tt.maxWorkers)
		})
	}
}

func TestCalculateOptimalConcurrency_EdgeCases(t *testing.T) {
	// Test boundary between 0 and 1 worker
	result := util.CalculateOptimalConcurrency(1, 100, 50)
	assert.Equal(t, 1, result, "Should return at least 1 worker for positive rate")

	// Test with very large numbers
	result = util.CalculateOptimalConcurrency(10000, 10000, 1000)
	assert.Equal(t, 1000, result, "Should be capped at maxWorkers")

	// Test with blocks as limiting factor
	result = util.CalculateOptimalConcurrency(1000, 5, 100)
	assert.Equal(t, 5, result, "Should not exceed totalBlocks")
}

func TestCalculateOptimalConcurrency_EstimatedReqPerBlock(t *testing.T) {
	// The function uses estimatedReqPerBlockPerMinute = 6
	// So for 180 req/min, we should get 30 workers (180 / 6)
	result := util.CalculateOptimalConcurrency(180, 100, 100)
	assert.Equal(t, 30, result, "180 req/min should result in 30 workers")

	// For 60 req/min, we should get 10 workers (60 / 6)
	result = util.CalculateOptimalConcurrency(60, 100, 100)
	assert.Equal(t, 10, result, "60 req/min should result in 10 workers")

	// For 360 req/min, we should get 60 workers (360 / 6)
	result = util.CalculateOptimalConcurrency(360, 100, 100)
	assert.Equal(t, 60, result, "360 req/min should result in 60 workers")
}
