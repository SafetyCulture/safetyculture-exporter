package util_test

import (
	"testing"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeFromString_RFC3339(t *testing.T) {
	input := "2006-01-02T15:04:05.000Z"
	result, err := util.TimeFromString(input)
	require.Nil(t, err)
	assert.EqualValues(t, "2006-01-02 15:04:05 +0000 UTC", result.String())
}

func TestTimeFromString_RFC1123(t *testing.T) {
	input := "Mon, 02 Jan 2006 15:04:05 MST"
	result, err := util.TimeFromString(input)
	require.Nil(t, err)
	assert.EqualValues(t, "2006-01-02 15:04:05 +0000 MST", result.String())
}

func TestTimeFromString_RFC822(t *testing.T) {
	input := "02 Jan 06 15:04 MST"
	result, err := util.TimeFromString(input)
	require.Nil(t, err)
	assert.EqualValues(t, "2006-01-02 15:04:00 +0000 MST", result.String())
}

func TestTimeFromString_RFC850(t *testing.T) {
	input := "Monday, 02-Jan-06 15:04:05 MST"
	result, err := util.TimeFromString(input)
	require.Nil(t, err)
	assert.EqualValues(t, "2006-01-02 15:04:05 +0000 MST", result.String())
}

func TestTimeFromString_ANSIC(t *testing.T) {
	input := "Mon Jan 2 15:04:05 2006"
	result, err := util.TimeFromString(input)
	require.Nil(t, err)
	assert.EqualValues(t, "2006-01-02 15:04:05 +0000 UTC", result.String())
}

func TestTimeFromString_UNIXDATE(t *testing.T) {
	input := "Mon Jan 2 15:04:05 MST 2006"
	result, err := util.TimeFromString(input)
	require.Nil(t, err)
	assert.EqualValues(t, "2006-01-02 15:04:05 +0000 MST", result.String())
}

func TestTimeFromString_ISO8601(t *testing.T) {
	input := "2006-01-02"
	result, err := util.TimeFromString(input)
	require.Nil(t, err)
	assert.EqualValues(t, "2006-01-02 00:00:00 +0000 UTC", result.String())
}

func TestTimeFromString_CUSTOM(t *testing.T) {
	input := "02 Jan 2006"
	result, err := util.TimeFromString(input)
	require.Nil(t, err)
	assert.EqualValues(t, "2006-01-02 00:00:00 +0000 UTC", result.String())
}
