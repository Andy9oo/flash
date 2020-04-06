package indexer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

type doclist struct {
	path  string
	list  map[uint32]string
	limit int
}

func newDocList(root string, limit int) *doclist {
	path := fmt.Sprintf("%v/index.doclist", root)

	l := doclist{
		path:  path,
		list:  make(map[uint32]string),
		limit: limit,
	}

	return &l
}

func (d *doclist) add(id uint32, file string) {
	d.list[id] = file

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
	for id, path := range d.list {
		binary.Write(buf, binary.LittleEndian, id)
		binary.Write(buf, binary.LittleEndian, uint32(len(path)))
		binary.Write(buf, binary.LittleEndian, []byte(path))
	}

	buf.WriteTo(f)

	d.list = make(map[uint32]string)
}

func (d *doclist) fetchPath(id uint32) (path string, ok bool) {
	// Check if the id is in memory
	if val, ok := d.list[id]; ok {
		return val, true
	}

	f, err := os.Open(d.path)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	found := false
	for !found {
		// Read in id
		buf := make([]byte, 4)
		n, err := f.Read(buf)
		if n == 0 || err != nil {
			return "", false
		}
		currentID := binary.LittleEndian.Uint32(buf)

		// Read in path length
		buf = make([]byte, 4)
		f.Read(buf)
		pathLength := binary.LittleEndian.Uint32(buf)

		if currentID == id {
			// Read in path
			buf = make([]byte, pathLength)
			f.Read(buf)
			path = string(buf)
			found = true
		} else {
			f.Seek(int64(pathLength), os.SEEK_CUR)
		}
	}
	return path, found
}
