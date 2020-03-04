package main

import (
	"fmt"
	"importer"
	"log"
	"os"
	"path/filepath"
)

func main() {
	file, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatal("Could not get absolute path for", os.Args[1])
	}

	channel := importer.GetTextChannel(file)

	for line := range channel {
		fmt.Println(line)
	}
}
