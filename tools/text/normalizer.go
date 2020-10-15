package text

import (
	"log"
	"regexp"
	"strings"
)

// Normalize takes the input text and returns a normalized version of it
func Normalize(input string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9\\s]+")
	if err != nil {
		log.Fatal(err)
	}
	chars := reg.ReplaceAllString(input, "")
	return strings.ToLower(chars)
}
