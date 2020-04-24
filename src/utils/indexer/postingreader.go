package indexer

import (
	"bytes"
)

// PostingReader type for efficiently reading posting lists sequentially
type PostingReader struct {
	NumDocs uint32
	buffer  *bytes.Buffer
}

// NewPostingReader creates a new posting reader
func NewPostingReader(buf *bytes.Buffer) *PostingReader {
	r := PostingReader{
		buffer:  buf,
		NumDocs: readInt32(buf),
	}
	return &r
}

// NextPosting will return the next posting in the posting list
func (pr *PostingReader) NextPosting() (id, frequency uint32, offsets []uint32, ok bool) {
	if pr.buffer.Len() == 0 {
		return 0, 0, nil, false
	}

	id = readInt32(pr.buffer)
	frequency = readInt32(pr.buffer)

	offsets = make([]uint32, frequency)
	for i := uint32(0); i < frequency; i++ {
		offsets[i] = readInt32(pr.buffer)
	}

	return id, frequency, offsets, true
}
