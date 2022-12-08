package httpapi

import (
	"fmt"
	"math"
	"net/http"
	"time"
)

// CheckForRetry specifies a policy for handling retries. It is called
// following each request with the response and error values returned by
// the http.Client. If it returns false, the Client stops retrying
// and returns the response to the caller. If it returns an error,
// that error value is returned in lieu of the error from the request.
type CheckForRetry func(resp *http.Response, err error) (bool, error)

// DefaultRetryPolicy provides a default callback for CheckForRetry.
func DefaultRetryPolicy(resp *http.Response, err error) (bool, error) {
	// Retry for a genuine error
	if err != nil {
		return true, err
	}
	switch status := resp.StatusCode; {
	case status >= http.StatusInternalServerError:
		return true, nil
	case status == http.StatusTooManyRequests:
		return true, nil
	case status == http.StatusNotFound:
		return false, nil
	}
	return false, nil
}

// Backoff specifies a policy for how long to wait between retries.
// It is called after a failing request to determine the amount of time
// that should pass before trying again.
type Backoff func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration

// DefaultBackoff provides a default callback for Backoff which will perform
// exponential backoff based on the attempt number and limited by the provided
// minimum and maximum durations.
// It also tries to parse XRateLimitReset response header when a http.StatusTooManyRequests
// (HTTP Code 429) is found in the resp parameter. Hence it will return the number of
// seconds the server states it may be ready to process more requests from this client.
func DefaultBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	if resp != nil && resp.StatusCode == http.StatusTooManyRequests {

		if s := resp.Header.Get(string(XRateLimitReset)); s != "" {
			if sleep, err := time.ParseDuration(fmt.Sprintf("%ss", s)); err == nil {
				// Allow 1 second of allowance.
				return sleep + 1
			}
		}
	}

	mult := math.Pow(2, float64(attemptNum)) * float64(min)
	sleep := time.Duration(mult)
	if float64(sleep) != mult || sleep > max {
		sleep = max
	}
	return sleep
}
