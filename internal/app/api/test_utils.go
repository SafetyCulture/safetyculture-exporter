package api

import (
	"time"
)

// GetTestAPIClient creates a new test apiClient
func GetTestClient(opts ...Opt) *Client {
	apiClient := NewClient("http://localhost:9999", "abc123", opts...)
	apiClient.RetryWaitMin = 10 * time.Millisecond
	apiClient.RetryWaitMax = 10 * time.Millisecond
	apiClient.CheckForRetry = DefaultRetryPolicy
	apiClient.RetryMax = 1
	return apiClient
}
