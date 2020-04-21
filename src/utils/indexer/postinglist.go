package indexer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"
)

type postingList struct {
	postings map[uint32]*posting
	docs     []uint32
	sorted   bool
}

type posting struct {
	docID     uint32
	frequency uint32
	offsets   []uint32
}

func newPostingList() *postingList {
	l := postingList{
		postings: make(map[uint32]*posting),
	}
	return &l
}

func decodePostingList(buf []byte) *postingList {
	l := newPostingList()

	offset := 0
	for offset < len(buf) {
		id := binary.LittleEndian.Uint32(buf[offset : offset+4])
		offset += 4

		frequency := binary.LittleEndian.Uint32(buf[offset : offset+4])
		offset += 4

		for i := uint32(0); i < frequency; i++ {
			pos := binary.LittleEndian.Uint32(buf[offset : offset+4])
			l.add(id, pos)
			offset += 4
		}
	}
	return l
}

func (l *postingList) add(docID uint32, offsets ...uint32) {
	if _, ok := l.postings[docID]; !ok {
		l.postings[docID] = &posting{docID: docID}
		l.docs = append(l.docs, docID)
		l.sorted = false
	}

	p := l.postings[docID]
	p.addOffsets(offsets)
}

func (l *postingList) getDocs() []uint32 {
	if !l.sorted {
		sort.Slice(l.docs, func(i, j int) bool { return l.docs[i] < l.docs[j] })
		l.sorted = true
	}

	return l.docs
}

func (p *posting) addOffsets(offsets []uint32) {
	p.offsets = append(p.offsets, offsets...)
	p.frequency += uint32(len(offsets))
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

func (l *postingList) Bytes() *bytes.Buffer {
	buf := new(bytes.Buffer)
	docs := l.getDocs()
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
