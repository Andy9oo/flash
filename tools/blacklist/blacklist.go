package blacklist

import (
	"regexp"
)

var blacklist []*regexp.Regexp

// Contains returns true if the path is in the blacklist
func Contains(s string) bool {
	for _, reg := range blacklist {
		if reg.MatchString(s) {
			return true
		}
	}
	return false
}

// Add adds the given pattern to the blacklist
func Add(patterns ...string) error {
	for _, p := range patterns {
		reg, err := regexp.Compile(p)
		if err != nil {
			return err
		}
		blacklist = append(blacklist, reg)
	}
	return nil
}

// Remove removes the given pattern from the
func Remove(pattern string) {
	for i, reg := range blacklist {
		if reg.String() == pattern {
			blacklist[i] = blacklist[len(blacklist)-1]
			blacklist = blacklist[:len(blacklist)-1]
			return
		}
	}
}

// GetPatterns returns all paterns in the blacklist
func GetPatterns() []string {
	patterns := make([]string, len(blacklist))
	for i, pattern := range blacklist {
		patterns[i] = pattern.String()
	}

	return patterns
}

// Reset empties the blacklist
func Reset() {
	blacklist = []*regexp.Regexp{}
}
