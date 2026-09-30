package lib

import "strings"

func Ptr[T any](v T) *T {
	return &v
}

func toPascalCase(s string) string {
	words := strings.Split(s, "_")
	for i, w := range words {
		words[i] = strings.Title(w)
	}
	return strings.Join(words, "")
}

func toCamelCase(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}
