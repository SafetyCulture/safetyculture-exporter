package util

import (
	"fmt"
	"regexp"
	"strconv"
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

// ParseDuration parses duration strings like "1d", "7d", "1w", "1m", "3m", "1y"
func ParseDuration(blockSize string) (time.Duration, error) {
	if blockSize == "" {
		return 0, fmt.Errorf("block size cannot be empty")
	}

	re := regexp.MustCompile(`^(\d+)([dwmy])$`)
	matches := re.FindStringSubmatch(blockSize)
	if len(matches) != 3 {
		return 0, fmt.Errorf("invalid block size format: %s (expected format: 1d, 1w, 1m, 1y)", blockSize)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value in block size: %s", matches[1])
	}

	if value <= 0 {
		return 0, fmt.Errorf("block size value must be positive: %d", value)
	}

	unit := matches[2]
	switch unit {
	case "d":
		return time.Duration(value) * 24 * time.Hour, nil
	case "w":
		return time.Duration(value) * 7 * 24 * time.Hour, nil
	case "m":
		return time.Duration(value) * 30 * 24 * time.Hour, nil
	case "y":
		return time.Duration(value) * 365 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unsupported time unit: %s (supported: d, w, m, y)", unit)
	}
}

// TimeBlock represents a time range with start and end times
type TimeBlock struct {
	Start time.Time
	End   time.Time
}

// GenerateTimeBlocks splits a time range into blocks of specified size
func GenerateTimeBlocks(start, end time.Time, blockSize time.Duration) []TimeBlock {
	if start.After(end) || start.Equal(end) {
		return []TimeBlock{}
	}

	var blocks []TimeBlock
	current := start

	for current.Before(end) {
		blockEnd := current.Add(blockSize)
		if blockEnd.After(end) {
			blockEnd = end
		}

		blocks = append(blocks, TimeBlock{
			Start: current,
			End:   blockEnd,
		})

		current = blockEnd
	}

	return blocks
}

// GenerateTimeBlocksFromString convenience function using string duration
func GenerateTimeBlocksFromString(start, end time.Time, blockSizeStr string) ([]TimeBlock, error) {
	if blockSizeStr == "" {
		return []TimeBlock{}, nil
	}

	blockSize, err := ParseDuration(blockSizeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse block size: %w", err)
	}

	return GenerateTimeBlocks(start, end, blockSize), nil
}
