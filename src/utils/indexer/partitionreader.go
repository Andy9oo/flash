package indexer

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

type partitionReader struct {
	file           *os.File
	currentTerm    string
	postingsLength uint32
	done           bool
}

func newPartitionReader(path string) *partitionReader {
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("Could not open file: %v\n", path)
	}

	pr := &partitionReader{
		file: f,
		done: false,
	}

	pr.fetchNextTerm()
	return pr
}

func (pr *partitionReader) compare(s string) int {
	return strings.Compare(pr.currentTerm, s)
}

func (pr *partitionReader) fetchPostingsLength() uint32 {
	buf := make([]byte, 4)
	pr.file.Read(buf)
	plen := binary.LittleEndian.Uint32(buf)
	pr.postingsLength = plen

	return plen
}

func (pr *partitionReader) fetchPostings() []byte {
	buf := make([]byte, pr.postingsLength)
	pr.file.Read(buf)

	return buf
}

func (pr *partitionReader) fetchNextTerm() (ok bool) {
	if pr.done {
		return false
	}

	buf := make([]byte, 4)
	pr.file.Read(buf)
	tlen := binary.LittleEndian.Uint32(buf)

	buf = make([]byte, tlen)
	n, err := pr.file.Read(buf)
	if n == 0 || err != nil {
		pr.done = true
		pr.file.Close()
		return false
	}

	pr.currentTerm = string(buf)
	return true
}
