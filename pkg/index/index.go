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
	"os"
	"path/filepath"
	"sort"
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
	curentPart *partition
}

// Info contains information about an index
type Info struct {
	NumDocs     uint32
	TotalLength int
}

// NewIndex creates a new index
func NewIndex(indexpath string) *Index {
	i := Index{
		dir:  indexpath,
		docs: doclist.NewList(indexpath, documentListLimit),
	}

	i.createDir()
	i.addPartition()

	return &i
}

// Load opens the index at the indexpath
func Load(indexpath string) *Index {
	i := &Index{dir: indexpath}
	err := i.loadInfo()
	if err != nil {
		i = NewIndex(indexpath)
	} else {
		i.docs = doclist.Load(indexpath, documentListLimit)
	}
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
			if i.curentPart.full() {
				i.mergeParitions()
			}

			i.curentPart.add(term, &postingEntry{i.docs.GetID(), offset})
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
	for _, p := range append(i.partitions, i.curentPart) {
		if buf, ok := p.GetBuffer(term); ok {
			readers = append(readers, postinglist.NewReader(buf))
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

func (i *Index) loadInfo() error {
	path := fmt.Sprintf("%v/index.info", i.dir)
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(f)
	buf := make([]byte, 4)

	for {
		n, err := reader.Read(buf)
		if n == 0 || err != nil {
			break
		}

		gen := int(binary.LittleEndian.Uint32(buf))
		part := loadPartition(i.dir, "postings", gen, partitionLimit, newIndexPartition())

		// Load in-memory index
		if gen == 0 {
			i.curentPart = part
		} else {
			i.partitions = append(i.partitions, part)
		}
	}
	return nil
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
	binary.Write(buf, binary.LittleEndian, uint32(i.curentPart.generation))
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
}

func (i *Index) addPartition() {
	if i.curentPart != nil {
		i.partitions = append(i.partitions, i.curentPart)
	}
	i.curentPart = newPartition(i.dir, "postings", 0, partitionLimit, newIndexPartition())
}

func (i *Index) mergeParitions() {
	// Dump current partition into temp file
	current := i.curentPart
	current.dump()

	// Sort partitions in order of generation
	sort.Slice(i.partitions, func(p1, p2 int) bool {
		return i.partitions[p1].generation < i.partitions[p2].generation
	})

	// Anticipate collisions
	g := 1
	var parts []*partition
	for p := 0; p < len(i.partitions); p++ {
		if i.partitions[p].generation == g {
			parts = append(parts, i.partitions[p])
			g++
		} else {
			break
		}
	}

	if len(parts) == 0 {
		oldPath := current.getPath()
		// Set current partition as final
		current.generation = 1
		os.Rename(oldPath, current.getPath())
		current.loadDict()
	} else {
		// Remove old partitions from the index
		i.partitions = i.partitions[len(parts):]
		i.curentPart = nil
		// Merge partitions
		p := merge(i.dir, "postings", g, append(parts, current))
		i.partitions = append(i.partitions, p)
		p.loadDict()
	}

	i.addPartition()
}

// ClearMemory writes any remaining partitions to disk
func (i *Index) ClearMemory() {
	i.dumpInfo()
	i.curentPart.dump()
	i.docs.Dump()
}

func (i *Index) createDir() {
	err := os.RemoveAll(i.dir)
	err = os.Mkdir(i.dir, 0755)
	if err != nil {
		log.Fatal("Could not create index directory")
	}
}
