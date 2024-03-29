package doclist

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Document datastructure
type Document struct {
	id     uint64
	path   string
	length uint32
}

// ID datastructure
type ID struct {
	uint64
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

// Bytes creates a byte buffer of the id
func (id *ID) Bytes() *bytes.Buffer {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, id)
	return buf
}

func (id *ID) String() string {
	return fmt.Sprint(id.uint64)
}
