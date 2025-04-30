package utils

import (
	"fmt"
	"regexp"
	"strings"
)

func ToSnakeCase(s string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchFirstCap.ReplaceAllString(s, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func ExtractNsPrefix(namespace string) (string, error) {
	pattern := `^([a-zA-Z]+)\d+$`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(namespace)

	if len(match) != 2 {
		return "", fmt.Errorf("failed to extract index from namespace %s", namespace)
	}

	return match[1], nil
}
