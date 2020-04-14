package indexer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"sort"
)

type partition struct {
	dir             string
	partitionNumber int
	postings        map[string]*postingList
	numPostings     uint32
}

func newPartition(dir string, partitionNumber int) *partition {
	p := partition{
		dir:             dir,
		partitionNumber: partitionNumber,
		postings:        make(map[string]*postingList),
	}

	return &p
}

func (p *partition) add(term string, docID uint32, offset uint32) {
	if _, ok := p.postings[term]; !ok {
		p.postings[term] = new(postingList)
	}

	p.postings[term].add(docID, offset)
	p.numPostings++
}

func (p *partition) full() bool {
	return p.numPostings >= postingsLimit
}

func (p *partition) dump() {
	if p.numPostings == 0 {
		return
	}

	f, err := os.Create(p.getPath())
	if err != nil {
		log.Fatal("Could not create index partition")
	}
	defer f.Close()

	terms := make([]string, len(p.postings))

	count := 0
	for t := range p.postings {
		terms[count] = t
		count++
	}

	sort.Strings(terms)

	buf := new(bytes.Buffer)
	for _, term := range terms {
		postings := p.postings[term].Bytes()

		binary.Write(buf, binary.LittleEndian, uint32(len(term)))
		binary.Write(buf, binary.LittleEndian, []byte(term))
		binary.Write(buf, binary.LittleEndian, uint32(len(postings)))
		binary.Write(buf, binary.LittleEndian, postings)
	}

	buf.WriteTo(f)
}

func (p *partition) getPath() string {
	return fmt.Sprintf("%v/p%d.part", p.dir, p.partitionNumber)
}
