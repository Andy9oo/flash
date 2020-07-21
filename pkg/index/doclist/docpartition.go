package doclist

import (
	"bytes"
	"encoding/binary"
	"flash/pkg/index/partition"
	"flash/tools/readers"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

// Partition implements the partition.Implementation interface for doclist
type Partition struct {
	data        map[string]*Document
	invalidDocs map[uint64]bool
}

// NewPartition returns a new partition
func NewPartition() partition.Implementation {
	p := Partition{
		data:        make(map[string]*Document),
		invalidDocs: make(map[uint64]bool),
	}

	return &p
}

// Add adds the document with the given id to the partition
func (p *Partition) Add(id string, val partition.Entry) {
	if doc, ok := val.(*Document); ok {
		p.data[id] = doc
	}
}

// Delete removes a doc from the doclist
func (p *Partition) Delete(id string) {
	docid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		panic(err)
	}

	// Invalidate the doc if the partition is on disk
	if p.Empty() {
		p.invalidDocs[docid] = true
		return
	}

	// If partition is in memory, remove the doc
	delete(p.data, id)
}

// Get returns the document with the given id
func (p *Partition) Get(id string) (val partition.Entry, ok bool) {
	if val, ok := p.data[id]; ok {
		return val, true
	}
	return nil, false
}

// Decode takes a byte buffer and decodes it to a document
func (p *Partition) Decode(buf *bytes.Buffer) (partition.Entry, bool) {
	id := readers.ReadUint64(buf)
	length := readers.ReadUint32(buf)
	plen := readers.ReadUint32(buf)
	pbuf := make([]byte, plen)
	io.ReadFull(buf, pbuf)

	doc := Document{
		id:     id,
		path:   string(pbuf),
		length: length,
	}

	valid := true
	if _, ok := p.invalidDocs[id]; ok {
		valid = false
	}

	return &doc, valid
}

// Merge will merge the partition readers, if there is more than one, this means that
// there was a collision in the doclist
func (p *Partition) Merge(readers []*partition.Reader, impls []partition.Implementation) partition.Entry {
	if len(readers) != 1 {
		log.Fatal("Collision occured in doclist")
	}
	readers[0].FetchDataLength()
	dp, _ := impls[0].(*Partition)
	if doc, ok := dp.Decode(readers[0].FetchData()); ok {
		return doc
	}
	return nil
}

// Empty returns true if the partition is empty
func (p *Partition) Empty() bool {
	return len(p.data) == 0
}

// Keys returns the list of doc ids
func (p *Partition) Keys() []string {
	keys := make([]string, 0, len(p.data))
	for k := range p.data {
		keys = append(keys, k)
	}
	return keys
}

// Clear clears the partition
func (p *Partition) Clear() {
	p.data = nil
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

// GC Performs garbage collection on the partition
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
		if doc, ok := p.Decode(reader.FetchData()); ok {
			data := doc.Bytes()
			buf := new(bytes.Buffer)
			key := reader.CurrentKey()
			binary.Write(buf, binary.LittleEndian, uint32(len(key)))
			binary.Write(buf, binary.LittleEndian, []byte(key))
			binary.Write(buf, binary.LittleEndian, uint32(data.Len()))
			binary.Write(buf, binary.LittleEndian, data.Bytes())
			buf.WriteTo(temp)
			size++
		}
		running = reader.NextKey()
	}
	return size
}
