package index

import (
	"errors"
	"flash/pkg/importer"
	"flash/pkg/index/doclist"
	"flash/pkg/index/postinglist"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/theckman/yacspin"
)

const (
	documentListLimit = 1 << 20
	partitionLimit    = 1 << 24
	dictionaryLimit   = 1 << 20
	blockSize         = 1 << 10
	chunkSize         = 1 << 10
)

// Index datastructure
type Index struct {
	dir        string
	dict       *dictionary
	docs       *doclist.DocList
	partitions []*partition
	numParts   int
}

// Info contains information about an index
type Info struct {
	NumDocs     uint32
	TotalLength int
}

// Build builds a new index for the given directory
func Build(indexpath, root string) *Index {
	i := Index{
		dir:      indexpath,
		docs:     doclist.NewList(indexpath, documentListLimit),
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
	i.docs.CalculateOffsets(blockSize)
	docDone := time.Now()

	spinner.Stop()

	fmt.Printf("Indexing: %v\nMerging: %v\nDictionary: %v\nDoclist: %v\n\nTotal: %v\n",
		indexDone.Sub(start), mergeDone.Sub(indexDone), dictDone.Sub(mergeDone), docDone.Sub(dictDone), time.Since(start))

	return &i
}

// Load opens the index at the indexpath
func Load(indexpath string) (i *Index, err error) {
	i = &Index{dir: indexpath}
	_, err = os.Stat(i.getPostingsPath())
	if err != nil {
		return nil, errors.New("Could not find index, has one been built?")
	}

	i.dict = loadDictionary(indexpath, dictionaryLimit)
	i.docs = doclist.Load(indexpath, documentListLimit)

	return i, nil
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
				go p.dump()
				p = i.addPartition()
			}

			p.add(term, i.docs.GetID(), offset)
			offset++
		}

		i.docs.Add(path, offset)
	}
}

// GetPostingReader returns a posting reader for a term
func (i *Index) GetPostingReader(term string) (*postinglist.Reader, bool) {
	if buf, ok := i.dict.getPostingBuffer(term); ok {
		return postinglist.NewReader(buf), true
	}

	return nil, false
}

// GetInfo returns information about the index
func (i *Index) GetInfo() *Info {
	info := Info{
		NumDocs:     i.docs.NumDocs(),
		TotalLength: i.docs.TotalLength(),
	}

	return &info
}

// GetDocInfo returns information about the given document
func (i *Index) GetDocInfo(id uint32) (path string, length uint32, ok bool) {
	if doc, ok := i.docs.Fetch(id); ok {
		return doc.Path(), doc.Length(), true
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
	i.docs.Dump()
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
