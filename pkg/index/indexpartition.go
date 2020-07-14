package index

import (
	"bytes"
	"encoding/binary"
	"flash/pkg/index/partition"
	"flash/pkg/index/postinglist"
)

// Partition implements the partition.Implementation interface
type Partition struct {
	data map[string]*postinglist.List
}

type postingEntry struct {
	docID  uint32
	offset uint32
}

// NewPartition creates a new indexPartition
func NewPartition() partition.Implementation {
	p := Partition{
		data: make(map[string]*postinglist.List),
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

// Get returns an entry for a given term from the index
func (p *Partition) Get(term string) (partition.Entry, bool) {
	if val, ok := p.data[term]; ok {
		return val, true
	}
	return nil, false
}

// Decode returns a posting list created from the buffer
func (p *Partition) Decode(buf *bytes.Buffer) partition.Entry {
	return postinglist.Decode(buf)
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
		r := postinglist.NewReader(readers[i].FetchData())

		for r.Read() {
			id, _, offsets := r.Data()
			plist.Add(id, offsets...)
		}
	}
	return plist
}

func (pe *postingEntry) Bytes() *bytes.Buffer {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, pe.docID)
	binary.Write(buf, binary.LittleEndian, pe.offset)
	return buf
}
