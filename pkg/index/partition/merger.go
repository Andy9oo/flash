package partition

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
)

type merger struct {
	dir        string
	output     *os.File
	part       *partition
	partitions []*partition
	readers    []*Reader
	finished   int
}

func merge(partitions []*partition, out *partition) {
	m := merger{dir: out.indexpath, partitions: partitions, part: out}
	m.createOutputFile()
	defer m.output.Close()

	m.openReaders(partitions)
	for m.finished < len(m.readers) {
		term, readers, impls := m.getNextTerm()

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint32(len(term)))
		binary.Write(buf, binary.LittleEndian, []byte(term))
		buf.WriteTo(m.output)

		m.mergeData(readers, impls)
		m.advanceTerms(readers)
	}

	m.deletePartitionFiles()
}

func (m *merger) getNextTerm() (term string, readers []*Reader, impls []Implementation) {
	for i := 0; i < len(m.readers); i++ {
		if m.readers[i].done {
			continue
		}
		cmp := m.readers[i].compare(term)
		if cmp == -1 || term == "" { // If the current term is less than the selected term
			term = m.readers[i].currentKey

			readers = readers[:0] // Reset slice keeping allocated memory
			impls = impls[:0]

			readers = append(readers, m.readers[i])
			impls = append(impls, m.partitions[i].impl)
		} else if cmp == 0 { // If the current term is equal to the selected term
			readers = append(readers, m.readers[i])
			impls = append(impls, m.partitions[i].impl)
		}
	}
	return term, readers, impls
}

func (m *merger) mergeData(readers []*Reader, impls []Implementation) {
	if merged := m.part.impl.Merge(readers, impls); merged != nil {
		buf := merged.Bytes()
		binary.Write(m.output, binary.LittleEndian, uint32(buf.Len()))
		m.output.Write(buf.Bytes())
	}
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
	for i := range m.partitions {
		os.Remove(m.partitions[i].getPath())
		// Remove old dictionaries of non-temp partitions
		if m.partitions[i].generation != 0 {
			os.Remove(m.partitions[i].dict.getPath())
		}
	}
}
