package version_test

import (
	"github.com/SafetyCulture/safetyculture-exporter/pkg/external/version"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVersion(t *testing.T) {
	r := version.GetVersion()
	assert.EqualValues(t, "0.0.0-dev", r)
}
