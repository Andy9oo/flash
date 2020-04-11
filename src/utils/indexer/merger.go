package indexer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type partitionMerger struct {
	dir      string
	file     *os.File
	readers  []*indexReader
	finished int
}

func newPartitionMerger(dir string) *partitionMerger {
	pm := partitionMerger{
		dir: dir,
	}

	return &pm
}

func (pm *partitionMerger) mergePartitions() {
	pm.makePostingsFile()
	defer pm.file.Close()

	pm.openReaders()

	for pm.finished < len(pm.readers) {
		term, readers := pm.getNextTerm()
		plen := pm.getPostingsLength(readers)

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint32(len(term)))
		binary.Write(buf, binary.LittleEndian, []byte(term))
		binary.Write(buf, binary.LittleEndian, plen)
		buf.WriteTo(pm.file)

		pm.writePostings(readers)
		pm.advanceTerms(readers)
	}

	pm.deletePartitionFiles()
}

func (pm *partitionMerger) getNextTerm() (term string, readers []*indexReader) {
	for i := 0; i < len(pm.readers); i++ {
		if pm.readers[i].done {
			continue
		}

		cmp := pm.readers[i].compare(term)
		if cmp == -1 || term == "" { // If the current term is less than the selected term
			term = pm.readers[i].currentTerm
			readers = readers[:0] // Reset slice keeping allocated memory
			readers = append(readers, pm.readers[i])
		} else if cmp == 0 { // If the current term is equal to the selected term
			readers = append(readers, pm.readers[i])
		}
	}

	return term, readers
}

func (pm *partitionMerger) getPostingsLength(readers []*indexReader) (length uint32) {
	for _, r := range readers {
		length += r.fetchPostingsLength()
	}

	return length
}

func (pm *partitionMerger) writePostings(readers []*indexReader) {
	postingReaders := make([]*postingReader, len(readers))

	for i := range readers {
		postingReaders[i] = newPostingReader(readers[i].file, readers[i].postingsLength)
	}

	mergePostings(postingReaders, pm.file)
}

func (pm *partitionMerger) advanceTerms(readers []*indexReader) {
	for _, r := range readers {
		if ok := r.fetchNextTerm(); !ok {
			pm.finished++
		}
	}
}

func (pm *partitionMerger) makePostingsFile() {
	path := fmt.Sprintf("%v/index.postings", pm.dir)

	f, err := os.Create(path)
	if err != nil {
		log.Fatal("Could not create index file")
	}

	pm.file = f
}

func (pm *partitionMerger) openReaders() {
	for _, file := range pm.getPartitionFiles() {
		path := fmt.Sprintf("%v/%v", pm.dir, file)
		pm.readers = append(pm.readers, newIndexReader(path))
	}
}

func (pm *partitionMerger) getPartitionFiles() []string {
	var files []string

	dir, err := os.Open(pm.dir)
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

func (pm *partitionMerger) deletePartitionFiles() {
	files := pm.getPartitionFiles()
	for _, file := range files {
		path := fmt.Sprintf("%v/%v", pm.dir, file)
		os.Remove(path)
	}
}
