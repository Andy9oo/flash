package partition

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
	readers   []*Reader
	finished  int
}

func merge(partitions []*partition, out *partition) {
	m := merger{dir: out.indexpath, paritions: partitions, part: out}
	m.createOutputFile()
	defer m.output.Close()

	m.openReaders(partitions)
	for m.finished < len(m.readers) {
		term, readers := m.getNextTerm()

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint32(len(term)))
		binary.Write(buf, binary.LittleEndian, []byte(term))
		buf.WriteTo(m.output)

		m.mergeData(readers)
		m.advanceTerms(readers)
	}

	m.deletePartitionFiles()
}

func (m *merger) getNextTerm() (term string, readers []*Reader) {
	for i := 0; i < len(m.readers); i++ {
		if m.readers[i].done {
			continue
		}
		cmp := m.readers[i].compare(term)
		if cmp == -1 || term == "" { // If the current term is less than the selected term
			term = m.readers[i].currentKey
			readers = readers[:0] // Reset slice keeping allocated memory
			readers = append(readers, m.readers[i])
		} else if cmp == 0 { // If the current term is equal to the selected term
			readers = append(readers, m.readers[i])
		}
	}
	return term, readers
}

func (m *merger) mergeData(readers []*Reader) {
	merged := m.part.impl.Merge(readers)
	buf := merged.Bytes()
	binary.Write(m.output, binary.LittleEndian, uint32(buf.Len()))
	m.output.Write(buf.Bytes())
}

func (m *merger) advanceTerms(readers []*Reader) {
	for i := range readers {
		if ok := readers[i].NextKey(); !ok {
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
		m.readers = append(m.readers, NewReader(partitions[i].getPath()))
	}
}

func (m *merger) deletePartitionFiles() {
	for i := range m.paritions {
		os.Remove(m.paritions[i].getPath())
		// Remove old dictionaries of non-temp partitions
		if m.paritions[i].generation != 0 {
			os.Remove(m.paritions[i].dict.getPath())
		}
	}
}
