package indexer

import (
	"fmt"
	"os"
	"sort"
	"strings"
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

	d.fecthOffsets()
	return &d
}

func (d *dictionary) getPostings(term string) (*postingList, bool) {
	indexPath := fmt.Sprintf("%v/index.postings", d.root)
	indexReader := newIndexReader(indexPath)
	term = strings.ToLower(term)

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

func (d *dictionary) fecthOffsets() {
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

	for key, val := range d.entries {
		f.WriteString(key)
		f.WriteString(fmt.Sprintf(":%v\n", val))
	}
}
