package index

import (
	"bytes"
	"encoding/binary"
	"flash/tools/readers"
	"log"
	"os"
	"strings"
)

// Reader for processing indexes
type Reader struct {
	file           *os.File
	currentTerm    string
	postingsLength uint32
	done           bool
}

// NewReader creates a new index reader
func NewReader(path string) *Reader {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Could not open file: %v\n", path)
	}

	r := &Reader{
		file: f,
		done: false,
	}

	r.fetchNextTerm()
	return r
}

func (r *Reader) fetchNextTerm() (ok bool) {
	if r.done {
		return false
	}

	tlen := readers.ReadUint32(r.file)
	buf := make([]byte, tlen)
	n, err := r.file.Read(buf)
	if n == 0 || err != nil {
		r.done = true
		r.file.Close()
		return false
	}

	r.currentTerm = string(buf)
	return true
}

func (r *Reader) fetchPostingsLength() uint32 {
	r.postingsLength = readers.ReadUint32(r.file)
	return r.postingsLength
}

func (r *Reader) fetchPostings() *bytes.Buffer {
	buf := make([]byte, r.postingsLength)
	r.file.Read(buf)
	return bytes.NewBuffer(buf)
}

func (r *Reader) skipPostings() {
	r.file.Seek(int64(r.postingsLength), os.SEEK_CUR)
}

func (r *Reader) fetchEntry(offset int64) (term string, buf *bytes.Buffer) {
	r.file.Seek(offset, os.SEEK_SET)
	r.fetchNextTerm()
	r.fetchPostingsLength()
	buf = r.fetchPostings()

	return r.currentTerm, buf
}

func (r *Reader) findPostings(term string, start int64, end int64) (buf *bytes.Buffer, ok bool) {
	r.file.Seek(start, os.SEEK_SET)
	blockSize := end - start

	block := make([]byte, blockSize)
	r.file.Read(block)

	i := int64(0)
	for i < blockSize {
		tlen := int64(binary.LittleEndian.Uint32(block[i : i+4]))
		i += 4

		t := string(block[i : i+tlen])
		i += tlen

		plen := int64(binary.LittleEndian.Uint32(block[i : i+4]))
		i += 4

		if t == term {
			return bytes.NewBuffer(block[i : i+plen]), true
		}

		i += plen
	}

	return nil, false
}

// Close index reader
func (r *Reader) Close() {
	r.file.Close()
	r.done = true
}

func (r *Reader) compare(s string) int {
	return strings.Compare(r.currentTerm, s)
}
