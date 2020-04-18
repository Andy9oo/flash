package indexer

import (
	"flash/src/utils/importer"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/theckman/yacspin"
)

// Byte sizes
const (
	_      = iota
	KB int = 1 << (10 * iota)
	MB
	GB
)

const (
	documentListLimit = 1 << 10
	partitionLimit    = 250 * MB
	dictionaryLimit   = 1 << 20
	chunkSize         = 1 << 10
)

// Index datastructure
type Index struct {
	dir        string
	dict       *dictionary
	docs       *doclist
	partitions []*partition
	numParts   int
}

// BuildIndex builds a new index for the given directory
func BuildIndex(root string) *Index {
	dir := fmt.Sprintf("%v/.index", root)

	i := Index{
		dir:      dir,
		docs:     newDocList(dir, documentListLimit),
		numParts: -1,
	}

	i.createDir()
	i.addPartition()

	cfg := yacspin.Config{
		CharSet:         yacspin.CharSets[59],
		Frequency:       100 * time.Millisecond,
		Suffix:          " Building Index",
		SuffixAutoColon: true,
		StopCharacter:   "✓",
		StopMessage:     "Done!",
		Colors:          []string{"fgYellow"},
		StopColors:      []string{"fgGreen"},
		ColorAll:        true,
	}

	spinner, _ := yacspin.New(cfg)
	spinner.Start()

	start := time.Now()
	spinner.Message("Indexing Directory")
	i.index(root)
	indexDone := time.Now()

	spinner.Message("Merging Partitions")
	i.mergeParitions()
	mergeDone := time.Now()

	spinner.Message("Loading Dictionary")
	i.dict = loadDictionary(i.dir, dictionaryLimit)
	dictDone := time.Now()
	spinner.Stop()

	indexTime := indexDone.Sub(start)
	mergeTime := mergeDone.Sub(indexDone)
	dictTime := dictDone.Sub(mergeDone)

	fmt.Printf("Indexing: %v\nMerging: %v\nDictionary: %v\n", indexTime, mergeTime, dictTime)
	return &i
}

// LoadIndex opens the index for the given root file
func LoadIndex(root string) (index *Index, ok bool) {
	dir := fmt.Sprintf("%v/.index", root)

	index = &Index{
		dir:  dir,
		docs: newDocList(dir, documentListLimit),
	}

	postingsFile := index.getPostingsPath()
	_, err := os.Stat(postingsFile)
	if err != nil {
		fmt.Println("No index found")
		return nil, false
	}

	index.dict = loadDictionary(dir, dictionaryLimit)
	return index, true
}

// Add adds the given file or directory to the index
func (i *Index) Add(path string) {
	stat, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	if stat.IsDir() {
		i.index(path)
	} else {
		textChannel := importer.GetTextChannel(path)
		i.docs.add(path)

		var offset uint32
		for term := range textChannel {
			p := i.partitions[len(i.partitions)-1]
			if p.full() {
				p.dump()
				p = i.addPartition()
			}

			term = strings.ToLower(term)
			p.add(term, i.docs.numDocs, offset)
			offset++
		}
	}
}

// First returns the first doccument and offset where a term occurs
func (i *Index) First(term string) (d uint32, o uint32, ok bool) {
	pl, ok := i.dict.getPostings(term)
	if !ok {
		return 0, 0, false
	}

	docs := pl.getDocs()
	firstID := docs[0]
	first := pl.postings[firstID]

	return firstID, first.offsets[0], true
}

// Last returns the last occurence of a term
func (i *Index) Last(term string) (d uint32, o uint32, ok bool) {
	pl, ok := i.dict.getPostings(term)
	if !ok {
		return 0, 0, false
	}

	docs := pl.getDocs()
	lastID := docs[len(docs)-1]
	last := pl.postings[lastID]

	return lastID, last.offsets[last.frequency-1], true
}

// Next returns the next occurence of a term after a given offset
func (i *Index) Next(term string, docid uint32, offset uint32) (d uint32, o uint32, ok bool) {
	pl, ok := i.dict.getPostings(term)
	if !ok {
		return 0, 0, false
	}

	docs := pl.getDocs()
	for i := 0; i < len(docs); i++ {
		if docs[i] < docid {
			continue
		}

		p := pl.postings[docs[i]]
		for i := 0; i < len(p.offsets); i++ {
			if p.offsets[i] > offset || p.docID > docid {
				return p.docID, p.offsets[i], true
			}
		}
	}

	return 0, 0, false
}

// Prev returns the previous occurence of a term before a given offset
func (i *Index) Prev(term string, docid uint32, offset uint32) (d uint32, o uint32, ok bool) {
	pl, ok := i.dict.getPostings(term)
	if !ok {
		return 0, 0, false
	}

	docs := pl.getDocs()
	for i := len(docs) - 1; i >= 0; i-- {
		if docs[i] > docid {
			continue
		}

		p := pl.postings[docs[i]]
		for i := len(p.offsets) - 1; i >= 0; i-- {
			if p.offsets[i] < offset || p.docID < docid {
				return p.docID, p.offsets[i], true
			}
		}
	}

	return 0, 0, false
}

// NextDoc returns the next document which contains a term
func (i *Index) NextDoc(term string, docid uint32) (d uint32, ok bool) {
	pl, ok := i.dict.getPostings(term)
	if !ok {
		return 0, false
	}

	docs := pl.getDocs()
	for i := 0; i < len(docs); i++ {
		if docs[i] > docid {
			return docs[i], true
		}
	}

	return 0, false
}

// PrevDoc returns the previous document which contains a term
func (i *Index) PrevDoc(term string, docid uint32) (d uint32, ok bool) {
	pl, ok := i.dict.getPostings(term)
	if !ok {
		return 0, false
	}

	docs := pl.getDocs()
	for i := len(docs) - 1; i >= 0; i-- {
		if docs[i] < docid {
			return docs[i], true
		}
	}

	return 0, false
}

func (i *Index) index(dir string) {
	visit := func(path string, info os.FileInfo, err error) error {
		if info.Name()[0:1] == "." {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.Mode().IsRegular() {
			i.Add(path)
		}

		return nil
	}

	err := filepath.Walk(dir, visit)
	if err != nil {
		fmt.Println(err)
	}

	i.clearMemory()
}

func (i *Index) addPartition() *partition {
	i.numParts++
	p := newPartition(i.dir, i.numParts)
	i.partitions = append(i.partitions, p)
	return p
}

func (i *Index) mergeParitions() {
	partitions := i.partitions

	for len(partitions) != 1 {
		numParts := len(partitions)
		numChunks := int(math.Ceil(float64(numParts) / chunkSize))
		chunks := make([][]*partition, 0, numChunks)

		for c := 0; c < numParts; c += chunkSize {
			if c+chunkSize >= numParts {
				chunks = append(chunks, partitions[c:])
			} else {
				chunks = append(chunks, partitions[c:c+chunkSize])
			}
		}

		partitions = make([]*partition, len(chunks))
		for c := range chunks {
			i.numParts++
			partitions[c] = merge(i.dir, i.numParts, chunks[c])
		}
	}

	i.partitions = make([]*partition, 0)
	i.numParts = 0

	os.Rename(partitions[0].getPath(), i.getPostingsPath())
}

func (i *Index) clearMemory() {
	for _, p := range i.partitions {
		p.dump()
	}
	i.docs.dump()
}

func (i *Index) createDir() {
	err := os.RemoveAll(i.dir)
	err = os.Mkdir(i.dir, 0755)
	if err != nil {
		log.Fatal("Could not create index directory")
	}
}

func (i *Index) getPostingsPath() string {
	return fmt.Sprintf("%v/index.postings", i.dir)
}
