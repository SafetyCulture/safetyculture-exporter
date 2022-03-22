package feed

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_FileExists_should_return_false_when_missing_file(t *testing.T) {
	result, err := fileExists("xyz123")
	assert.Nil(t, err)
	assert.False(t, result)
}
