package events_test

import (
	"fmt"
	"testing"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/events"
	"github.com/stretchr/testify/assert"
)

func TestBuildEventError_INFO(t *testing.T) {
	infoNonFatal := events.BuildNewEventError(
		events.ErrorSeverityInfo,
		events.ErrorSubSystemDB,
		false,
		fmt.Errorf("some error"),
	)
	assert.True(t, infoNonFatal.IsErrorType())
	assert.False(t, infoNonFatal.IsFeedInfoType())
	assert.True(t, infoNonFatal.Error.IsInfo())
	assert.False(t, infoNonFatal.Error.IsFatal())
	assert.EqualValues(t, "some error", infoNonFatal.Error.Error())
}

func TestBuildEventError_WARNING(t *testing.T) {
	infoNonFatal := events.BuildNewEventError(
		events.ErrorSeverityWarning,
		events.ErrorSubSystemDB,
		false,
		fmt.Errorf("some error"),
	)
	assert.True(t, infoNonFatal.IsErrorType())
	assert.False(t, infoNonFatal.IsFeedInfoType())
	assert.True(t, infoNonFatal.Error.IsWarn())
	assert.False(t, infoNonFatal.Error.IsFatal())
	assert.EqualValues(t, "some error", infoNonFatal.Error.Error())
}

func TestBuildEventError_ERROR(t *testing.T) {
	infoNonFatal := events.BuildNewEventError(
		events.ErrorSeverityError,
		events.ErrorSubSystemDB,
		true,
		fmt.Errorf("some error"),
	)
	assert.True(t, infoNonFatal.IsErrorType())
	assert.False(t, infoNonFatal.IsFeedInfoType())
	assert.True(t, infoNonFatal.Error.IsError())
	assert.True(t, infoNonFatal.Error.IsFatal())
	assert.EqualValues(t, "some error", infoNonFatal.Error.Error())
}
