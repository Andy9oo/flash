package indexer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	documentListLimit = uint32(1e4)
)

// Index datastructure
type Index struct {
	dir  string
	dict *dictionary
	docs *doclist
}

// BuildIndex builds a new index for the given directory
func BuildIndex(root string) *Index {
	dir := fmt.Sprintf("%v/.index", root)

	i := Index{
		dir:  dir,
		docs: newDocList(dir, 100),
	}

	i.mkdir()

	fmt.Println("Building index...")
	i.index(root)

	fmt.Println("Loading Dictionary...")
	i.dict = loadDictionary(i.dir, 4096)

	fmt.Println("Done!")
	return &i
}

func (i *Index) index(dir string) {
	partition := newPartition(dir, 0)
	var docID uint32

	visit := func(path string, info os.FileInfo, err error) error {
		if info.Name()[0:1] == "." {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.Mode().IsRegular() {
			partition.add(path, docID)
			i.docs.add(docID, path)
			docID++
		}

		return nil
	}

	err := filepath.Walk(dir, visit)
	if err != nil {
		fmt.Println(err)
	}

	partition.dump()
	i.docs.dump()

	pm := newPartitionMerger(i.dir)
	pm.mergePartitions()
}

func (i *Index) mkdir() {
	err := os.RemoveAll(i.dir)
	err = os.Mkdir(i.dir, 0755)
	if err != nil {
		log.Fatal("Could not create index directory")
	}
}
