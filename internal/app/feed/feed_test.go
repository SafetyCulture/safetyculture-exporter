package feed_test

import (
	"github.com/SafetyCulture/iauditor-exporter/internal/app/feed"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetLogString_WhenNoArgumentsArePassed(t *testing.T) {
	result := feed.GetLogString("TEST_FEED", nil)
	assert.EqualValues(t, "TEST_FEED: ", result)
}

func TestGetLogString_WhenMinimumArgumentsArePassed(t *testing.T) {
	cfg := feed.LogStringConfig{RemainingRecords: 10}
	result := feed.GetLogString("TEST_FEED", &cfg)
	assert.EqualValues(t, "TEST_FEED: 10 remaining.", result)
}

func TestGetLogString_WhenHTTPDuration(t *testing.T) {
	cfg := feed.LogStringConfig{
		RemainingRecords: 40,
		HTTPDuration:     time.Second * 5,
	}
	result := feed.GetLogString("TEST_FEED", &cfg)
	assert.EqualValues(t, "TEST_FEED: 40 remaining. Last http call was 5000ms.", result)
}

func TestGetLogString_WhenExporterDuration(t *testing.T) {
	cfg := feed.LogStringConfig{
		RemainingRecords: 20,
		ExporterDuration: time.Second * 2,
	}
	result := feed.GetLogString("TEST_FEED", &cfg)
	assert.EqualValues(t, "TEST_FEED: 20 remaining. Last export operation was 2000ms.", result)
}

func TestGetLogString_WhenAllDataIsPresent(t *testing.T) {
	cfg := feed.LogStringConfig{
		RemainingRecords: 50,
		ExporterDuration: time.Second * 1,
		HTTPDuration:     time.Millisecond * 44,
	}
	result := feed.GetLogString("TEST_FEED", &cfg)
	assert.EqualValues(t, "TEST_FEED: 50 remaining. Last http call was 44ms. Last export operation was 1000ms.", result)
}
