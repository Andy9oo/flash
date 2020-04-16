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

const (
	documentListLimit = 1024
	postingsLimit     = 4096
	dictionaryLimit   = 4096
	chunkSize         = 1000
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

	spinner.Message("Indexing Directory")
	i.index(root)

	spinner.Message("Merging Partitions")
	i.mergeParitions()

	spinner.Message("Loading Dictionary")
	i.dict = loadDictionary(i.dir, dictionaryLimit)

	spinner.Stop()
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

func (i *Index) First() {

}

func (i *Index) Last() {

}

func (i *Index) Next() {

}

func (i *Index) Prev() {

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
			i.add(path)
		}

		return nil
	}

	err := filepath.Walk(dir, visit)
	if err != nil {
		fmt.Println(err)
	}

	i.clearMemory()
}

func (i *Index) add(file string) {
	textChannel := importer.GetTextChannel(file)
	i.docs.add(file)

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
