package indexer

import (
	"encoding/gob"
	"flash/src/utils/importer"
	"fmt"
	"log"
	"os"
	"strings"
)

// Index datastructure
type Index struct {
	RootDir      string                  // root directory that the index is used for
	FileName     string                  // filename used to store the index
	Index        map[string]*postingList // the actual index
	DocumentList map[int]string          // map of document ids to their filepath
}

// NewIndex creates a new index for the given directory
func NewIndex(root string) *Index {
	i := make(map[string]*postingList)
	documentList := make(map[int]string)

	index := Index{
		RootDir:      root,
		Index:        i,
		DocumentList: documentList,
	}

	return &index
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
		offset++
	}
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
		for p := l.Head; p != nil; p = p.Next {

			fmt.Printf("[%v, %v]", p.DocID, p.Offset)

			if p.Next != nil {
				fmt.Printf(" -> ")
			}
		}
		fmt.Printf("\n")
	}
}
