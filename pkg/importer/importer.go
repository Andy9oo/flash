package importer

import (
	"context"
	"flash/tools/text"
	"fmt"
	"log"
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

func getText(filepath string, c chan string) {
	defer close(c)

	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println("Couldn't open file:", filepath)
		return
	}

	defer file.Close()

	tikaport := viper.GetString("tikaport")
	client := tika.NewClient(nil, "http://localhost:"+tikaport)
	body, err := client.Parse(context.Background(), file)
	if err != nil {
		log.Fatal(err)
	}

	words := strings.Split(body, " ")
	for _, word := range words {
		c <- text.Normalize(word)
	}
}
