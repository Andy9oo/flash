package doclist

import (
	"bytes"
	"encoding/binary"
	"flash/pkg/index/partition"
)

// Document datastructure
type Document struct {
	id     uint64
	path   string
	length uint32
}

// ID returns the documents id
func (d *Document) ID() uint64 {
	return d.id
}

// Path returns the documents path
func (d *Document) Path() string {
	return d.path
}

// Length returns the documents length
func (d *Document) Length() uint32 {
	return d.length
}

// Bytes creates a byte buffer from the document
func (d *Document) Bytes() *bytes.Buffer {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, d.id)
	binary.Write(buf, binary.LittleEndian, d.length)
	binary.Write(buf, binary.LittleEndian, uint32(len(d.path)))
	binary.Write(buf, binary.LittleEndian, []byte(d.path))
	return buf
}

// Equal returns true if the given document has the same path
func (d *Document) Equal(val partition.Entry) bool {
	if doc, ok := val.(*Document); ok && doc.path == d.path {
		return true
	}
	return false
}
