package postinglist

import (
	"bytes"
	"flash/tools/readers"
)

// Reader type for efficiently reading posting lists sequentially
type Reader struct {
	numDocs     uint32
	invalidDocs map[uint64]bool
	buffer      *bytes.Buffer
	id          uint64
	frequency   uint32
}

// NewReader creates a new posting reader
func NewReader(buf *bytes.Buffer, invalidDocs map[uint64]bool) *Reader {
	r := Reader{
		buffer:      buf,
		numDocs:     readers.ReadUint32(buf),
		invalidDocs: invalidDocs,
	}
	return &r
}

func (r *Reader) Read() (ok bool) {
	if r.buffer.Len() == 0 {
		return false
	}

	r.id = readers.ReadUint64(r.buffer)
	r.frequency = readers.ReadUint32(r.buffer)

	if _, ok := r.invalidDocs[r.id]; ok {
		return r.Read()
	}

	return true
}

// Data will return the data which has been read
func (r *Reader) Data() (id uint64, frequency uint32) {
	return r.id, r.frequency
}

// NumDocs returns the number of documents in the posting list
func (r *Reader) NumDocs() uint32 {
	return r.numDocs
}
