package util

import (
	"time"
)

const TimeISO8601 = "2006-01-02"
const timeCustomDate = "02 Jan 2006"

// TimeFromString will try to match the input to known formats
func TimeFromString(input string) (time.Time, error) {
	var formats = []string{
		time.RFC3339, time.RFC822, time.RFC1123,
		time.RFC850, time.ANSIC, time.UnixDate,
		TimeISO8601, timeCustomDate,
	}
	var lastErr error

	for _, format := range formats {
		t, err := time.Parse(format, input)
		if err != nil {
			lastErr = err
			continue
		}
		return t, nil
	}
	return time.Time{}, lastErr
}
