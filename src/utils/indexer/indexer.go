package indexer

import (
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

	return NewIndex(dir)
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
