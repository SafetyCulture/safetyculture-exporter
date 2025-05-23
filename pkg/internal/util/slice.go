package util

import (
	"fmt"
)

// SplitSliceInBatch splits a slice into batches and calls the callback function for each batch.
func SplitSliceInBatch[T any](size int, collection []T, fn func(batch []T) error) error {
	if size == 0 {
		return fmt.Errorf("batch size cannot be 0")
	}

	for i := 0; i < len(collection); i += size {
		j := i + size
		if j > len(collection) {
			j = len(collection)
		}

		if err := fn(collection[i:j]); err != nil {
			return err
		}
	}
	return nil
}

// DeduplicateList - based on a collection type T and a function that returns the unique KEY.
func DeduplicateList[T any](elements []*T, pkFun func(element *T) string) []*T {
	var dMap = map[string]*T{}
	var filteredValues []*T

	if len(elements) == 0 {
		return filteredValues
	}

	// use a map to de-dupe
	for _, row := range elements {
		mapPk := pkFun(row)
		if _, ok := dMap[mapPk]; !ok {
			// will skip duplications
			dMap[mapPk] = row
			filteredValues = append(filteredValues, row)
		}
	}

	return filteredValues
}
