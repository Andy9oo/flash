package postinglist

import (
	"bytes"
	"encoding/binary"
	"flash/tools/readers"
	"sort"
)

// List type
type List struct {
	postings map[uint32]*Posting
	docs     []uint32
	sorted   bool
}

// Posting type
type Posting struct {
	docID     uint32
	frequency uint32
	offsets   []uint32
}

// NewList creates a new posting list
func NewList() *List {
	l := List{
		postings: make(map[uint32]*Posting),
	}
	return &l
}

// Decode creates a new posting list from a buffer
func Decode(buf *bytes.Buffer) *List {
	l := NewList()

	numDocs := readers.ReadUint32(buf)
	l.docs = make([]uint32, 0, numDocs)
	for buf.Len() > 0 {
		id := readers.ReadUint32(buf)
		frequency := readers.ReadUint32(buf)
		for i := uint32(0); i < frequency; i++ {
			pos := readers.ReadUint32(buf)
			l.Add(id, pos)
		}
	}
	return l
}

// Add adds the offsets to the entry for the given doc
func (l *List) Add(docID uint32, offsets ...uint32) {
	var p *Posting
	var ok bool

	if p, ok = l.postings[docID]; !ok {
		p = &Posting{docID: docID}
		l.postings[docID] = p
		l.docs = append(l.docs, docID)
		l.sorted = false
	}

	p.addOffsets(offsets)
}

func (p *Posting) addOffsets(offsets []uint32) {
	p.offsets = append(p.offsets, offsets...)
	p.frequency += uint32(len(offsets))
}

// GetDocs returns a list of documents in the postinglist
func (l *List) GetDocs() []uint32 {
	if !l.sorted {
		sort.Slice(l.docs, func(i, j int) bool { return l.docs[i] < l.docs[j] })
		l.sorted = true
	}

	return l.docs
}

// Bytes gives the posting list as a byte buffer
func (l *List) Bytes() *bytes.Buffer {
	buf := new(bytes.Buffer)
	docs := l.GetDocs()
	binary.Write(buf, binary.LittleEndian, uint32(len(docs)))
	for _, id := range docs {
		p := l.postings[id]
		binary.Write(buf, binary.LittleEndian, p.docID)
		binary.Write(buf, binary.LittleEndian, p.frequency)
		for i := 0; i < len(p.offsets); i++ {
			binary.Write(buf, binary.LittleEndian, p.offsets[i])
		}
	}
	return buf
}
