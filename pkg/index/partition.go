package index

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
	size            int
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
		p.postings[term] = newPostingList()
		p.size += 8
	}
	p.postings[term].add(docID, offset)
	p.size += 4
}

func (p *partition) full() bool {
	return p.size >= partitionLimit
}

func (p *partition) dump() {
	if len(p.postings) == 0 {
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
		pl := p.postings[term]
		postings := pl.Bytes()

		binary.Write(buf, binary.LittleEndian, uint32(len(term)))
		binary.Write(buf, binary.LittleEndian, []byte(term))
		binary.Write(buf, binary.LittleEndian, uint32(postings.Len()))
		binary.Write(buf, binary.LittleEndian, postings.Bytes())
	}

	buf.WriteTo(f)
	p.postings = nil
	p.size = 0
}

func (p *partition) getPath() string {
	return fmt.Sprintf("%v/p%d.part", p.dir, p.partitionNumber)
}
