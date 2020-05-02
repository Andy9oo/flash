package index

import (
	"bytes"
	"encoding/binary"
	"flash/tools/readers"
	"log"
	"os"
	"strings"
)

type indexReader struct {
	file           *os.File
	currentTerm    string
	postingsLength uint32
	done           bool
}

func newIndexReader(path string) *indexReader {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Could not open file: %v\n", path)
	}

	r := &indexReader{
		file: f,
		done: false,
	}

	r.fetchNextTerm()
	return r
}

func (r *indexReader) fetchNextTerm() (ok bool) {
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

func (r *indexReader) fetchPostingsLength() uint32 {
	r.postingsLength = readers.ReadUint32(r.file)
	return r.postingsLength
}

func (r *indexReader) fetchPostings() *bytes.Buffer {
	buf := make([]byte, r.postingsLength)
	r.file.Read(buf)
	return bytes.NewBuffer(buf)
}

func (r *indexReader) skipPostings() {
	r.file.Seek(int64(r.postingsLength), os.SEEK_CUR)
}

func (r *indexReader) fetchEntry(offset int64) (term string, buf *bytes.Buffer) {
	r.file.Seek(offset, os.SEEK_SET)
	r.fetchNextTerm()
	r.fetchPostingsLength()
	buf = r.fetchPostings()

	return r.currentTerm, buf
}

func (r *indexReader) findPostings(term string, start int64, end int64) (buf *bytes.Buffer, ok bool) {
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

func (r *indexReader) close() {
	r.file.Close()
	r.done = true
}

func (r *indexReader) compare(s string) int {
	return strings.Compare(r.currentTerm, s)
}
