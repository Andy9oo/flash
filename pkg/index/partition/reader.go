package partition

import (
	"bytes"
	"encoding/binary"
	"flash/tools/readers"
	"log"
	"os"
	"strings"
)

// Reader for processing partitions
type Reader struct {
	file       *os.File
	currentKey string
	dataLength uint32
	done       bool
}

// NewReader creates a new partition reader
func NewReader(target string) *Reader {
	f, err := os.Open(target)
	if err != nil {
		log.Fatalf("Could not open file: %v\n", target)
	}

	r := &Reader{
		file: f,
		done: false,
	}

	r.NextKey()
	return r
}

// NextKey reads the next key from the partition, returns false if there isn't another key
func (r *Reader) NextKey() (ok bool) {
	if r.done {
		return false
	}

	keyLen := readers.ReadUint32(r.file)
	buf := make([]byte, keyLen)
	n, err := r.file.Read(buf)
	if n == 0 || err != nil {
		r.done = true
		r.file.Close()
		return false
	}

	r.currentKey = string(buf)
	return true
}

// FetchDataLength reads the length of the data section for the current key
func (r *Reader) FetchDataLength() uint32 {
	r.dataLength = readers.ReadUint32(r.file)
	return r.dataLength
}

// FetchData reads in the data portion for the current key
func (r *Reader) FetchData() *bytes.Buffer {
	buf := make([]byte, r.dataLength)
	r.file.Read(buf)
	return bytes.NewBuffer(buf)
}

// SkipData will seek the file past the data section without reading it
func (r *Reader) SkipData() {
	r.file.Seek(int64(r.dataLength), os.SEEK_CUR)
}

func (r *Reader) fetchEntry(offset int64) (term string, buf *bytes.Buffer) {
	r.file.Seek(offset, os.SEEK_SET)
	r.NextKey()
	r.FetchDataLength()
	buf = r.FetchData()

	return r.currentKey, buf
}

func (r *Reader) findEntry(key string, start int64, end int64) (buf *bytes.Buffer, ok bool) {
	r.file.Seek(start, os.SEEK_SET)
	blockSize := end - start

	block := make([]byte, blockSize)
	r.file.Read(block)

	i := int64(0)
	for i < blockSize {
		keyLen := int64(binary.LittleEndian.Uint32(block[i : i+4]))
		i += 4

		k := string(block[i : i+keyLen])
		i += keyLen

		dataLen := int64(binary.LittleEndian.Uint32(block[i : i+4]))
		i += 4

		if k == key {
			return bytes.NewBuffer(block[i : i+dataLen]), true
		}

		i += dataLen
	}

	return nil, false
}

// Close the reader
func (r *Reader) Close() {
	r.file.Close()
	r.done = true
}

func (r *Reader) compare(s string) int {
	return strings.Compare(r.currentKey, s)
}
