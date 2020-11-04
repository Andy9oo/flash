package text

import (
	"regexp"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var reg = regexp.MustCompile("\\p{P}+")

// Normalize takes the input text and returns a normalized version of it
func Normalize(input string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC, cases.Lower(language.English))
	normalized, _, _ := transform.String(t, input)
	normalized = reg.ReplaceAllString(normalized, " ")
	return normalized
}
