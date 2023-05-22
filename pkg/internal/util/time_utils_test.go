package util_test

import (
	"fmt"
	"testing"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeFromString(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
		err      error
	}{
		"RFC3339": {
			input:    "2023-04-02T15:04:05.000Z",
			expected: "2023-04-02 15:04:05 +0000 UTC",
		},
		"RFC1123": {
			input:    "Mon, 02 Jan 2006 15:04:05 MST",
			expected: "2006-01-02 15:04:05 +0000 MST",
		},
		"RFC822": {
			input:    "02 Jan 06 15:04 MST",
			expected: "2006-01-02 15:04:00 +0000 MST",
		},
		"RFC850": {
			input:    "Monday, 02-Jan-06 15:04:05 MST",
			expected: "2006-01-02 15:04:05 +0000 MST",
		},
		"ANSIC": {
			input:    "Mon Jan 2 15:04:05 2006",
			expected: "2006-01-02 15:04:05 +0000 UTC",
		},
		"UNIXDATE": {
			input:    "Mon Jan 2 15:04:05 MST 2006",
			expected: "2006-01-02 15:04:05 +0000 MST",
		},
		"ISO8601": {
			input:    "2006-01-02",
			expected: "2006-01-02 00:00:00 +0000 UTC",
		},
		"CUSTOM": {
			input:    "02 Jan 2006",
			expected: "2006-01-02 00:00:00 +0000 UTC",
		},
		"INVALID": {
			input: "invalid",
			err:   fmt.Errorf(`parsing time "invalid" as "02 Jan 2006": cannot parse "invalid" as "02"`),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := util.TimeFromString(test.input)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else if test.err == nil {
				assert.EqualValues(t, test.expected, result.String())
			} else {
				require.Fail(t, "unexpected error")
			}
		})
	}
}
