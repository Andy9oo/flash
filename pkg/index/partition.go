package index

import (
	"bytes"
	"encoding/binary"
	"flash/pkg/index/postinglist"
	"fmt"
	"log"
	"os"
	"sort"
)

type partition struct {
	indexpath       string
	partitionNumber int
	postings        map[string]*postinglist.List
	size            int
}

func newPartition(indexpath string, partitionNumber int) *partition {
	p := partition{
		indexpath:       indexpath,
		partitionNumber: partitionNumber,
		postings:        make(map[string]*postinglist.List),
	}

	return &p
}

func (p *partition) add(term string, docID uint32, offset uint32) {
	if _, ok := p.postings[term]; !ok {
		p.postings[term] = postinglist.NewList()
	}
	p.postings[term].Add(docID, offset)
	p.size++
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
	return fmt.Sprintf("%v/p%d.part", p.indexpath, p.partitionNumber)
}
