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
)

// IDPartition implements the partition.Implementation interface for doclist
type IDPartition struct {
	data        map[string]*ID
	invalidDocs map[string]bool
}

// NewIDPartition returns a new partition
func NewIDPartition() partition.Implementation {
	p := IDPartition{
		data:        make(map[string]*ID),
		invalidDocs: make(map[string]bool),
	}

	return &p
}

// Add adds the document with the given id to the partition
func (p *IDPartition) Add(path string, val partition.Entry) {
	if id, ok := val.(*ID); ok {
		p.data[path] = id
	}
}

// Delete removes a doc from the doclist
func (p *IDPartition) Delete(path string) {
	// Invalidate the doc if the partition is on disk
	if p.Empty() {
		p.invalidDocs[path] = true
		return
	}

	// If partition is in memory, remove the doc
	delete(p.data, path)
}

// Get returns the document with the given id
func (p *IDPartition) Get(path string) (val partition.Entry, ok bool) {
	if val, ok := p.data[path]; ok {
		return val, true
	}
	return nil, false
}

// Decode takes a byte buffer and decodes it to a document
func (p *IDPartition) Decode(path string, buf *bytes.Buffer) (partition.Entry, bool) {
	if _, ok := p.invalidDocs[path]; ok {
		return nil, false
	}

	id := readers.ReadUint64(buf)
	return &ID{id}, true
}

// Merge will merge the partition readers, if there is more than one, this means that
// there was a collision in the doclist
func (p *IDPartition) Merge(readers []*partition.Reader, impls []partition.Implementation) partition.Entry {
	if len(readers) != 1 {
		log.Fatal("Collision occured in doclist")
	}
	readers[0].FetchDataLength()
	idp, _ := impls[0].(*IDPartition)
	if doc, ok := idp.Decode(readers[0].CurrentKey(), readers[0].FetchData()); ok {
		return doc
	}
	return nil
}

// Empty returns true if the partition is empty
func (p *IDPartition) Empty() bool {
	return len(p.data) == 0
}

// Keys returns the list of doc ids
func (p *IDPartition) Keys() []string {
	keys := make([]string, 0, len(p.data))
	for k := range p.data {
		keys = append(keys, k)
	}
	return keys
}

// Clear clears the partition
func (p *IDPartition) Clear() {
	p.data = nil
}

// LoadInfo loads in information about the partition into memory
func (p *IDPartition) LoadInfo(r io.Reader) {
	num := readers.ReadUint32(r)
	for i := uint32(0); i < num; i++ {
		klen := readers.ReadUint32(r)
		kbuf := make([]byte, klen)
		io.ReadFull(r, kbuf)
		p.invalidDocs[string(kbuf)] = true
	}
}

// GetInfo returns a buffer containing info that must be saved about the partition
func (p *IDPartition) GetInfo() *bytes.Buffer {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(len(p.invalidDocs)))
	for key := range p.invalidDocs {
		binary.Write(buf, binary.LittleEndian, uint32(len(key)))
		binary.Write(buf, binary.LittleEndian, []byte(key))
	}
	return buf
}

// GC Performs garbage collection on the partition
func (p *IDPartition) GC(reader *partition.Reader, out string) (size int) {
	temp, err := os.Create(out)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer temp.Close()

	running := true
	for running {
		reader.FetchDataLength()
		if id, ok := p.Decode(reader.CurrentKey(), reader.FetchData()); ok {
			data := id.Bytes()
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
