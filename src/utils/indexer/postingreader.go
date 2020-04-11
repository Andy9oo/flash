package indexer

import (
	"bytes"
	"encoding/binary"
	"os"
)

type postingReader struct {
	file          *os.File
	currentDoc    uint32
	frequency     uint32
	postingLength uint32
	offset        uint32
	done          bool
}

func newPostingReader(file *os.File, postingLength uint32) *postingReader {
	reader := postingReader{
		file:          file,
		postingLength: postingLength,
		offset:        0,
		done:          false,
	}

	reader.fetchNextDocID()
	return &reader
}

func (r *postingReader) fetchNextDocID() (ok bool) {
	if r.offset+4 > r.postingLength {
		r.done = true
		return false
	}

	buf := make([]byte, 4)
	r.file.Read(buf)
	r.offset += 4

	r.currentDoc = binary.LittleEndian.Uint32(buf)
	return true
}

func (r *postingReader) fetchFrequency() uint32 {
	buf := make([]byte, 4)
	r.file.Read(buf)
	r.offset += 4

	r.frequency = binary.LittleEndian.Uint32(buf)
	return r.frequency
}

func (r *postingReader) fetchOffsets() []byte {
	buf := make([]byte, r.frequency*4)
	r.file.Read(buf)
	r.offset += r.frequency * 4

	return buf
}

func mergePostings(postingReaders []*postingReader, out *os.File) {
	var finished int
	var selectedReaders []*postingReader
	for finished < len(postingReaders) {
		var currentDoc uint32
		firstRun := true
		for _, r := range postingReaders {
			if r.done {
				continue
			}

			if r.currentDoc < currentDoc || firstRun {
				firstRun = false
				currentDoc = r.currentDoc
				selectedReaders = selectedReaders[:0]
				selectedReaders = append(selectedReaders, r)
			} else if r.currentDoc == currentDoc {
				selectedReaders = append(selectedReaders, r)
			}
		}

		var frequency uint32
		for i := range selectedReaders {
			frequency += selectedReaders[i].fetchFrequency()
		}

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, currentDoc)
		binary.Write(buf, binary.LittleEndian, frequency)
		buf.WriteTo(out)

		for i := range selectedReaders {
			out.Write(selectedReaders[i].fetchOffsets())
			if ok := selectedReaders[i].fetchNextDocID(); !ok {
				finished++
			}
		}
	}
}
