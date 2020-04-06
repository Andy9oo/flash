package indexer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	partitionSizeLimit = uint32(1e5)
	documentListLimit  = uint32(1e4)
)

// Index datastructure
type Index struct {
	dir  string
	dict *dictionary
	docs *doclist
}

// BuildIndex builds a new index for the given directory
func BuildIndex(root string) *Index {
	dir := fmt.Sprintf("%v/.index", root)

	i := Index{
		dir:  dir,
		dict: newDictionary(dir, 4096),
		docs: newDocList(dir, 100),
	}

	i.mkdir()
	i.index(root)
	return &i
}

func (i *Index) index(dir string) {
	partition := newPartition(dir, 0)
	var docID uint32

	visit := func(path string, info os.FileInfo, err error) error {
		if info.Name()[0:1] == "." {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.Mode().IsRegular() {
			partition.add(path, docID)
			i.docs.add(docID, path)
			docID++
		}

		return nil
	}

	err := filepath.Walk(dir, visit)
	if err != nil {
		fmt.Println(err)
	}

	partition.dump()
	i.docs.dump()
	i.mergePartitions()
}

func (i *Index) mergePartitions() {
	f := i.createPostingsFile()
	defer f.Close()

	readers := i.getPartitionReaders()

	var finished int
	var selectedReaders []*partitionReader
	for finished < len(readers) {
		term := ""
		for i := 0; i < len(readers); i++ {
			if readers[i].done {
				continue
			}

			cmp := readers[i].compare(term)
			if cmp == -1 || term == "" { // If the current term is less than the selected term
				term = readers[i].currentTerm
				selectedReaders = selectedReaders[:0] // Reset slice keeping allocated memory
				selectedReaders = append(selectedReaders, readers[i])
			} else if cmp == 0 { // If the current term is equal to the selected term
				selectedReaders = append(selectedReaders, readers[i])
			}
		}

		var total uint32
		for _, r := range selectedReaders {
			total += r.fetchPostingsLength()
		}

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint32(len(term)))
		binary.Write(buf, binary.LittleEndian, []byte(term))
		binary.Write(buf, binary.LittleEndian, total)
		buf.WriteTo(f)

		for _, r := range selectedReaders {
			f.Write(r.fetchPostings())
			if ok := r.fetchNextTerm(); !ok {
				finished++
			}
		}
	}
	i.deletePartitionFiles()
}

func (i *Index) createPostingsFile() *os.File {
	path := fmt.Sprintf("%v/index.postings", i.dir)

	f, err := os.Create(path)
	if err != nil {
		log.Fatal("Could not create index file")
	}

	return f
}

func (i *Index) getPartitionReaders() []*partitionReader {
	var readers []*partitionReader

	for _, file := range i.getPartitionFiles() {
		path := fmt.Sprintf("%v/%v", i.dir, file)
		readers = append(readers, newPartitionReader(path))
	}

	return readers
}

func (i *Index) getPartitionFiles() []string {
	var files []string

	dir, err := os.Open(i.dir)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()

	all, err := dir.Readdirnames(-1)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range all {
		if filepath.Ext(file) == ".part" {
			files = append(files, file)
		}
	}

	return files
}

func (i *Index) deletePartitionFiles() {
	files := i.getPartitionFiles()
	for _, file := range files {
		path := fmt.Sprintf("%v/%v", i.dir, file)
		os.Remove(path)
	}
}

func (i *Index) mkdir() {
	err := os.RemoveAll(i.dir)
	err = os.Mkdir(i.dir, 0755)
	if err != nil {
		log.Fatal("Could not create index directory")
	}
}
