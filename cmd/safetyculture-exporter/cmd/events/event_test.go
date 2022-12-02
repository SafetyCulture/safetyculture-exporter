package events_test

import (
	"fmt"
	"testing"

	"github.com/SafetyCulture/safetyculture-exporter/cmd/safetyculture-exporter/cmd/events"
	"github.com/stretchr/testify/assert"
)

func TestBuildEventError_INFO(t *testing.T) {
	infoNonFatal := events.BuildEventError(
		events.ErrorSeverityInfo,
		false,
		"simple message",
		"detailed message",
		fmt.Errorf("some error"),
	)
	assert.True(t, infoNonFatal.IsErrorType())
	assert.False(t, infoNonFatal.IsFeedInfoType())
	assert.True(t, infoNonFatal.Error.IsInfo())
	assert.EqualValues(t, "simple message", infoNonFatal.Error.SimpleMessage)
	assert.EqualValues(t, "detailed message", infoNonFatal.Error.DetailedMessage)
	assert.False(t, infoNonFatal.Error.IsFatal())
	assert.EqualValues(t, "some error", infoNonFatal.Error.Error())
}

func TestBuildEventError_WARNING(t *testing.T) {
	infoNonFatal := events.BuildEventError(
		events.ErrorSeverityWarning,
		false,
		"simple message",
		"detailed message",
		fmt.Errorf("some error"),
	)
	assert.True(t, infoNonFatal.IsErrorType())
	assert.False(t, infoNonFatal.IsFeedInfoType())
	assert.True(t, infoNonFatal.Error.IsWarn())
	assert.EqualValues(t, "simple message", infoNonFatal.Error.SimpleMessage)
	assert.EqualValues(t, "detailed message", infoNonFatal.Error.DetailedMessage)
	assert.False(t, infoNonFatal.Error.IsFatal())
	assert.EqualValues(t, "some error", infoNonFatal.Error.Error())
}

func TestBuildEventError_ERROR(t *testing.T) {
	infoNonFatal := events.BuildEventError(
		events.ErrorSeverityError,
		true,
		"simple message",
		"detailed message",
		fmt.Errorf("some error"),
	)
	assert.True(t, infoNonFatal.IsErrorType())
	assert.False(t, infoNonFatal.IsFeedInfoType())
	assert.True(t, infoNonFatal.Error.IsError())
	assert.EqualValues(t, "simple message", infoNonFatal.Error.SimpleMessage)
	assert.EqualValues(t, "detailed message", infoNonFatal.Error.DetailedMessage)
	assert.True(t, infoNonFatal.Error.IsFatal())
	assert.EqualValues(t, "some error", infoNonFatal.Error.Error())
}
