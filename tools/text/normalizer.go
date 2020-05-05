package text

import "strings"

// Normalize takes the input text and returns a normalized version of it
func Normalize(input string) string {
	return strings.ToLower(input)
}
