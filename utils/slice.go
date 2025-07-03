package utils

func RemoveDuplicates(slice []string) []string {
	if len(slice) == 0 {
		return slice
	}

	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
