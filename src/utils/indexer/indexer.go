package indexer

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// BuildIndex creates an index file for the given directory
func BuildIndex(dir string) *Index {
	index := NewIndex(dir, 1e6, 0)

	indexDir := fmt.Sprintf("%v/.index", dir)
	os.Mkdir(indexDir, 0755)

	docID := 0
	visit := func(path string, info os.FileInfo, err error) error {
		if info.Name()[0:1] == "." {
			if info.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		if info.Mode().IsRegular() {
			index.Add(path, docID)
			docID++
		}

		return nil
	}

	err := filepath.Walk(dir, visit)

	// Write remaining entries to a partition file
	if index.memoryConsumption > 0 {
		index.createIndexParition()
	}

	if err != nil {
		fmt.Println(err)
	}

	return index
}

// LoadIndex reads an index in from a file
func LoadIndex(path string) *Index {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	var i Index

	decoder := gob.NewDecoder(f)
	err = decoder.Decode(&i)
	if err != nil {
		fmt.Println(err)
	}

	return &i
}
