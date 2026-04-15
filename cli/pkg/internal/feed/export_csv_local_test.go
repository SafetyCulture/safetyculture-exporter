package feed

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FileExists_should_return_false_when_missing_file(t *testing.T) {
	result, err := fileExists("xyz123")
	assert.NoError(t, err)
	assert.False(t, result)
}
