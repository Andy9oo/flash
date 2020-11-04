package importer

import (
	"testing"
)

func TestGetChannel(t *testing.T) {
	channel := GetTextChannel("./testdata/plaintext_test.txt")
	if channel == nil {
		t.Fail()
	}
}

func TestGetText(t *testing.T) {
	channel := GetTextChannel("./testdata/plaintext_test.txt")
	expected := []string{"hello", "world", "plaintext", "test", "txt"}

	i := 0
	for word := range channel {
		if word != expected[i] {
			t.Fail()
		}
		i++
	}
}

func TestMissingFile(t *testing.T) {
	channel := make(chan string)
	err := getText("missing", channel)
	if err == nil {
		t.Fail()
	}
}

func TestUnreadableFile(t *testing.T) {
	channel := make(chan string)
	err := getText("./testdata/unopenable.txt", channel)
	if err == nil {
		t.Fail()
	}
}
