package index

import (
	"bytes"
	"encoding/binary"
	"flash/pkg/index/partition"
	"flash/pkg/index/postinglist"
	"flash/tools/readers"
	"fmt"
	"io"
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
func (p *Partition) Decode(term string, buf *bytes.Buffer) (partition.Entry, bool) {
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
func (p *Partition) Merge(readers []*partition.Reader, impls []partition.Implementation) partition.Entry {
	plist := postinglist.NewList()
	for i := 0; i < len(readers); i++ {
		if ip, ok := impls[i].(*Partition); ok {
			readers[i].FetchDataLength()
			r := postinglist.NewReader(readers[i].FetchData(), ip.invalidDocs)

			for r.Read() {
				id, _, offsets := r.Data()
				plist.Add(id, offsets...)
			}
		}
	}
	return plist
}

// LoadInfo loads in information about the partition into memory
func (p *Partition) LoadInfo(r io.Reader) {
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

// GC performs garbage collection on the partition
func (p *Partition) GC(reader *partition.Reader, out string) (size int) {
	temp, err := os.Create(out)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer temp.Close()

	running := true
	for running {
		reader.FetchDataLength()
		pr := postinglist.NewReader(reader.FetchData(), p.invalidDocs)

		postingBuf := new(bytes.Buffer)
		numDocs := 0
		for pr.Read() {
			id, freq, offsets := pr.Data()
			binary.Write(postingBuf, binary.LittleEndian, id)
			binary.Write(postingBuf, binary.LittleEndian, freq)
			for i := 0; i < len(offsets); i++ {
				binary.Write(postingBuf, binary.LittleEndian, offsets[i])
			}

			numDocs++
			size += int(freq)
		}

		if numDocs > 0 {
			buf := new(bytes.Buffer)
			key := reader.CurrentKey()

			binary.Write(buf, binary.LittleEndian, uint32(len(key)))
			binary.Write(buf, binary.LittleEndian, []byte(key))
			binary.Write(buf, binary.LittleEndian, uint32(4+postingBuf.Len()))
			binary.Write(buf, binary.LittleEndian, uint32(numDocs))

			buf.WriteTo(temp)
			postingBuf.WriteTo(temp)
		}
		running = reader.NextKey()
	}

	p.invalidDocs = make(map[uint64]bool)
	return size
}

func (pe *postingEntry) Bytes() *bytes.Buffer {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, pe.docID)
	binary.Write(buf, binary.LittleEndian, pe.offset)
	return buf
}
