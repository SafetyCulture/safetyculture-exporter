package util

// CalculateOptimalConcurrency determines optimal worker count based on rate limit
//
// Logic:
// - Each worker processes one block at a time
// - Each block makes multiple paginated requests (unknown count)
// - Conservative approach: assume each block uses ~6 requests/min on average
// - With 180 req/min = 3 req/sec, and assuming ~6 req per block per minute
// - Optimal workers = min(blocks, floor(rate / expected_req_per_block_per_min))
//
// Parameters:
// - ratePerMinute: API rate limit (e.g., 180)
// - totalBlocks: Number of time blocks to process
// - maxWorkers: Maximum workers allowed (safety limit)
//
// Returns: Number of workers to use
func CalculateOptimalConcurrency(ratePerMinute int, totalBlocks int, maxWorkers int) int {
	if ratePerMinute <= 0 {
		return 1 // Fallback to sequential
	}

	// Conservative estimate: each block uses ~6 req/min on average
	// This accounts for pagination, retries, etc.
	estimatedReqPerBlockPerMinute := 6

	// Calculate how many blocks we can process concurrently
	optimalWorkers := ratePerMinute / estimatedReqPerBlockPerMinute

	// Apply constraints
	if optimalWorkers < 1 {
		optimalWorkers = 1
	}
	if optimalWorkers > totalBlocks {
		optimalWorkers = totalBlocks
	}
	if optimalWorkers > maxWorkers {
		optimalWorkers = maxWorkers
	}

	return optimalWorkers
}
