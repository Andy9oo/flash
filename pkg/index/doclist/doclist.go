package doclist

import (
	"bytes"
	"encoding/binary"
	"flash/tools/readers"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
)

// DocList type
type DocList struct {
	path        string
	docs        []Document
	docLimit    uint32
	totalDocs   uint32
	totalLength int
	infoPath    string
	offsets     map[uint32]int64
	keys        []int
}

// NewList creates a new doclist
func NewList(indexpath string, limit uint32) *DocList {
	path := fmt.Sprintf("%v/index.doclist", indexpath)
	infoPath := fmt.Sprintf("%v/doclist.info", indexpath)

	l := DocList{
		path:     path,
		infoPath: infoPath,
		offsets:  make(map[uint32]int64),
		docLimit: limit,
	}

	return &l
}

// Load loads a doclist for the given index
func Load(indexpath string, limit uint32) *DocList {
	l := NewList(indexpath, limit)
	l.loadInfo(int64(limit))
	return l
}

// Add adds the given file to the doclist
func (d *DocList) Add(file string, length uint32) {
	doc := Document{
		id:     d.totalDocs,
		path:   file,
		length: length,
	}

	d.docs = append(d.docs, doc)
	if uint32(len(d.docs)) >= d.docLimit {
		d.Dump()
	}

	d.totalLength += int(length)
	d.totalDocs++
}

// CalculateOffsets calculates the file offsets using the given blocksize
func (d *DocList) CalculateOffsets(blockSize int64) {
	f, err := os.Open(d.path)
	if err != nil {
		log.Fatal("Could not open doclist file")
	}
	defer f.Close()

	var remainingBytes int64
	var offset int64
	for i := uint32(0); i < d.totalDocs; i++ {
		id := readers.ReadUint32(f)
		_ = readers.ReadUint32(f) // Read length
		pathLength := readers.ReadUint32(f)

		// 12 bytes for id, length, and pathLength
		numBytes := int64(12 + pathLength)
		remainingBytes -= numBytes

		if remainingBytes <= 0 || i == d.totalDocs-1 {
			d.offsets[id] = offset
			d.keys = append(d.keys, int(id))
			remainingBytes = blockSize
		}

		f.Seek(int64(pathLength), os.SEEK_CUR)
		offset += numBytes
	}

	d.sortKeys()
	d.dumpInfo()
}

// Fetch gets the document with the given id
func (d *DocList) Fetch(id uint32) (doc *Document, ok bool) {
	f, err := os.Open(d.path)
	if err != nil {
		return nil, false
	}
	defer f.Close()

	if offset, ok := d.offsets[id]; ok {
		f.Seek(offset, os.SEEK_SET)
		doc := d.readDoc(f)
		return doc, true
	}

	pos := sort.SearchInts(d.keys, int(id)) - 1
	if pos == len(d.keys)-1 {
		return nil, false
	}

	start := d.offsets[uint32(d.keys[pos])]
	end := d.offsets[uint32(d.keys[pos+1])]

	return d.findDoc(f, id, start, end)
}

// Dump writes the current documents in the doclist to file
func (d *DocList) Dump() {
	if len(d.docs) == 0 {
		return
	}

	f, err := os.OpenFile(d.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Could not open document list file")
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	for _, doc := range d.docs {
		binary.Write(buf, binary.LittleEndian, doc.id)
		binary.Write(buf, binary.LittleEndian, doc.length)
		binary.Write(buf, binary.LittleEndian, uint32(len(doc.path)))
		binary.Write(buf, binary.LittleEndian, []byte(doc.path))
	}
	buf.WriteTo(f)

	d.docs = d.docs[:0]
}

// GetID returns a unique id in the doclist
func (d *DocList) GetID() uint32 {
	return d.totalDocs
}

// TotalLength returns the total length of all the documents added to the doclist
func (d *DocList) TotalLength() int {
	return d.totalLength
}

// NumDocs returns the total number of documents added to the doclist
func (d *DocList) NumDocs() uint32 {
	return d.totalDocs
}

func (d *DocList) loadInfo(blockSize int64) {
	f, err := os.Open(d.infoPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	// Load info
	d.totalDocs = readers.ReadUint32(f)
	d.totalLength = int(readers.ReadUint32(f))
	numOffsets := readers.ReadUint32(f)

	// Load offsets
	for i := uint32(0); i < numOffsets; i++ {
		id := readers.ReadUint32(f)
		offset := readers.ReadUint64(f)
		d.offsets[id] = int64(offset)
		d.keys = append(d.keys, int(id))
	}
	d.sortKeys()
}

func (d *DocList) dumpInfo() {
	f, err := os.Create(d.infoPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, d.totalDocs)
	binary.Write(buf, binary.LittleEndian, uint32(d.totalLength))
	binary.Write(buf, binary.LittleEndian, uint32(len(d.offsets)))
	for id, offset := range d.offsets {
		binary.Write(buf, binary.LittleEndian, id)
		binary.Write(buf, binary.LittleEndian, uint64(offset))
	}

	buf.WriteTo(f)
}

func (d *DocList) findDoc(f *os.File, id uint32, start, end int64) (doc *Document, ok bool) {
	f.Seek(start, os.SEEK_SET)

	blockSize := end - start
	block := make([]byte, blockSize)
	f.Read(block)

	buf := bytes.NewBuffer(block)
	for buf.Len() > 0 {
		doc := d.readDoc(buf)
		if doc.id == id {
			return doc, true
		}
	}

	return nil, false
}

func (d *DocList) readDoc(reader io.Reader) *Document {
	id := readers.ReadUint32(reader)
	len := readers.ReadUint32(reader)
	plen := readers.ReadUint32(reader)

	pbuf := make([]byte, plen)
	reader.Read(pbuf)

	return &Document{
		id:     id,
		path:   string(pbuf),
		length: len,
	}
}

func (d *DocList) sortKeys() {
	sort.Slice(d.keys, func(i, j int) bool { return d.keys[i] < d.keys[j] })
}
