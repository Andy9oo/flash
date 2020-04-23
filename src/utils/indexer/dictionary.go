package indexer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"sort"
)

type dictionary struct {
	root      string
	path      string
	blockSize int64
	entries   map[string]int64
}

func loadDictionary(root string, blockSize int64) *dictionary {
	path := fmt.Sprintf("%v/index.dict", root)
	d := dictionary{
		root:      root,
		path:      path,
		blockSize: blockSize,
		entries:   make(map[string]int64),
	}

	_, err := os.Stat(path)
	if err != nil {
		d.calculateOffsets()
		d.dump()
	} else {
		d.loadOffsets()
	}

	return &d
}

func (d *dictionary) getPostings(term string) (*postingList, bool) {
	postingsFile := fmt.Sprintf("%v/index.postings", d.root)
	indexReader := newIndexReader(postingsFile)
	defer indexReader.close()

	if offset, ok := d.entries[term]; ok {
		_, postings := indexReader.fetchEntry(offset)
		return postings, true
	}

	keys := make([]string, 0, len(d.entries))
	for key := range d.entries {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	pos := sort.SearchStrings(keys, term) - 1

	if pos == len(keys)-1 {
		return nil, false
	}

	start := d.entries[keys[pos]]
	end := d.entries[keys[pos+1]]

	return indexReader.findPostings(term, start, end)
}

func (d *dictionary) getNumDocs(term string) int {
	pl, ok := d.getPostings(term)
	if !ok {
		return 0
	}

	return len(pl.docs)
}

func (d *dictionary) getFrequency(term string, doc uint32) int {
	if pl, ok := d.getPostings(term); ok {
		if postings, ok := pl.postings[doc]; ok {
			return int(postings.frequency)
		}
	}

	return 0
}

func (d *dictionary) loadOffsets() {
	f, err := os.Open(d.path)
	if err != nil {
		fmt.Println("Could not open dictionary file")
		return
	}
	defer f.Close()

	numTerms := readInt32(f)
	for i := uint32(0); i < numTerms; i++ {
		tlen := readInt32(f)

		tbuf := make([]byte, tlen)
		f.Read(tbuf)

		offset := readInt64(f)
		d.entries[string(tbuf)] = int64(offset)
	}
}

func (d *dictionary) calculateOffsets() {
	postingsFile := fmt.Sprintf("%v/index.postings", d.root)
	reader := newIndexReader(postingsFile)

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
	f, err := os.Create(d.path)
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
