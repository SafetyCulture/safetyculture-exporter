package util

// GenericCollectionMapper will map from []T to []R by providing a transformation function
func GenericCollectionMapper[T, R any](itemsSource []T, transformFn func(T) R) []R {
	itemsDestination := make([]R, len(itemsSource))
	for i, item := range itemsSource {
		itemsDestination[i] = transformFn(item)
	}
	return itemsDestination
}
