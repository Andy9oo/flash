package main

import (
	"flash/src/utils/indexer"
	"flash/src/utils/search"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"
)

func main() {
	path, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatal("Could not get filepath")
	}

	cpu, err := os.Create(path + "/cpu.profile")
	if err != nil {
		log.Fatal(err)
	}

	// index := indexer.BuildIndex(path)
	index, _ := indexer.LoadIndex(path)

	engine := search.NewEngine(index)

	pprof.StartCPUProfile(cpu)

	fmt.Printf("\nSearching...\n")
	start := time.Now()

	results := engine.Search("is", 10)
	fmt.Printf("Found %v results in %v:\n", len(results), time.Since(start))
	for i := range results {
		path, _ := index.GetPath(results[i].ID)
		fmt.Printf("%v. %v (%v)\n", i+1, path, results[i].Score)
	}

	pprof.StopCPUProfile()
}
