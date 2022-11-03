package feed_test

import (
	"testing"

	"github.com/SafetyCulture/safetyculture-exporter/internal/app/feed"
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

	dedupedList := feed.DeduplicateList(
		func(s *student) string { return s.id },
		originalList)

	require.Len(t, dedupedList, 4)
	assert.EqualValues(t, "José", dedupedList[0].name)
	assert.EqualValues(t, "Merry", dedupedList[1].name)
	assert.EqualValues(t, "Konstantin", dedupedList[2].name)
	assert.EqualValues(t, "Alfonso", dedupedList[3].name)
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
