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
	Head   *postingGroup
	Tail   *postingGroup
	length uint32
}

type postingGroup struct {
	Postings []posting
	Next     *postingGroup
}

type posting struct {
	DocID  uint32
	Offset uint32
}

func newPostingGroup() *postingGroup {
	pGroup := postingGroup{
		Postings: make([]posting, 0, 16),
		Next:     nil,
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
		DocID:  docID,
		Offset: offset,
	}

	if l.Head == nil {
		pGroup := newPostingGroup()
		l.Head = pGroup
		l.Tail = pGroup
	}

	if cap(l.Tail.Postings) == postingLimit && len(l.Tail.Postings) == postingLimit {
		pGroup := newPostingGroup()
		l.Tail.Next = pGroup
		l.Tail = pGroup
	}

	l.Tail.Postings = append(l.Tail.Postings, p)
	l.length++
}

func (l *postingList) String() string {
	var out string

	for g := l.Head; g != nil; g = g.Next {
		for i := 0; i < len(g.Postings); i++ {
			posting := g.Postings[i]
			out += fmt.Sprintf("%v;%v", posting.DocID, posting.Offset)
			if i != len(g.Postings)-1 || g.Next != nil {
				out += "|"
			}
		}
	}
	return out
}

func (l *postingList) Bytes() []byte {
	buf := new(bytes.Buffer)

	for g := l.Head; g != nil; g = g.Next {
		for i := 0; i < len(g.Postings); i++ {
			posting := g.Postings[i]

			binary.Write(buf, binary.LittleEndian, posting.DocID)
			binary.Write(buf, binary.LittleEndian, posting.Offset)
		}
	}

	return buf.Bytes()
}
