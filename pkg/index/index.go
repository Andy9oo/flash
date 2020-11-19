package index

import (
	"flash/pkg/importer"
	"flash/pkg/index/doclist"
	"flash/pkg/index/partition"
	"flash/pkg/index/postinglist"
	"flash/tools/blacklist"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/spf13/viper"
)

// Index datastructure
type Index struct {
	dir       string
	docs      *doclist.DocList
	collector *partition.Collector
	blacklist *blacklist.Blacklist
}

// Info contains information about an index
type Info struct {
	NumDocs   uint32
	AvgLength float64
}

// NewIndex creates a new index
func NewIndex(indexpath string) *Index {
	i := Index{
		dir:       indexpath,
		docs:      doclist.NewList(indexpath),
		collector: partition.NewCollector(indexpath, "postings", NewPartition),
		blacklist: &blacklist.Blacklist{},
	}

	i.blacklist.Add(viper.GetStringSlice("blacklist")...)
	i.createDir()
	return &i
}

// Load opens the index at the indexpath
func Load(indexpath string) *Index {
	i := &Index{
		dir:       indexpath,
		collector: partition.NewCollector(indexpath, "postings", NewPartition),
		blacklist: &blacklist.Blacklist{},
	}

	i.blacklist.Add(viper.GetStringSlice("blacklist")...)
	err := i.collector.Load()
	if err != nil {
		i = NewIndex(indexpath)
	} else {
		i.docs = doclist.Load(indexpath)
	}
	return i
}

// Add adds the given file or directory to the index
func (i *Index) Add(path string, lock *sync.RWMutex) {
	if i.blacklist.Contains(path) {
		return
	}

	stat, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(stat.Name()) > 0 && stat.Name()[0:1] == "." {
		return
	}

	var id uint64
	if sys, ok := stat.Sys().(*syscall.Stat_t); ok {
		id = sys.Ino
	}

	if !stat.IsDir() {
		lock.Lock()
		if _, ok := i.docs.FetchPath(path); ok {
			fmt.Println(path, "already in the index, removing and readding")
			i.Delete(path)
		}

		textChannel := importer.GetTextChannel(path)
		var offset uint32
		for term := range textChannel {
			i.collector.Add(term, &postingEntry{id})
			offset++
		}
		i.docs.Add(id, path, offset)
		lock.Unlock()
	} else {
		i.addDir(path, lock)
	}
}

// Delete removes the given file from the index
func (i *Index) Delete(path string) {
	for _, id := range i.docs.GetIDs(path) {
		i.collector.Delete(id.String())
		i.docs.Delete(id.String(), path)
	}
}

// GetPostingReaders returns a list of posting readers for the given term
func (i *Index) GetPostingReaders(term string) []*postinglist.Reader {
	bufs, impls := i.collector.GetBuffers(term)
	var readers []*postinglist.Reader
	for i := range bufs {
		if ip, ok := impls[i].(*Partition); ok {
			readers = append(readers, postinglist.NewReader(bufs[i], ip.invalidDocs))
		}
	}
	return readers
}

// GetInfo returns information about the index
func (i *Index) GetInfo() *Info {
	info := Info{
		NumDocs:   i.docs.NumDocs(),
		AvgLength: i.docs.AvgLength(),
	}
	return &info
}

// GetDocInfo returns information about the given document
func (i *Index) GetDocInfo(id uint64) (path string, length uint32, ok bool) {
	if doc, ok := i.docs.FetchID(id); ok {
		return doc.Path(), doc.Length(), true
	}
	return "", 0, false
}

func (i *Index) addDir(dir string, lock *sync.RWMutex) {
	visit := func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}

		if len(info.Name()) > 0 && info.Name()[0:1] == "." {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.Mode().IsRegular() {
			i.Add(path, lock)
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

// GetPath returns the path of the index
func (i *Index) GetPath() string {
	return i.dir
}

func (i *Index) createDir() {
	if _, err := os.Stat(i.dir); err != nil {
		fmt.Println(i.dir)
		err = os.MkdirAll(i.dir, 0755)
		if err != nil {
			log.Fatal("Could not create index directory")
		}
	}
}

// Blacklist blacklists the given pattern for the index
func (i *Index) Blacklist(pattern string) error {
	return i.blacklist.Add(pattern)
}

// RemoveBlacklist removes the given patern from the index
func (i *Index) RemoveBlacklist(pattern string) {
	i.blacklist.Remove(pattern)
}

// GetBlacklist returns the list of patterns in the blacklist
func (i *Index) GetBlacklist() []string {
	return i.blacklist.GetPatterns()
}

// ResetBlacklist removes all patterns from the blacklist
func (i *Index) ResetBlacklist() {
	i.blacklist.Reset()
}
