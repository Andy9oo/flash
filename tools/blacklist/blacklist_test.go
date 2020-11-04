package blacklist

import (
	"testing"
)

func TestAdd(t *testing.T) {
	b := &Blacklist{}
	b.Add("test")
	patterns := b.GetPatterns()
	if len(patterns) != 1 || patterns[0] != "test" {
		t.Fail()
	}
}

func TestInvalidRegex(t *testing.T) {
	b := &Blacklist{}
	err := b.Add("[")
	if err == nil {
		t.Fail()
	}
}

func TestGetPatterns(t *testing.T) {
	patterns := []string{"pattern1", "pattern2", "pattern3"}
	b := &Blacklist{}
	b.Add(patterns...)
	for i, pattern := range b.GetPatterns() {
		if pattern != patterns[i] {
			t.Fail()
		}
	}
}

func TestContainsPostive(t *testing.T) {
	b := &Blacklist{}
	b.Add("test")
	if b.Contains("test") == false {
		t.Fail()
	}
}

func TestContainsNegative(t *testing.T) {
	b := &Blacklist{}
	b.Add("test")
	if b.Contains("failure") == true {
		t.Fail()
	}
}

func TestRemove(t *testing.T) {
	patterns := []string{"pattern1", "pattern2", "pattern3"}
	b := &Blacklist{}
	b.Add(patterns...)
	b.Remove("pattern2")
	if len(b.GetPatterns()) != 2 {
		t.Fail()
	}
}

func TestReset(t *testing.T) {
	b := &Blacklist{}
	b.Add("test")
	b.Reset()
	if len(b.GetPatterns()) != 0 {
		t.Fail()
	}
}
