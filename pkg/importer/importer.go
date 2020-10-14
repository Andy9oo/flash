package importer

import (
	"context"
	"flash/tools/text"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-tika/tika"
	"github.com/spf13/viper"
)

// GetTextChannel Returns a channel from which the text of a file is exported
func GetTextChannel(filepath string) chan string {
	channel := make(chan string, 100)
	go getText(filepath, channel)

	return channel
}

func getText(path string, c chan string) {
	defer close(c)

	stat, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	name := stat.Name()
	name = name[0 : len(name)-len(filepath.Ext(name))]

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Couldn't open file:", path)
		return
	}

	defer file.Close()

	tikaport := viper.GetString("tikaport")
	client := tika.NewClient(nil, "http://localhost:"+tikaport)
	body, err := client.Parse(context.Background(), file)
	if err != nil {
		fmt.Println(err)
		return
	}

	words := strings.Fields(body + " " + name)
	for _, word := range words {
		c <- text.Normalize(word)
	}
}
