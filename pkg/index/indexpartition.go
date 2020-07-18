package index

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flash/pkg/index/partition"
	"flash/pkg/index/postinglist"
	"flash/tools/readers"
	"fmt"
	"os"
	"strconv"
)

// Partition implements the partition.Implementation interface
type Partition struct {
	data        map[string]*postinglist.List
	invalidDocs map[uint64]bool
}

type postingEntry struct {
	docID  uint64
	offset uint32
}

// NewPartition creates a new indexPartition
func NewPartition() partition.Implementation {
	p := Partition{
		data:        make(map[string]*postinglist.List),
		invalidDocs: make(map[uint64]bool),
	}

	return &p
}

// Add adds a term and an entry to the index
func (p *Partition) Add(term string, entry partition.Entry) {
	switch entry.(type) {
	case *postingEntry:
		e := entry.(*postingEntry)
		if _, ok := p.data[term]; !ok {
			p.data[term] = postinglist.NewList()
		}
		p.data[term].Add(e.docID, e.offset)
	case *postinglist.List:
		p.data[term] = entry.(*postinglist.List)
	}
}

// Delete removes a document from the index
func (p *Partition) Delete(doc string) {
	id, err := strconv.ParseUint(doc, 10, 64)
	if err != nil {
		panic(err)
	}

	// Invalidate the doc if the partition is on disk
	if p.Empty() {
		p.invalidDocs[id] = true
		return
	}

	// If partition is in memory, remove the postings for the doc
	for term, pl := range p.data {
		pl.Delete(id)
		if pl.Empty() {
			delete(p.data, term)
		}
	}
}

// Get returns an entry for a given term from the index
func (p *Partition) Get(term string) (partition.Entry, bool) {
	if val, ok := p.data[term]; ok {
		return val, true
	}
	return nil, false
}

// Decode returns a posting list created from the buffer
func (p *Partition) Decode(buf *bytes.Buffer) (partition.Entry, bool) {
	return postinglist.Decode(buf, p.invalidDocs)
}

// Empty returns true if the index is empty
func (p *Partition) Empty() bool {
	return len(p.data) == 0
}

// Keys returns the set of keys added to the index
func (p *Partition) Keys() []string {
	keys := make([]string, 0, len(p.data))
	for k := range p.data {
		keys = append(keys, k)
	}
	return keys
}

// Clear removes the data from the partition
func (p *Partition) Clear() {
	p.data = nil
}

// Merge merges the posting lists given by the set of readers
func (p *Partition) Merge(readers []*partition.Reader) partition.Entry {
	plist := postinglist.NewList()
	for i := 0; i < len(readers); i++ {
		readers[i].FetchDataLength()
		r := postinglist.NewReader(readers[i].FetchData(), p.invalidDocs)

		for r.Read() {
			id, _, offsets := r.Data()
			plist.Add(id, offsets...)
		}
	}
	return plist
}

// LoadInfo loads in information about the partition into memory
func (p *Partition) LoadInfo(path string) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	r := bufio.NewReader(f)

	num := readers.ReadUint32(r)
	for i := uint32(0); i < num; i++ {
		key := readers.ReadUint64(r)
		p.invalidDocs[key] = true
	}
}

// GetInfo returns a buffer containing info that must be saved about the partition
func (p *Partition) GetInfo() *bytes.Buffer {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(len(p.invalidDocs)))
	for key := range p.invalidDocs {
		binary.Write(buf, binary.LittleEndian, key)
	}
	return buf
}

func (pe *postingEntry) Bytes() *bytes.Buffer {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, pe.docID)
	binary.Write(buf, binary.LittleEndian, pe.offset)
	return buf
}
