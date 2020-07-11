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
	indexpath  string
	generation int
	postings   map[string]*postinglist.List
	dict       *dictionary
	size       int
}

func newPartition(indexpath string, generation int) *partition {
	p := partition{
		indexpath:  indexpath,
		generation: generation,
		postings:   make(map[string]*postinglist.List),
	}

	return &p
}

func loadPartition(indexpath string, generation int) *partition {
	p := newPartition(indexpath, generation)

	if generation == 0 {
		p.loadPostings()
	} else {
		p.dict = loadDictionary(p.getPath(), dictionaryLimit)
	}

	return p
}

// GetPostingReader returns a posting reader for a term
func (p *partition) GetPostingReader(term string) (*postinglist.Reader, bool) {
	if p.generation == 0 {
		if pl, ok := p.postings[term]; ok {
			return postinglist.NewReader(pl.Bytes()), true
		}
		return nil, false
	}

	if buf, ok := p.dict.getBuffer(term); ok {
		return postinglist.NewReader(buf), true
	}
	return nil, false
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

	p.bytes().WriteTo(f)
	p.postings = nil
	p.size = 0
}

func (p *partition) bytes() *bytes.Buffer {
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

	return buf
}

func (p *partition) loadPostings() {
	reader := NewReader(p.getPath())
	defer reader.Close()

	for !reader.done {
		term := reader.currentKey
		reader.fetchDataLength()
		buf := reader.fetchData()

		p.postings[term] = postinglist.Decode(buf)
		reader.nextKey()
	}
}

func (p *partition) loadDict() {
	p.dict = loadDictionary(p.getPath(), dictionaryLimit)
}

func (p *partition) getPath() string {
	if p.generation == 0 {
		return fmt.Sprintf("%v/temp.postings", p.indexpath)
	}
	return fmt.Sprintf("%v/part_%d.postings", p.indexpath, p.generation)
}
