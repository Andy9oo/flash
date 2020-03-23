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
	index := NewIndex(dir)

	docID := 0
	visit := func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsRegular() {
			index.Add(path, docID)
			docID++
		}

		return nil
	}

	err := filepath.Walk(dir, visit)
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
