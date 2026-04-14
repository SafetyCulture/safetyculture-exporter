package feed

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "carriage return is removed",
			input:    "Foo\rBar",
			expected: "Foo-Bar",
		},
		{
			name:     "CRLF is removed",
			input:    "Foo\r\nBar",
			expected: "Foo--Bar",
		},
		{
			name:     "address with embedded CRLF line breaks",
			input:    "Acme Corp, 10 High Street\r\n\r\n, W1A 1AA, - REF- 12345",
			expected: "Acme-Corp,-10-High-Street----,-W1A-1AA,---REF--12345",
		},
		{
			name:     "forward slashes are removed",
			input:    "Test / Name // Here",
			expected: "Test-Name-Here",
		},
		{
			name:     "special characters are removed",
			input:    "Name?with*special|chars:<>\"end",
			expected: "Name-with-special-chars----end",
		},
		{
			name:     "tabs and newlines are removed",
			input:    "Name\twith\nnewlines",
			expected: "Name-with-newlines",
		},
		{
			name:     "plain name unchanged",
			input:    "Simple-Name",
			expected: "Simple-Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeName(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.NotContains(t, result, "\r")
			assert.NotContains(t, result, "\n")
		})
	}
}
