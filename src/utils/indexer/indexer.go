package indexer

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
)

// BuildIndex creates an index file for the given directory
func BuildIndex(dir string) *Index {
	err := createIndexDir(dir)
	if err != nil {
		log.Fatal("Could not create index directory")
	}

	index := NewIndex(dir, 1e6, 0)
	index.indexDirectory(dir)

	return index
}

func createIndexDir(root string) error {
	indexDir := fmt.Sprintf("%v/.index", root)

	err := os.RemoveAll(indexDir)
	if err != nil {
		return err
	}

	err = os.Mkdir(indexDir, 0755)
	return err
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
