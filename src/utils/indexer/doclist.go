package indexer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

type doclist struct {
	path        string
	list        map[uint32]*document
	limit       int
	numDocs     uint32
	totalLength int
}

type document struct {
	path   string
	length uint32
}

func newDocList(root string, limit int) *doclist {
	path := fmt.Sprintf("%v/index.doclist", root)

	l := doclist{
		path:  path,
		list:  make(map[uint32]*document),
		limit: limit,
	}

	return &l
}

func (d *doclist) add(file string, length uint32) {
	d.list[d.numDocs] = &document{
		path:   file,
		length: length,
	}

	d.totalLength += int(length)
	d.numDocs++

	if len(d.list) >= d.limit {
		d.dump()
	}

}

func (d *doclist) dump() {
	if len(d.list) == 0 {
		return
	}

	f, err := os.OpenFile(d.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Could not open document list file")
	}

	defer f.Close()

	buf := new(bytes.Buffer)
	for id, doc := range d.list {
		binary.Write(buf, binary.LittleEndian, id)
		binary.Write(buf, binary.LittleEndian, doc.length)
		binary.Write(buf, binary.LittleEndian, uint32(len(doc.path)))
		binary.Write(buf, binary.LittleEndian, []byte(doc.path))
	}

	buf.WriteTo(f)
	d.list = make(map[uint32]*document)
}

func (d *doclist) fetchDoc(id uint32) (doc *document, ok bool) {
	// Check if the id is in memory
	if doc, ok := d.list[id]; ok {
		return doc, true
	}

	f, err := os.Open(d.path)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	for {
		// Read in id
		buf := make([]byte, 4)
		n, err := f.Read(buf)
		if n == 0 || err != nil {
			return nil, false
		}
		currentID := binary.LittleEndian.Uint32(buf)

		buf = make([]byte, 4)
		f.Read(buf)
		fileLength := binary.LittleEndian.Uint32(buf)

		// Read in path length
		buf = make([]byte, 4)
		f.Read(buf)
		pathLength := binary.LittleEndian.Uint32(buf)

		if currentID == id {
			// Read in path
			buf = make([]byte, pathLength)
			f.Read(buf)

			path := string(buf)

			doc = &document{
				path:   path,
				length: fileLength,
			}
			return doc, true
		}

		f.Seek(int64(pathLength), os.SEEK_CUR)
	}
}
