package utils

func RemoveDuplicates[T comparable](slice []T) []T {
	if len(slice) == 0 {
		return slice
	}

	seen := make(map[T]struct{}, len(slice))
	result := make([]T, 0, len(slice))

	for _, item := range slice {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}
