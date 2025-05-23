package util_test

import (
	"fmt"
	"testing"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitSliceInBatch(t *testing.T) {
	tests := map[string]struct {
		size        int
		collection  []string
		fn          func([]string) error
		expectedErr error
	}{
		"batch size is 0": {
			size:        0,
			expectedErr: fmt.Errorf("batch size cannot be 0"),
		},
		"batch size is greater than the collection size": {
			size:       100,
			collection: []string{"a", "b", "c", "d", "e", "f", "g"},
			fn: func(strings []string) error {
				require.True(t, len(strings) == 7)
				return nil
			},
		},
		"when not divisible": {
			size:       3,
			collection: []string{"a", "b", "c", "d", "e", "f", "g"},
			fn: func(strings []string) error {
				require.True(t, len(strings) == 3 || len(strings) == 1)
				return nil
			},
		},
		"when is divisible": {
			size:       3,
			collection: []string{"a", "b", "c", "d", "e", "f"},
			fn: func(strings []string) error {
				require.True(t, len(strings) == 3)
				return nil
			},
		},
		"when errors out": {
			size:       3,
			collection: []string{"a", "b", "c", "d", "e", "f"},
			fn: func(strings []string) error {
				if strings[0] == "d" {
					return fmt.Errorf("error in processing function")
				}
				return nil
			},
			expectedErr: fmt.Errorf("error in processing function"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := util.SplitSliceInBatch(tt.size, tt.collection, tt.fn)
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestDeduplicateList(t *testing.T) {
	type human struct {
		ID   string
		Name string
		Age  int
	}

	tests := []struct {
		name         string
		input        []*human
		expect       []*human
		customAssert func(t *testing.T, output []*human)
	}{
		{
			name:   "when empty",
			input:  []*human{},
			expect: []*human{},
			customAssert: func(t *testing.T, output []*human) {
				assert.Empty(t, output)
			},
		},
		{
			name: "when there no duplicates",
			input: []*human{
				{
					ID:   "1",
					Name: "George",
					Age:  10,
				},
				{
					ID:   "2",
					Name: "Michael",
					Age:  20,
				},
			},
			expect: []*human{
				{
					ID:   "1",
					Name: "George",
					Age:  10,
				},
				{
					ID:   "2",
					Name: "Michael",
					Age:  20,
				},
			},
			customAssert: func(t *testing.T, output []*human) {
				assert.Len(t, output, 2)
			},
		},
		{
			name: "when there are duplicates",
			input: []*human{
				{
					ID:   "1",
					Name: "George",
					Age:  10,
				},
				{
					ID:   "2",
					Name: "Michael",
					Age:  20,
				},
				{
					ID:   "1",
					Name: "Thomas",
					Age:  30,
				},
				{
					ID:   "1",
					Name: "John",
					Age:  40,
				},
			},
			expect: []*human{
				{
					ID:   "1",
					Name: "George",
					Age:  10,
				},
				{
					ID:   "2",
					Name: "Michael",
					Age:  20,
				},
			},
			customAssert: func(t *testing.T, output []*human) {
				assert.Len(t, output, 2)
				assert.Equal(t, "1", output[0].ID)
				assert.Equal(t, "George", output[0].Name)
				assert.Equal(t, 10, output[0].Age)
				assert.Equal(t, "2", output[1].ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.DeduplicateList(tt.input, func(element *human) string {
				return element.ID
			})
			tt.customAssert(t, result)
		})
	}
}
