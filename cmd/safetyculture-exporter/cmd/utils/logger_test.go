package util_test

import (
	"testing"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestGetLogger_should_return_same_instance_every_time(t *testing.T) {
	t1 := util.GetLogger()
	t2 := util.GetLogger()

	assert.Equal(t, t1, t2)
}
