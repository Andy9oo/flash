package importer

import (
	"flash/tools/tika"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func setupServer() *tika.Server {
	tp, _ := filepath.Abs("../../tools/tika/tika.jar")

	viper.Set("tikapath", tp)
	viper.Set("tikaport", "9999")

	server := tika.GetServer()
	fmt.Println(server)
	server.StartServer()
	return server
}

func TestGetChannel(t *testing.T) {
	server := setupServer()
	defer server.StopServer()

	channel := GetTextChannel("./testdata/plaintext_test.txt")
	if channel == nil {
		t.Fail()
	}
}

func TestGetText(t *testing.T) {
	server := setupServer()
	defer server.StopServer()

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
	server := setupServer()
	defer server.StopServer()

	channel := make(chan string)
	err := getText("missing", channel)
	if err == nil {
		t.Fail()
	}
}
