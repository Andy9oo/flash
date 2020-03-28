package main

import (
	"flash/src/utils/indexer"
	"log"
	"os"
	"path/filepath"
)

func main() {
	path, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatal("Could not get filepath")
	}

	indexer.BuildIndex(path)
}
