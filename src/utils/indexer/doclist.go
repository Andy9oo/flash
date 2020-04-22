package indexer

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
)

type doclist struct {
	path        string
	docs        []document
	docLimit    int
	totalDocs   uint32
	totalLength int
	infoPath    string
	offsets     map[uint32]int64
}

type document struct {
	id     uint32
	path   string
	length uint32
}

func newDocList(root string, limit int) *doclist {
	path := fmt.Sprintf("%v/index.doclist", root)
	infoPath := fmt.Sprintf("%v/doclist.info", root)

	l := doclist{
		path:     path,
		infoPath: infoPath,
		offsets:  make(map[uint32]int64),
		docLimit: limit,
	}

	return &l
}

func loadDocList(root string, limit int, blockSize int64) *doclist {
	l := newDocList(root, limit)
	l.loadInfo(blockSize)
	return l
}

func (d *doclist) add(file string, length uint32) {
	doc := document{
		id:     d.totalDocs,
		path:   file,
		length: length,
	}

	d.docs = append(d.docs, doc)
	if len(d.docs) >= d.docLimit {
		d.dumpFiles()
	}

	d.totalLength += int(length)
	d.totalDocs++
}

func (d *doclist) loadInfo(blockSize int64) {
	f, err := os.Open(d.infoPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	// Load info
	d.totalDocs = readInt32(reader)
	d.totalLength = int(readInt32(reader))
	numOffsets := readInt32(reader)

	// Load offsets
	for i := uint32(0); i < numOffsets; i++ {
		id := readInt32(reader)
		offset := readInt64(reader)
		d.offsets[id] = int64(offset)
	}
}

func (d *doclist) calculateOffsets(blockSize int64) {
	f, err := os.Open(d.path)
	if err != nil {
		log.Fatal("Could not open doclist file")
	}
	defer f.Close()

	var remainingBytes int64
	var offset int64
	for i := uint32(0); i < d.totalDocs; i++ {
		id := readInt32(f)
		_ = readInt32(f) // Read length
		pathLength := readInt32(f)

		// 12 bytes for id, length, and pathLength
		numBytes := int64(12 + pathLength)
		remainingBytes -= numBytes

		if remainingBytes <= 0 || i == d.totalDocs-1 {
			d.offsets[id] = offset
			remainingBytes = blockSize
		}

		f.Seek(int64(pathLength), os.SEEK_CUR)
		offset += numBytes
	}

	d.dumpInfo()
}

func (d *doclist) dumpFiles() {
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

func (d *doclist) dumpInfo() {
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

func (d *doclist) fetchDoc(id uint32) (doc *document, ok bool) {
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

	keys := make([]int, 0, len(d.offsets))
	for key := range d.offsets {
		keys = append(keys, int(key))
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	pos := sort.SearchInts(keys, int(id)) - 1

	if pos == len(keys)-1 {
		return nil, false
	}

	start := d.offsets[uint32(keys[pos])]
	end := d.offsets[uint32(keys[pos+1])]

	return d.findDoc(f, id, start, end)
}

func (d *doclist) findDoc(f *os.File, id uint32, start, end int64) (doc *document, ok bool) {
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

func (d *doclist) readDoc(reader io.Reader) *document {
	id := readInt32(reader)
	len := readInt32(reader)
	plen := readInt32(reader)

	pbuf := make([]byte, plen)
	reader.Read(pbuf)

	return &document{
		id:     id,
		path:   string(pbuf),
		length: len,
	}
}
