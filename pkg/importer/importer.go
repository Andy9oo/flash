package importer

import (
	"context"
	"flash/tools/text"
	"os"
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

func getText(path string, c chan string) error {
	defer close(c)

	stat, err := os.Stat(path)
	if err != nil {
		return err
	}

	name := stat.Name()
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer file.Close()

	tikaport := viper.GetString("tikaport")
	if tikaport == "" {
		tikaport = "9998"
	}

	client := tika.NewClient(nil, "http://localhost:"+tikaport)
	body, _ := client.Parse(context.Background(), file)

	words := strings.Fields(text.Normalize(body + " " + name))
	for _, word := range words {
		c <- word
	}

	return nil
}
