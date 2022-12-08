package feed_test

import (
	"testing"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/feed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeduplicateList_when_not_empty(t *testing.T) {
	type student struct {
		id   string
		name string
	}

	originalList := []*student{
		{"ID1", "George"},
		{"ID2", "Maria"},
		{"ID3", "Konstantin"},
		{"ID1", "José"},
		{"ID2", "Merry"},
		{"ID4", "Alfonso"},
	}

	dedupedList := feed.DeduplicateList[student](
		func(s *student) string { return s.id },
		originalList)

	require.Len(t, dedupedList, 4)

	values := map[string]string{}
	for _, val := range dedupedList {
		values[val.id] = val.name
	}
	assert.EqualValues(t, "José", values["ID1"])
	assert.EqualValues(t, "Merry", values["ID2"])
	assert.EqualValues(t, "Konstantin", values["ID3"])
	assert.EqualValues(t, "Alfonso", values["ID4"])
}

func TestDeduplicateList_when_nil(t *testing.T) {
	type student struct {
		id   string
		name string
	}

	dedupedList := feed.DeduplicateList(
		func(s *student) string { return s.id },
		nil)

	assert.Empty(t, dedupedList)
}

func TestDeduplicateList_when_empty(t *testing.T) {
	type student struct {
		id   string
		name string
	}

	dedupedList := feed.DeduplicateList(
		func(s *student) string { return s.id },
		[]*student{})

	assert.Empty(t, dedupedList)
}
