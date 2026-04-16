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
