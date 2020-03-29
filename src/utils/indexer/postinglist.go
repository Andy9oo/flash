package indexer

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

const postingSize = 8

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

func getPostingList(file string, offset int64) (string, *postingList) {
	f, err := os.Open(file)
	if err != nil {
		fmt.Printf("Could not open file: %v\n", file)
	}
	defer f.Close()

	f.Seek(offset, 0)

	reader := bufio.NewReader(f)
	term, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
	}

	term = strings.TrimRight(term, "\n")
	buf, err := reader.ReadBytes('\n')
	if err != nil {
		fmt.Println(err)
	}

	l := new(postingList)
	for i := 0; i*postingSize < len(buf)-1; i++ {
		start := i * postingSize
		middle := start + postingSize/2
		end := start + postingSize

		docID := binary.LittleEndian.Uint32(buf[start:middle])
		offset := binary.LittleEndian.Uint32(buf[middle:end])

		l.add(docID, offset)
	}

	return term, l
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

	if cap(l.Tail.Postings) == 256 && len(l.Tail.Postings) == 256 {
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
