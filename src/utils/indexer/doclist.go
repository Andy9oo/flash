package indexer

import (
	"bufio"
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
		binary.Write(buf, binary.LittleEndian, []byte("\n"))
		binary.Write(buf, binary.LittleEndian, []byte(path))
		binary.Write(buf, binary.LittleEndian, []byte("\n"))
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

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		fmt.Println(scanner.Bytes())
		if binary.LittleEndian.Uint32(scanner.Bytes()) == id {
			scanner.Scan()
			return scanner.Text(), true
		}
		scanner.Scan()
	}

	return "", false
}
