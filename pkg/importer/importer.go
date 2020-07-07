package importer

import (
	"bufio"
	"flash/tools/text"
	"fmt"
	"os"
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

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		c <- text.Normalize(scanner.Text())
	}
}