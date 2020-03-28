package indexer

import (
	"bytes"
	"encoding/gob"
	"flash/src/utils/importer"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Index datastructure
type Index struct {
	RootDir           string                  // root directory that the index is used for
	Index             map[string]*postingList // the actual index
	DocumentList      map[int]string          // map of document ids to their filepath
	partitionNumber   int
	memoryConsumption int
	memoryLimit       int
}

// NewIndex creates a new index for the given directory
func NewIndex(root string, memLimit int, partitionNumber int) *Index {
	i := make(map[string]*postingList)
	documentList := make(map[int]string)

	index := Index{
		RootDir:           root,
		Index:             i,
		DocumentList:      documentList,
		partitionNumber:   partitionNumber,
		memoryConsumption: 0,
		memoryLimit:       memLimit,
	}

	return &index
}

func (i *Index) indexDirectory(dir string) {
	docID := 0
	visit := func(path string, info os.FileInfo, err error) error {
		if info.Name()[0:1] == "." {
			if info.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		if info.Mode().IsRegular() {
			i.Add(path, docID)
			docID++
		}

		return nil
	}

	err := filepath.Walk(dir, visit)
	if err != nil {
		fmt.Println(err)
	}

	// Write remaining entries to a partition file
	if i.memoryConsumption > 0 {
		i.createIndexParition()
	}
}

// Add inserts a doument into the index
func (i *Index) Add(path string, docID int) {
	textChannel := importer.GetTextChannel(path)
	i.DocumentList[docID] = path

	offset := 0
	for word := range textChannel {
		word = strings.ToLower(word)

		var pList *postingList

		if val, ok := i.Index[word]; ok {
			pList = val
		} else {
			pList = new(postingList)
			i.Index[word] = pList
		}

		pList.Add(docID, offset)
		i.memoryConsumption++
		offset++

		if i.memoryConsumption > i.memoryLimit {
			i.createIndexParition()
			i.reset(i.partitionNumber + 1)
		}
	}
}

func (i *Index) createIndexParition() {
	partitionPath := fmt.Sprintf("%v/.index/p%d.index", i.RootDir, i.partitionNumber)

	f, err := os.Create(partitionPath)
	if err != nil {
		log.Fatal("Could not create index partition")
	}

	defer f.Close()

	keys := make([]string, len(i.Index))
	count := 0
	for k := range i.Index {
		keys[count] = k
		count++
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	for _, key := range keys {
		buf.WriteString(fmt.Sprintf("%v:%v\n", key, i.Index[key].String()))
	}

	f.Write(buf.Bytes())
}

func (i *Index) reset(partitionNumber int) {
	*i = *NewIndex(i.RootDir, i.memoryLimit, partitionNumber)
}

// Save saves index into the give file
func (i *Index) Save(path string) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	encoder := gob.NewEncoder(f)
	err = encoder.Encode(i)
	if err != nil {
		fmt.Println(err)
	}
}

// Print prints the index
func (i *Index) Print() {
	for t, l := range i.Index {
		fmt.Printf("%v: ", t)
		for g := l.Head; g != nil; g = g.Next {
			for c, p := range g.Postings {
				fmt.Printf("[%v, %v]", p.DocID, p.Offset)

				if c != len(g.Postings)-1 || g.Next != nil {
					fmt.Printf(" -> ")
				}
			}
			fmt.Printf("\n")
		}
	}
}
