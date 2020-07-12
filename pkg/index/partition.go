package index

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"sort"
)

type partitionImpl interface {
	add(key string, val partitionEntry)
	get(key string) (val partitionEntry, ok bool)
	decode(*bytes.Buffer) partitionEntry
	empty() bool
	keys() []string
	clear()
}

type partitionEntry interface {
	Bytes() *bytes.Buffer
}

type partition struct {
	indexpath  string
	extension  string
	generation int
	impl       partitionImpl
	dict       *Dictionary
	size       int
	limit      int
}

func newPartition(indexpath, extension string, generation, limit int, impl partitionImpl) *partition {
	p := partition{
		indexpath:  indexpath,
		extension:  extension,
		generation: generation,
		impl:       impl,
		limit:      limit,
	}

	return &p
}

func loadPartition(indexpath, extension string, generation, limit int, impl partitionImpl) *partition {
	p := newPartition(indexpath, extension, generation, limit, impl)

	if generation == 0 {
		p.loadData()
	} else {
		p.dict = loadDictionary(p.getPath(), dictionaryLimit)
	}

	return p
}

// GetPostingReader returns a posting reader for a term
func (p *partition) GetBuffer(term string) (*bytes.Buffer, bool) {
	if p.generation == 0 {
		if val, ok := p.impl.get(term); ok {
			return val.Bytes(), true
		}
		return nil, false
	}

	if buf, ok := p.dict.getBuffer(term); ok {
		return buf, true
	}

	return nil, false
}

func (p *partition) add(key string, val partitionEntry) {
	p.impl.add(key, val)
	p.size++
}

func (p *partition) full() bool {
	return p.size >= p.limit
}

func (p *partition) dump() {
	if p.impl.empty() {
		return
	}

	f, err := os.Create(p.getPath())
	if err != nil {
		log.Fatal("Could not create index partition")
	}
	defer f.Close()

	p.bytes().WriteTo(f)
	p.impl.clear()
	p.size = 0
}

func (p *partition) bytes() *bytes.Buffer {
	keys := p.impl.keys()
	sort.Strings(keys)

	buf := new(bytes.Buffer)
	for _, key := range keys {
		data, _ := p.impl.get(key)
		dataBuf := data.Bytes()

		binary.Write(buf, binary.LittleEndian, uint32(len(key)))
		binary.Write(buf, binary.LittleEndian, []byte(key))
		binary.Write(buf, binary.LittleEndian, uint32(dataBuf.Len()))
		binary.Write(buf, binary.LittleEndian, dataBuf.Bytes())
	}

	return buf
}

func (p *partition) loadData() {
	reader := NewReader(p.getPath())
	defer reader.Close()

	for !reader.done {
		key := reader.currentKey
		reader.fetchDataLength()
		buf := reader.fetchData()
		p.add(key, p.impl.decode(buf))
		reader.nextKey()
	}
}

func (p *partition) loadDict() {
	p.dict = loadDictionary(p.getPath(), dictionaryLimit)
}

func (p *partition) getPath() string {
	if p.generation == 0 {
		return fmt.Sprintf("%v/temp.%v", p.indexpath, p.extension)
	}
	return fmt.Sprintf("%v/part_%d.%v", p.indexpath, p.generation, p.extension)
}
