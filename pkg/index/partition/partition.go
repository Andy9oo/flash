package partition

import (
	"bufio"
	"bytes"
	"compress/flate"
	"encoding/binary"
	"flash/tools/readers"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
)

// Implementation represents a partition implementation
type Implementation interface {
	Add(key string, val Entry)
	Delete(key string)
	Get(key string) (val Entry, ok bool)
	Decode(key string, buf *bytes.Buffer) (val Entry, ok bool)
	Merge([]*Reader, []Implementation) Entry
	Empty() bool
	Keys() []string
	LoadInfo(io.Reader)
	GetInfo() *bytes.Buffer
	Clear()
	GC(*Reader, string) (size int)
}

// Entry is used as values inserted into the partitions
type Entry interface {
	Bytes() *bytes.Buffer
}

type partition struct {
	indexpath         string
	extension         string
	generation        int
	impl              Implementation
	dict              *Dictionary
	size              int
	deleted           int
	deletionThreshold float64
	limit             int
}

const dictionaryLimit = 1 << 20

func newPartition(indexpath, extension string, generation, limit int, impl Implementation) *partition {
	p := partition{
		indexpath:         indexpath,
		extension:         extension,
		generation:        generation,
		impl:              impl,
		limit:             limit,
		deletionThreshold: float64(generation*limit) * 0.1,
	}

	return &p
}

func loadPartition(indexpath, extension string, generation, limit int, impl Implementation) *partition {
	p := newPartition(indexpath, extension, generation, limit, impl)
	p.loadInfo()

	if generation == 0 {
		fmt.Println("Loading data")
		p.loadData()
	} else {
		p.dict = loadDictionary(p.getPath(), dictionaryLimit)
	}

	return p
}

func (p *partition) updateGeneration(gen int) {
	p.generation = gen
	p.deletionThreshold = float64(gen*p.limit) * 0.1
}

func (p *partition) getBuffer(key string) (*bytes.Buffer, bool) {
	if p.generation == 0 {
		if val, ok := p.impl.Get(key); ok {
			return val.Bytes(), true
		}
		return nil, false
	}

	if buf, ok := p.dict.getBuffer(key); ok {
		return buf, true
	}

	return nil, false
}

func (p *partition) getEntry(key string) (Entry, bool) {
	if buf, ok := p.getBuffer(key); ok {
		return p.impl.Decode(key, buf)
	}
	return nil, false
}

func (p *partition) add(key string, val Entry) {
	p.impl.Add(key, val)
	p.size++
}

func (p *partition) delete(key string) {
	p.impl.Delete(key)
	p.deleted++
	p.size--

	if p.deleted > int(p.deletionThreshold) && p.generation != 0 {
		preader := NewReader(p.getPath())
		p.size = p.impl.GC(preader, p.getPath()+".temp")
		os.Rename(p.getPath()+".temp", p.getPath())
		os.Remove(p.dict.getPath())
		p.loadDict()
		p.deleted = 0
	}
}

func (p *partition) full() bool {
	return p.size >= p.limit
}

func (p *partition) dump() {
	f, err := os.Create(p.getPath())
	if err != nil {
		log.Fatal("Could not create index partition")
	}
	defer f.Close()

	p.bytes().WriteTo(f)
	p.impl.Clear()
}

func (p *partition) bytes() *bytes.Buffer {
	keys := p.impl.Keys()
	sort.Strings(keys)

	buf := new(bytes.Buffer)
	for _, key := range keys {
		data, _ := p.impl.Get(key)
		dataBuf := data.Bytes()

		compressed := new(bytes.Buffer)
		f, _ := flate.NewWriter(compressed, flate.BestSpeed)
		io.Copy(f, dataBuf)
		f.Close()

		binary.Write(buf, binary.LittleEndian, uint32(len(key)))
		binary.Write(buf, binary.LittleEndian, []byte(key))
		binary.Write(buf, binary.LittleEndian, uint32(compressed.Len()))
		binary.Write(buf, binary.LittleEndian, compressed.Bytes())
	}

	return buf
}

func (p *partition) loadData() {
	reader := NewReader(p.getPath())
	defer reader.Close()

	for !reader.done {
		key := reader.currentKey
		reader.FetchDataLength()
		buf := reader.FetchData()
		f := flate.NewReader(buf)
		decompressed := new(bytes.Buffer)
		io.Copy(decompressed, f)
		f.Close()

		fmt.Println(decompressed.Len())

		if val, ok := p.impl.Decode(key, decompressed); ok {
			p.add(key, val)
		}
		reader.NextKey()
	}
}

func (p *partition) loadDict() {
	p.dict = loadDictionary(p.getPath(), dictionaryLimit)
}

func (p *partition) loadInfo() {
	f, err := os.Open(p.getInfoPath())
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	p.deleted = int(readers.ReadUint32(r))
	p.impl.LoadInfo(r)
}

func (p *partition) dumpInfo() {
	path := fmt.Sprintf(p.getInfoPath())
	f, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(p.deleted))
	binary.Write(buf, binary.LittleEndian, p.impl.GetInfo().Bytes())
	buf.WriteTo(f)
}

func (p *partition) deleteFiles() {
	os.Remove(p.getPath())
	os.Remove(p.dict.getPath())
	os.Remove(p.getInfoPath())
}

func (p *partition) getPath() string {
	if p.generation == 0 {
		return fmt.Sprintf("%v/temp.%v", p.indexpath, p.extension)
	}
	return fmt.Sprintf("%v/part_%d.%v", p.indexpath, p.generation, p.extension)
}

func (p *partition) getInfoPath() string {
	return fmt.Sprintf("%v.info", p.getPath())
}
