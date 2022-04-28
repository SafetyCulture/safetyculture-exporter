package version_test

import (
	"github.com/SafetyCulture/iauditor-exporter/internal/app/version"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetVersion(t *testing.T) {
	r := version.GetVersion()
	assert.EqualValues(t, "0.0.0-dev", r)
}
