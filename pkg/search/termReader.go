package search

import (
	"flash/pkg/index/postinglist"
)

type termReader struct {
	preaders        []*postinglist.Reader
	nextDoc         uint32
	frequency       uint32
	numDocs         uint32
	finishedReaders []bool
	finished        int
}

func newTermReader(preaders []*postinglist.Reader) *termReader {
	tr := termReader{
		preaders:        preaders,
		finishedReaders: make([]bool, len(preaders)),
	}

	for i, pr := range tr.preaders {
		if pr.Read() {
			id, freq, _ := pr.Data()

			if id < tr.nextDoc || i == 0 {
				tr.nextDoc = id
				tr.frequency = freq
			} else if id == tr.nextDoc {
				tr.frequency += freq
			}

			tr.numDocs += pr.NumDocs()
		} else {
			tr.finishedReaders[i] = true
			tr.finished++
		}
	}

	return &tr
}

func (tr *termReader) advanceDoc() {

	currentDoc := tr.nextDoc
	selected := false
	for i, pr := range tr.preaders {
		// Continue if the reader has already finsihed
		if tr.finishedReaders[i] {
			continue
		}

		// Advance a reader if it has the same term as the current doc
		if id, _, _ := pr.Data(); id == currentDoc {
			if ok := pr.Read(); !ok {
				tr.finishedReaders[i] = true
				tr.finished++
				continue
			}
		}

		id, freq, _ := pr.Data()

		if id < tr.nextDoc || !selected {
			tr.nextDoc = id
			tr.frequency = freq
			selected = true
		} else if id == tr.nextDoc {
			tr.frequency += freq
		}
	}
}

func (tr *termReader) done() bool {
	return tr.finished == len(tr.preaders)
}
