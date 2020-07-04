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

type Partition struct {
	indexpath  string
	generation int
	postings   map[string]*postinglist.List
	dict       *dictionary
	size       int
}

func newPartition(indexpath string, generation int) *Partition {
	p := Partition{
		indexpath:  indexpath,
		generation: generation,
		postings:   make(map[string]*postinglist.List),
	}

	return &p
}

func loadPartition(indexpath string, generation int) *Partition {
	p := newPartition(indexpath, generation)
	p.dict = loadDictionary(indexpath, generation, dictionaryLimit)

	return p
}

// GetPostingReader returns a posting reader for a term
func (p *Partition) GetPostingReader(term string) (*postinglist.Reader, bool) {
	if buf, ok := p.dict.getPostingBuffer(term); ok {
		return postinglist.NewReader(buf), true
	}
	return nil, false
}

func (p *Partition) add(term string, docID uint32, offset uint32) {
	if _, ok := p.postings[term]; !ok {
		p.postings[term] = postinglist.NewList()
	}
	p.postings[term].Add(docID, offset)
	p.size++
}

func (p *Partition) full() bool {
	return p.size >= partitionLimit
}

func (p *Partition) dump() {
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

	p.dict = loadDictionary(p.indexpath, p.generation, dictionaryLimit)
}

func (p *Partition) getPath() string {
	return fmt.Sprintf("%v/part_%d.postings", p.indexpath, p.generation)
}
