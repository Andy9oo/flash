package indexer

import (
	"flash/src/utils/importer"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
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
	documentListLimit = 1 << 20
	partitionLimit    = 250 * MB
	dictionaryLimit   = 1 << 20
	blockSize         = 1 << 10
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

// IndexInfo contains information about an index
type IndexInfo struct {
	NumDocs     uint32
	TotalLength int
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

	spinner.Message("Loading Documents")
	i.docs.calculateOffsets(blockSize)
	docsDone := time.Now()

	spinner.Stop()

	indexTime := indexDone.Sub(start)
	mergeTime := mergeDone.Sub(indexDone)
	dictTime := dictDone.Sub(mergeDone)
	docTime := docsDone.Sub(dictDone)

	fmt.Printf("Indexing: %v\nMerging: %v\nDictionary: %v\nDocs: %v\n", indexTime, mergeTime, dictTime, docTime)
	return &i
}

// LoadIndex opens the index for the given root file
func LoadIndex(root string) (i *Index, ok bool) {
	dir := fmt.Sprintf("%v/.index", root)

	i = &Index{dir: dir}
	_, err := os.Stat(i.getPostingsPath())
	if err != nil {
		fmt.Println("No index found")
		return nil, false
	}

	i.dict = loadDictionary(dir, dictionaryLimit)
	i.docs = loadDocList(dir, documentListLimit, blockSize)

	return i, true
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
		var offset uint32

		for term := range textChannel {
			p := i.partitions[len(i.partitions)-1]
			if p.full() {
				p.dump()
				p = i.addPartition()
			}

			p.add(term, i.docs.totalDocs, offset)
			offset++
		}

		i.docs.add(path, offset)
	}
}

// GetPostingReader returns a posting reader for a term
func (i *Index) GetPostingReader(term string) (*PostingReader, bool) {
	if buf, ok := i.dict.getPostingBuffer(term); ok {
		return NewPostingReader(buf), true
	}

	return nil, false
}

// GetInfo returns information about the index
func (i *Index) GetInfo() *IndexInfo {
	info := IndexInfo{
		NumDocs:     i.docs.totalDocs,
		TotalLength: i.docs.totalLength,
	}

	return &info
}

// GetDocInfo returns information about the given document
func (i *Index) GetDocInfo(id uint32) (path string, length uint32, ok bool) {
	if doc, ok := i.docs.fetchDoc(id); ok {
		return doc.path, doc.length, true
	}
	return "", 0, false
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
	i.docs.dumpFiles()
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
