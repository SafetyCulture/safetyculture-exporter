package logger_test

import (
	"github.com/SafetyCulture/safetyculture-exporter/pkg/logger"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLogger_should_return_same_instance_every_time(t *testing.T) {
	t1 := logger.GetLogger()
	t2 := logger.GetLogger()

	assert.Equal(t, t1, t2)
}
