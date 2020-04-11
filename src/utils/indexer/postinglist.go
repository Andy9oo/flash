package indexer

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type postingList struct {
	head   *posting
	tail   *posting
	length uint32
}

type posting struct {
	docID     uint32
	frequency uint32
	offsets   []uint32
	next      *posting
}

func newPosting(docID uint32) *posting {
	return &posting{
		docID:     docID,
		frequency: 0,
		next:      nil,
	}
}

func getPostingList(buf []byte) *postingList {
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

func (l *postingList) add(docID uint32, offset uint32) {
	if l.head == nil {
		p := newPosting(docID)

		l.head = p
		l.tail = p
		l.length++
	}

	if l.tail.docID != docID {
		p := newPosting(docID)

		l.tail.next = p
		l.tail = p
		l.length++
	}

	l.tail.offsets = append(l.tail.offsets, offset)
	l.tail.frequency++
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
