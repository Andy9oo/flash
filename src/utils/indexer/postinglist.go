package indexer

type postingList struct {
	Head   *posting
	Tail   *posting
	Length int
}

type posting struct {
	DocID  int
	Offset int
	Next   *posting
}

func (l *postingList) Add(docID int, offset int) {

	p := &posting{
		DocID:  docID,
		Offset: offset,
		Next:   nil,
	}

	if l.Length == 0 {
		l.Head = p
		l.Tail = p
	} else {
		l.Tail.Next = p
		l.Tail = p
	}

	l.Length++
}
