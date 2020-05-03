package postinglist

import (
	"bytes"
	"flash/tools/readers"
)

// Reader type for efficiently reading posting lists sequentially
type Reader struct {
	numDocs   uint32
	buffer    *bytes.Buffer
	id        uint32
	frequency uint32
	offsets   []uint32
}

// NewReader creates a new posting reader
func NewReader(buf *bytes.Buffer) *Reader {
	r := Reader{
		buffer:  buf,
		numDocs: readers.ReadUint32(buf),
	}
	return &r
}

func (r *Reader) Read() (ok bool) {
	if r.buffer.Len() == 0 {
		return false
	}

	r.id = readers.ReadUint32(r.buffer)
	r.frequency = readers.ReadUint32(r.buffer)

	r.offsets = make([]uint32, r.frequency)
	for i := uint32(0); i < r.frequency; i++ {
		r.offsets[i] = readers.ReadUint32(r.buffer)
	}

	return true
}

// Data will return the data which has been read
func (r *Reader) Data() (id, frequency uint32, offsets []uint32) {
	return r.id, r.frequency, r.offsets
}

// NumDocs returns the number of documents in the posting list
func (r *Reader) NumDocs() uint32 {
	return r.numDocs
}
