package partition

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"sort"
)

type Implementation interface {
	Add(key string, val Entry)
	Get(key string) (val Entry, ok bool)
	Decode(*bytes.Buffer) Entry
	Merge([]*Reader) Entry
	Empty() bool
	Keys() []string
	Clear()
}

type Entry interface {
	Bytes() *bytes.Buffer
}

type Partition struct {
	indexpath  string
	extension  string
	generation int
	impl       Implementation
	dict       *Dictionary
	size       int
	limit      int
}

const dictionaryLimit = 1 << 20

func NewPartition(indexpath, extension string, generation, limit int, impl Implementation) *Partition {
	p := Partition{
		indexpath:  indexpath,
		extension:  extension,
		generation: generation,
		impl:       impl,
		limit:      limit,
	}

	return &p
}

func LoadPartition(indexpath, extension string, generation, limit int, impl Implementation) *Partition {
	p := NewPartition(indexpath, extension, generation, limit, impl)

	if generation == 0 {
		p.loadData()
	} else {
		p.dict = loadDictionary(p.getPath(), dictionaryLimit)
	}

	return p
}

// GetPostingReader returns a posting reader for a term
func (p *Partition) GetBuffer(term string) (*bytes.Buffer, bool) {
	if p.generation == 0 {
		if val, ok := p.impl.Get(term); ok {
			return val.Bytes(), true
		}
		return nil, false
	}

	if buf, ok := p.dict.getBuffer(term); ok {
		return buf, true
	}

	return nil, false
}

func (p *Partition) add(key string, val Entry) {
	p.impl.Add(key, val)
	p.size++
}

func (p *Partition) full() bool {
	return p.size >= p.limit
}

func (p *Partition) dump() {
	if p.impl.Empty() {
		return
	}

	f, err := os.Create(p.getPath())
	if err != nil {
		log.Fatal("Could not create index partition")
	}
	defer f.Close()

	p.bytes().WriteTo(f)
	p.impl.Clear()
	p.size = 0
}

func (p *Partition) bytes() *bytes.Buffer {
	keys := p.impl.Keys()
	sort.Strings(keys)

	buf := new(bytes.Buffer)
	for _, key := range keys {
		data, _ := p.impl.Get(key)
		dataBuf := data.Bytes()

		binary.Write(buf, binary.LittleEndian, uint32(len(key)))
		binary.Write(buf, binary.LittleEndian, []byte(key))
		binary.Write(buf, binary.LittleEndian, uint32(dataBuf.Len()))
		binary.Write(buf, binary.LittleEndian, dataBuf.Bytes())
	}

	return buf
}

func (p *Partition) loadData() {
	reader := NewReader(p.getPath())
	defer reader.Close()

	for !reader.done {
		key := reader.currentKey
		reader.FetchDataLength()
		buf := reader.FetchData()
		p.add(key, p.impl.Decode(buf))
		reader.NextKey()
	}
}

func (p *Partition) loadDict() {
	p.dict = loadDictionary(p.getPath(), dictionaryLimit)
}

func (p *Partition) getPath() string {
	if p.generation == 0 {
		return fmt.Sprintf("%v/temp.%v", p.indexpath, p.extension)
	}
	return fmt.Sprintf("%v/part_%d.%v", p.indexpath, p.generation, p.extension)
}
