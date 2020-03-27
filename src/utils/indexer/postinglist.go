package indexer

import (
	"fmt"
)

type postingList struct {
	Head *postingGroup
	Tail *postingGroup
}

type postingGroup struct {
	Postings []posting
	Next     *postingGroup
}

type posting struct {
	DocID  int
	Offset int
}

func newPostingGroup() *postingGroup {
	pGroup := postingGroup{
		Postings: make([]posting, 0, 16),
		Next:     nil,
	}

	return &pGroup
}

func (l *postingList) Add(docID int, offset int) {

	p := posting{
		DocID:  docID,
		Offset: offset,
	}

	if l.Head == nil {
		pGroup := newPostingGroup()
		l.Head = pGroup
		l.Tail = pGroup
	}

	if cap(l.Tail.Postings) == 256 && len(l.Tail.Postings) == 256 {
		pGroup := newPostingGroup()
		l.Tail.Next = pGroup
		l.Tail = pGroup
	}

	l.Tail.Postings = append(l.Tail.Postings, p)
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
