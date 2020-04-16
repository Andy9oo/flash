package indexer

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type postingList struct {
	head *posting
	tail *posting
}

type posting struct {
	docID     uint32
	frequency uint32
	offsets   []uint32
	next      *posting
	prev      *posting
}

func decodePostingList(buf []byte) *postingList {
	l := new(postingList)

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
	if l.head == nil || docID < l.head.docID {
		p := &posting{docID: docID}
		p.addOffsets(offsets)
		l.push(p)
		return
	}

	for p := l.head; p != nil; p = p.next {
		if p.docID == docID {
			p.addOffsets(offsets)
			return
		}

		if p.next == nil || (p.docID < docID && docID < p.next.docID) {
			next := p.next
			current := &posting{docID: docID, next: next, prev: p}

			p.next = current
			current.addOffsets(offsets)
			if next == nil {
				l.tail = current
			} else {
				next.prev = current
			}
			return
		}
	}
}

func (l *postingList) push(p *posting) {
	if l.head == nil {
		l.head = p
		l.tail = p
	} else {
		temp := l.head.next
		p.next = temp
		temp.prev = p
		l.head = p
	}
}

func (p *posting) addOffsets(offsets []uint32) {
	p.offsets = append(p.offsets, offsets...)
	p.frequency += uint32(len(offsets))
}

func (l *postingList) String() string {
	var out string
	for p := l.head; p != nil; p = p.next {
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

func (l *postingList) Bytes() []byte {
	buf := new(bytes.Buffer)
	for p := l.head; p != nil; p = p.next {
		binary.Write(buf, binary.LittleEndian, p.docID)
		binary.Write(buf, binary.LittleEndian, p.frequency)
		for i := 0; i < len(p.offsets); i++ {
			binary.Write(buf, binary.LittleEndian, p.offsets[i])
		}
	}
	return buf.Bytes()
}
