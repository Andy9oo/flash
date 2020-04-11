package indexer

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	postingSize  = 8
	postingLimit = 256
)

type postingList struct {
	head   *postingGroup
	tail   *postingGroup
	length uint32
}

type postingGroup struct {
	postings []posting
	next     *postingGroup
}

type posting struct {
	docID  uint32
	offset uint32
}

func newPostingGroup() *postingGroup {
	pGroup := postingGroup{
		postings: make([]posting, 0, 16),
		next:     nil,
	}

	return &pGroup
}

func getPostingList(buf []byte) *postingList {
	l := new(postingList)
	for i := 0; i*postingSize < len(buf)-1; i++ {
		start := i * postingSize
		middle := start + postingSize/2
		end := start + postingSize

		docID := binary.LittleEndian.Uint32(buf[start:middle])
		offset := binary.LittleEndian.Uint32(buf[middle:end])

		l.add(docID, offset)
	}

	return l
}

func (l *postingList) add(docID uint32, offset uint32) {
	p := posting{
		docID:  docID,
		offset: offset,
	}

	if l.head == nil {
		pGroup := newPostingGroup()
		l.head = pGroup
		l.tail = pGroup
	}

	if cap(l.tail.postings) == postingLimit && len(l.tail.postings) == postingLimit {
		pGroup := newPostingGroup()
		l.tail.next = pGroup
		l.tail = pGroup
	}

	l.tail.postings = append(l.tail.postings, p)
	l.length++
}

func (l *postingList) String() string {
	var out string

	for g := l.head; g != nil; g = g.next {
		for i := 0; i < len(g.postings); i++ {
			posting := g.postings[i]
			out += fmt.Sprintf("%v;%v", posting.docID, posting.offset)
			if i != len(g.postings)-1 || g.next != nil {
				out += "|"
			}
		}
	}
	return out
}

func (l *postingList) Bytes() []byte {
	buf := new(bytes.Buffer)

	for g := l.head; g != nil; g = g.next {
		for i := 0; i < len(g.postings); i++ {
			posting := g.postings[i]

			binary.Write(buf, binary.LittleEndian, posting.docID)
			binary.Write(buf, binary.LittleEndian, posting.offset)
		}
	}

	return buf.Bytes()
}
