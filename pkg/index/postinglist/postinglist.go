package postinglist

import (
	"bytes"
	"encoding/binary"
	"flash/tools/readers"
	"sort"
)

// List type
type List struct {
	postings map[uint64]*Posting
	docs     []uint64
	sorted   bool
}

// Posting type
type Posting struct {
	docID     uint64
	frequency uint32
	offsets   []uint32
}

// NewList creates a new posting list
func NewList() *List {
	l := List{
		postings: make(map[uint64]*Posting),
	}
	return &l
}

// Decode creates a new posting list from a buffer
func Decode(buf *bytes.Buffer, invalidDocs map[uint64]bool) (*List, bool) {
	l := NewList()

	numDocs := readers.ReadUint32(buf)
	l.docs = make([]uint64, 0, numDocs)
	for buf.Len() > 0 {
		id := readers.ReadUint64(buf)

		valid := true
		if _, ok := invalidDocs[id]; ok {
			valid = false
		}

		frequency := readers.ReadUint32(buf)
		for i := uint32(0); i < frequency; i++ {
			pos := readers.ReadUint32(buf)
			if valid {
				l.Add(id, pos)
			}
		}
	}
	return l, len(l.docs) != 0
}

// Add adds the offsets to the entry for the given doc
func (l *List) Add(docID uint64, offsets ...uint32) {
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

// Delete removes the given doc from the postinglist
func (l *List) Delete(docID uint64) {
	if _, ok := l.postings[docID]; ok {
		delete(l.postings, docID)
		for i, doc := range l.GetDocs() {
			if doc == docID {
				l.docs[i] = l.docs[len(l.docs)-1]
				l.docs[len(l.docs)-1] = 0
				l.docs = l.docs[:len(l.docs)-1]
				l.sorted = false
			}
		}
	}
}

// Empty returns true if the posting list contains no postings
func (l *List) Empty() bool {
	return len(l.postings) == 0
}

func (p *Posting) addOffsets(offsets []uint32) {
	p.offsets = append(p.offsets, offsets...)
	p.frequency += uint32(len(offsets))
}

// GetDocs returns a list of documents in the postinglist
func (l *List) GetDocs() []uint64 {
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
