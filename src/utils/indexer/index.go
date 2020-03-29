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
	root         string            // root directory that the index is used for
	dictionary   map[string]uint64 // map from terms to file offsets
	documentList map[uint32]string // map of document ids to their filepath
}

// NewIndex creates a new index for the given directory
func NewIndex(root string) *Index {
	index := Index{
		root:         root,
		dictionary:   make(map[string]uint64),
		documentList: make(map[uint32]string),
	}

	index.addDir(root)
	return &index
}

func (i *Index) addDir(dir string) {
	partition := newIndexPartition(dir, 0)
	var docID uint32

	visit := func(path string, info os.FileInfo, err error) error {
		if info.Name()[0:1] == "." {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.Mode().IsRegular() {
			partition.addDoc(path, docID)
			i.addToDocList(docID, path)
			docID++
		}

		return nil
	}

	err := filepath.Walk(dir, visit)
	if err != nil {
		fmt.Println(err)
	}

	// Write remaining entries to a partition file
	if partition.memoryUsage > 0 {
		partition.writeToFile()
	}

	if len(i.documentList) > 0 {
		i.writeDocList()
	}
}

func (i *Index) addToDocList(docID uint32, path string) {
	i.documentList[docID] = path

	if uint32(len(i.documentList)) > documentListLimit {
		i.writeDocList()
	}
}

func (i *Index) writeDocList() {
	path := fmt.Sprintf("%v/.index/documents.doclist", i.root)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Could not open document list file")
	}

	defer f.Close()

	buf := new(bytes.Buffer)
	for id, path := range i.documentList {
		binary.Write(buf, binary.LittleEndian, id)
		binary.Write(buf, binary.LittleEndian, []byte("\n"))
		binary.Write(buf, binary.LittleEndian, []byte(path))
		binary.Write(buf, binary.LittleEndian, []byte("\n"))
	}

	buf.WriteTo(f)
	i.documentList = make(map[uint32]string)
}
