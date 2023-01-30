package util

// DeduplicateList a list of T type and maintains the latest value
func DeduplicateList[T any](pkFun func(element *T) string, elements []*T) []*T {
	var dMap = map[string]*T{}
	var filteredValues []*T

	if len(elements) == 0 {
		return filteredValues
	}

	for _, row := range elements {
		mapPk := pkFun(row)
		dMap[mapPk] = row
	}

	for _, row := range dMap {
		filteredValues = append(filteredValues, row)
	}
	return filteredValues
}
