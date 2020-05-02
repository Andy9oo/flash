package index

import (
	"bytes"
	"encoding/binary"
	"flash/tools/readers"
	"fmt"
	"sort"
)

// PostingList type
type postingList struct {
	postings map[uint32]*posting
	docs     []uint32
	sorted   bool
}

// Posting type
type posting struct {
	docID     uint32
	frequency uint32
	offsets   []uint32
}

// NewPostingList creates a new posting list
func newPostingList() *postingList {
	l := postingList{
		postings: make(map[uint32]*posting),
	}
	return &l
}

func decodePostingList(buf *bytes.Buffer) *postingList {
	l := newPostingList()

	numDocs := readers.ReadUint32(buf)
	l.docs = make([]uint32, 0, numDocs)
	for buf.Len() > 0 {
		id := readers.ReadUint32(buf)
		frequency := readers.ReadUint32(buf)
		for i := uint32(0); i < frequency; i++ {
			pos := readers.ReadUint32(buf)
			l.add(id, pos)
		}
	}
	return l
}

func (l *postingList) add(docID uint32, offsets ...uint32) {
	var p *posting
	var ok bool

	if p, ok = l.postings[docID]; !ok {
		p = &posting{docID: docID}
		l.postings[docID] = p
		l.docs = append(l.docs, docID)
		l.sorted = false
	}

	p.addOffsets(offsets)
}

func (p *posting) addOffsets(offsets []uint32) {
	p.offsets = append(p.offsets, offsets...)
	p.frequency += uint32(len(offsets))
}

func (l *postingList) getDocs() []uint32 {
	if !l.sorted {
		sort.Slice(l.docs, func(i, j int) bool { return l.docs[i] < l.docs[j] })
		l.sorted = true
	}

	return l.docs
}

func (l *postingList) String() string {
	var out string

	docs := l.getDocs()
	for _, id := range docs {
		p := l.postings[id]
		out += fmt.Sprintf("(%v, %v, [", p.docID, p.frequency)
		for i := 0; i < len(p.offsets); i++ {
			out += fmt.Sprintf("%v", p.offsets[i])
			if i != len(p.offsets)-1 {
				out += ","
			}
		}
		out += "]) "
	}
	return out
}

// Bytes gives the posting list as a byte buffer
func (l *postingList) Bytes() *bytes.Buffer {
	buf := new(bytes.Buffer)
	docs := l.getDocs()
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
