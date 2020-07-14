package doclist

import (
	"bytes"
	"flash/pkg/index/partition"
	"flash/tools/readers"
	"log"
)

// Partition implements the partition.Implementation interface for doclist
type Partition struct {
	data map[string]*Document
}

// NewPartition returns a new partition
func NewPartition() partition.Implementation {
	p := Partition{
		data: make(map[string]*Document),
	}

	return &p
}

// Add adds the document with the given id to the partition
func (p *Partition) Add(id string, val partition.Entry) {
	if doc, ok := val.(*Document); ok {
		p.data[id] = doc
	}
}

// Get returns the document with the given id
func (p *Partition) Get(id string) (val partition.Entry, ok bool) {
	if val, ok := p.data[id]; ok {
		return val, true
	}
	return nil, false
}

// Decode takes a byte buffer and decodes it to a document
func (p *Partition) Decode(buf *bytes.Buffer) partition.Entry {
	id := readers.ReadUint64(buf)
	length := readers.ReadUint32(buf)
	plen := readers.ReadUint32(buf)
	pbuf := make([]byte, plen)
	buf.Read(pbuf)

	doc := Document{
		id:     id,
		path:   string(pbuf),
		length: length,
	}

	return &doc
}

func (p *Partition) Merge(readers []*partition.Reader) partition.Entry {
	if len(readers) != 1 {
		log.Fatal("Collision occured in doclist")
	}
	readers[0].FetchDataLength()
	return p.Decode(readers[0].FetchData())
}

// Empty returns true if the partition is empty
func (p *Partition) Empty() bool {
	return len(p.data) == 0
}

func (p *Partition) Keys() []string {
	keys := make([]string, 0, len(p.data))
	for k := range p.data {
		keys = append(keys, k)
	}
	return keys
}

func (p *Partition) Clear() {
	p.data = nil
}
