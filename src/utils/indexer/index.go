package indexer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	partitionSizeLimit = uint32(1e5)
	documentListLimit  = uint32(1e4)
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
		dict: newDictionary(dir, 4096),
		docs: newDocList(dir, 100),
	}

	i.mkdir()
	i.index(root)
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
	i.mergePartitions()
}

func (i *Index) mergePartitions() {
	path := fmt.Sprintf("%v/index.postings", i.dir)

	f, err := os.Create(path)
	if err != nil {
		log.Fatal("Could not create index partition")
	}
	defer f.Close()

	readers := i.getPartitionReaders()

	reader := readers[0]
	var currentTerm string
	var prevTerm string

	for reader != nil {
		reader = nil
		for i := 0; i < len(readers); i++ {
			if !readers[i].done {
				if reader == nil || readers[i].compare(currentTerm) == -1 {
					reader = readers[i]
					currentTerm = readers[i].currentTerm
				}
			}
		}

		if reader != nil {
			buf := new(bytes.Buffer)
			if currentTerm != prevTerm {
				binary.Write(buf, binary.LittleEndian, []byte("\n"))
				binary.Write(buf, binary.LittleEndian, []byte(currentTerm))
				binary.Write(buf, binary.LittleEndian, []byte("\n"))
			}
			binary.Write(buf, binary.LittleEndian, reader.getPostings())
			buf.WriteTo(f)

			reader.advanceCurrentTerm()
		}
	}
	i.deletePartitionFiles()
}

func (i *Index) getPartitionReaders() []*partitionReader {
	var readers []*partitionReader

	for _, file := range i.getPartitionFiles() {
		path := fmt.Sprintf("%v/%v", i.dir, file)
		readers = append(readers, newPartitionReader(path))
	}

	return readers
}

func (i *Index) getPartitionFiles() []string {
	var files []string

	dir, err := os.Open(i.dir)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()

	all, err := dir.Readdirnames(-1)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range all {
		if filepath.Ext(file) == ".part" {
			files = append(files, file)
		}
	}

	return files
}

func (i *Index) deletePartitionFiles() {
	files := i.getPartitionFiles()
	for _, file := range files {
		path := fmt.Sprintf("%v/%v", i.dir, file)
		os.Remove(path)
	}
}

func (i *Index) mkdir() {
	err := os.RemoveAll(i.dir)
	err = os.Mkdir(i.dir, 0755)
	if err != nil {
		log.Fatal("Could not create index directory")
	}
}
