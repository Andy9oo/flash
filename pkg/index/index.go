package index

import (
	"flash/pkg/importer"
	"flash/pkg/index/doclist"
	"flash/pkg/index/partition"
	"flash/pkg/index/postinglist"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

// Index datastructure
type Index struct {
	dir       string
	docs      *doclist.DocList
	collector *partition.Collector
}

// Info contains information about an index
type Info struct {
	NumDocs     uint32
	TotalLength int
}

// NewIndex creates a new index
func NewIndex(indexpath string) *Index {
	i := Index{
		dir:       indexpath,
		docs:      doclist.NewList(indexpath),
		collector: partition.NewCollector(indexpath, "postings", NewPartition),
	}
	i.createDir()
	return &i
}

// Load opens the index at the indexpath
func Load(indexpath string) *Index {
	i := &Index{
		dir:       indexpath,
		collector: partition.NewCollector(indexpath, "postings", NewPartition),
	}

	err := i.collector.Load()
	if err != nil {
		i = NewIndex(indexpath)
	} else {
		i.docs = doclist.Load(indexpath)
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

	var id uint64
	if sys, ok := stat.Sys().(*syscall.Stat_t); ok {
		id = sys.Ino
	} else {
		fmt.Printf("Not a syscall.Stat_t")
		return
	}

	if !stat.IsDir() {
		textChannel := importer.GetTextChannel(path)

		var offset uint32
		for term := range textChannel {
			i.collector.Add(term, &postingEntry{id, offset})
			offset++
		}
		i.docs.Add(id, path, offset)
	} else {
		i.index(path)
	}
}

// GetPostingReaders returns a list of posting readers for the given term
func (i *Index) GetPostingReaders(term string) []*postinglist.Reader {
	var readers []*postinglist.Reader
	for _, buf := range i.collector.GetBuffers(term) {
		readers = append(readers, postinglist.NewReader(buf))
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
func (i *Index) GetDocInfo(id uint64) (path string, length uint32, ok bool) {
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
}

// ClearMemory writes any remaining partitions to disk
func (i *Index) ClearMemory() {
	i.collector.ClearMemory()
	i.docs.ClearMemory()
}

func (i *Index) createDir() {
	err := os.RemoveAll(i.dir)
	err = os.Mkdir(i.dir, 0755)
	if err != nil {
		log.Fatal("Could not create index directory")
	}
}
