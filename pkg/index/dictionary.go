package index

import (
	"bytes"
	"encoding/binary"
	"flash/tools/readers"
	"fmt"
	"os"
	"sort"
)

type dictionary struct {
	indexpath  string
	generation int
	blockSize  int64
	entries    map[string]int64
	keys       []string
}

func loadDictionary(indexpath string, generation int, blockSize int64) *dictionary {
	d := dictionary{
		indexpath:  indexpath,
		generation: generation,
		blockSize:  blockSize,
		entries:    make(map[string]int64),
	}

	_, err := os.Stat(d.getPath())
	if err != nil {
		d.calculateOffsets()
		d.dump()
	} else {
		d.loadOffsets()
	}

	d.keys = make([]string, 0, len(d.entries))
	for key := range d.entries {
		d.keys = append(d.keys, key)
	}
	sort.Strings(d.keys)

	return &d
}

func (d *dictionary) getPostingBuffer(term string) (*bytes.Buffer, bool) {
	postingsFile := fmt.Sprintf("%v/part_%d.postings", d.indexpath, d.generation)
	indexReader := NewReader(postingsFile)
	defer indexReader.Close()

	if offset, ok := d.entries[term]; ok {
		_, buf := indexReader.fetchEntry(offset)
		return buf, true
	}

	pos := sort.SearchStrings(d.keys, term) - 1
	if pos == -1 || pos == len(d.keys)-1 {
		return nil, false
	}

	start := d.entries[d.keys[pos]]
	end := d.entries[d.keys[pos+1]]

	if buf, ok := indexReader.findPostings(term, start, end); ok {
		return buf, true
	}

	return nil, false
}

func (d *dictionary) loadOffsets() {
	f, err := os.Open(d.getPath())
	if err != nil {
		fmt.Println("Could not open dictionary file")
		return
	}
	defer f.Close()

	numTerms := readers.ReadUint32(f)
	for i := uint32(0); i < numTerms; i++ {
		tlen := readers.ReadUint32(f)

		tbuf := make([]byte, tlen)
		f.Read(tbuf)

		offset := readers.ReadUint64(f)
		d.entries[string(tbuf)] = int64(offset)
	}
}

func (d *dictionary) calculateOffsets() {
	postingsFile := fmt.Sprintf("%v/part_%d.postings", d.indexpath, d.generation)
	reader := NewReader(postingsFile)

	var remainingBytes int64
	var offset int64
	for {
		numBytes := int64(len(reader.currentTerm)) + int64(reader.fetchPostingsLength()) + 8 // 8 bytes used for offsets
		remainingBytes -= numBytes
		if remainingBytes <= 0 {
			d.entries[reader.currentTerm] = offset
			remainingBytes = d.blockSize
		}

		reader.skipPostings()
		reader.fetchNextTerm()
		offset += numBytes

		if reader.done {
			d.entries[reader.currentTerm] = offset - numBytes
			return
		}
	}
}

func (d *dictionary) dump() {
	f, err := os.Create(d.getPath())
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(len(d.entries)))
	for key, offset := range d.entries {
		binary.Write(buf, binary.LittleEndian, uint32(len(key)))
		binary.Write(buf, binary.LittleEndian, []byte(key))
		binary.Write(buf, binary.LittleEndian, uint64(offset))
	}

	buf.WriteTo(f)
}

func (d *dictionary) getPath() string {
	return fmt.Sprintf("%v/part_%d.dict", d.indexpath, d.generation)
}
