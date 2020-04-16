package indexer

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
)

type merger struct {
	dir       string
	output    *os.File
	part      *partition
	paritions []*partition
	readers   []*indexReader
	finished  int
}

func merge(dir string, outID int, partitions []*partition) *partition {
	m := merger{dir: dir, paritions: partitions, part: newPartition(dir, outID)}
	m.createOutputFile()
	defer m.output.Close()

	m.openReaders(partitions)
	for m.finished < len(m.readers) {
		term, readers := m.getNextTerm()

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint32(len(term)))
		binary.Write(buf, binary.LittleEndian, []byte(term))
		buf.WriteTo(m.output)

		m.mergePostings(readers)
		m.advanceTerms(readers)
	}

	m.deletePartitionFiles()
	return m.part
}

func (m *merger) getNextTerm() (term string, readers []*indexReader) {
	for i := 0; i < len(m.readers); i++ {
		if m.readers[i].done {
			continue
		}
		cmp := m.readers[i].compare(term)
		if cmp == -1 || term == "" { // If the current term is less than the selected term
			term = m.readers[i].currentTerm
			readers = readers[:0] // Reset slice keeping allocated memory
			readers = append(readers, m.readers[i])
		} else if cmp == 0 { // If the current term is equal to the selected term
			readers = append(readers, m.readers[i])
		}
	}
	return term, readers
}

func (m *merger) mergePostings(readers []*indexReader) {
	var plist postingList
	for i := 0; i < len(readers); i++ {
		readers[i].fetchPostingsLength()
		l := decodePostingList(readers[i].fetchPostings())

		for p := l.head; p != nil; p = p.next {
			plist.add(p.docID, p.offsets...)
		}
	}

	buf := plist.Bytes()
	binary.Write(m.output, binary.LittleEndian, uint32(len(buf)))
	m.output.Write(buf)
}

func (m *merger) advanceTerms(readers []*indexReader) {
	for i := range readers {
		if ok := readers[i].fetchNextTerm(); !ok {
			m.finished++
		}
	}
}

func (m *merger) createOutputFile() {
	path := m.part.getPath()
	f, err := os.Create(path)
	if err != nil {
		log.Fatal("Could not create index file")
	}
	m.output = f
}

func (m *merger) openReaders(partitions []*partition) {
	for i := range partitions {
		m.readers = append(m.readers, newIndexReader(partitions[i].getPath()))
	}
}

func (m *merger) deletePartitionFiles() {
	for i := range m.paritions {
		os.Remove(m.paritions[i].getPath())
	}
}
