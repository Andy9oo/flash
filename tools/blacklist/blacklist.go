package blacklist

import (
	"regexp"
)

// Blacklist struct
type Blacklist struct {
	patterns []*regexp.Regexp
}

// Contains returns true if the path is in the blacklist
func (b *Blacklist) Contains(s string) bool {
	for _, reg := range b.patterns {
		if reg.MatchString(s) {
			return true
		}
	}
	return false
}

// Add adds the given pattern to the blacklist
func (b *Blacklist) Add(patterns ...string) error {
	for _, p := range patterns {
		reg, err := regexp.Compile(p)
		if err != nil {
			return err
		}
		b.patterns = append(b.patterns, reg)
	}
	return nil
}

// Remove removes the given pattern from the
func (b *Blacklist) Remove(pattern string) {
	for i, reg := range b.patterns {
		if reg.String() == pattern {
			b.patterns[i] = b.patterns[len(b.patterns)-1]
			b.patterns = b.patterns[:len(b.patterns)-1]
			return
		}
	}
}

// GetPatterns returns all paterns in the blacklist
func (b *Blacklist) GetPatterns() []string {
	patterns := make([]string, len(b.patterns))
	for i, pattern := range b.patterns {
		patterns[i] = pattern.String()
	}
	return patterns
}

// Reset empties the blacklist
func (b *Blacklist) Reset() {
	b.patterns = []*regexp.Regexp{}
}
