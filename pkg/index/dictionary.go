package index

import (
	"bytes"
	"encoding/binary"
	"flash/tools/readers"
	"fmt"
	"os"
	"sort"
	"strings"
)

type dictionary struct {
	target    string
	blockSize int64
	entries   map[string]int64
	keys      []string
}

func loadDictionary(target string, blockSize int64) *dictionary {
	d := dictionary{
		target:    target,
		blockSize: blockSize,
		entries:   make(map[string]int64),
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

func (d *dictionary) getBuffer(key string) (*bytes.Buffer, bool) {
	reader := NewReader(d.target)
	defer reader.Close()

	if offset, ok := d.entries[key]; ok {
		_, buf := reader.fetchEntry(offset)
		return buf, true
	}

	pos := sort.SearchStrings(d.keys, key) - 1
	if pos == -1 || pos == len(d.keys)-1 {
		return nil, false
	}

	start := d.entries[d.keys[pos]]
	end := d.entries[d.keys[pos+1]]

	if buf, ok := reader.findEntry(key, start, end); ok {
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

	numKeys := readers.ReadUint32(f)
	for i := uint32(0); i < numKeys; i++ {
		klen := readers.ReadUint32(f)

		kbuf := make([]byte, klen)
		f.Read(kbuf)

		offset := readers.ReadUint64(f)
		d.entries[string(kbuf)] = int64(offset)
	}
}

func (d *dictionary) calculateOffsets() {
	reader := NewReader(d.target)

	var remainingBytes int64
	var offset int64
	for {
		numBytes := int64(len(reader.currentKey)) + int64(reader.fetchDataLength()) + 8 // 8 bytes used for offsets
		remainingBytes -= numBytes
		if remainingBytes <= 0 {
			d.entries[reader.currentKey] = offset
			remainingBytes = d.blockSize
		}

		reader.skipData()
		reader.nextKey()
		offset += numBytes

		if reader.done {
			d.entries[reader.currentKey] = offset - numBytes
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
	parts := strings.Split(d.target, "\\.")
	return fmt.Sprintf("%v.dict", parts[0])
}
