package main

import (
	"flash/src/utils/indexer"
	"flash/src/utils/search"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	path, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatal("Could not get filepath")
	}

	start := time.Now()
	// index := indexer.BuildIndex(path)
	index, _ := indexer.LoadIndex(path)

	engine := search.NewEngine(index)

	fmt.Println("Searching...")
	searchStart := time.Now()

	results := engine.Search(os.Args[2], 10)

	fmt.Printf("Found %v results in %v:\n", len(results), time.Since(searchStart))
	for i := range results {
		path, _, _ := index.GetDocInfo(results[i].ID)
		fmt.Printf("%v. %v (%v)\n", i+1, path, results[i].Score)
	}

	fmt.Printf("\nTotal Time: %v\n", time.Since(start))
}
