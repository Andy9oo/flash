package index

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flash/pkg/importer"
	"flash/pkg/index/doclist"
	"flash/pkg/index/postinglist"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
)

const (
	documentListLimit = 1 << 20
	partitionLimit    = 1 << 18
	dictionaryLimit   = 1 << 20
	blockSize         = 1 << 10
	chunkSize         = 1 << 10
)

// Index datastructure
type Index struct {
	dir        string
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
	i.index(root)
	i.dumpInfo()

	return &i
}

// Load opens the index at the indexpath
func Load(indexpath string) *Index {
	i := &Index{dir: indexpath}
	i.loadInfo()
	i.docs = doclist.Load(indexpath, documentListLimit)
	return i
}

// Add adds the given file or directory to the index
func (i *Index) Add(path string) {
	stat, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	if !stat.IsDir() {
		textChannel := importer.GetTextChannel(path)
		var offset uint32

		for term := range textChannel {
			p := i.partitions[len(i.partitions)-1]
			if p.full() {
				p.dump()
				p = i.addPartition()
			}

			p.add(term, i.docs.GetID(), offset)
			offset++
		}

		i.docs.Add(path, offset)
	} else {
		i.index(path)
	}
}

// GetPostingReaders returns a list of posting readers for the given term
func (i *Index) GetPostingReaders(term string) []*postinglist.Reader {
	var readers []*postinglist.Reader
	for _, p := range i.partitions {
		if r, ok := p.GetPostingReader(term); ok {
			readers = append(readers, r)
		}
	}
	return readers
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

func (i *Index) loadInfo() {
	path := fmt.Sprintf("%v/index.info", i.dir)
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	reader := bufio.NewReader(f)
	buf := make([]byte, 4)

	for {
		n, err := reader.Read(buf)
		if n == 0 || err != nil {
			break
		}

		gen := binary.LittleEndian.Uint32(buf)

		i.partitions = append(i.partitions, loadPartition(i.dir, int(gen)))
	}
}

func (i *Index) dumpInfo() {
	path := fmt.Sprintf("%v/index.info", i.dir)
	f, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	buf := new(bytes.Buffer)
	for p := range i.partitions {
		binary.Write(buf, binary.LittleEndian, uint32(i.partitions[p].generation))
	}

	buf.WriteTo(f)
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
	p := newPartition(i.dir, i.numParts) // TODO change to zero
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
