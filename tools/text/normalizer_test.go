package text

import (
	"testing"
)

func TestLower(t *testing.T) {
	res := Normalize("Hello World")
	if res != "hello world" {
		t.Fail()
	}
}

func TestPunctuation(t *testing.T) {
	res := Normalize("hello, world!")
	if res != "hello  world " {
		t.Fail()
	}
}

func TestAccents(t *testing.T) {
	res := Normalize("äÄáÉíâç")
	if res != "aaaeiac" {
		t.Error(res)
	}
}
