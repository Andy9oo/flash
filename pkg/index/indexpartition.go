package index

import (
	"bytes"
	"encoding/binary"
	"flash/pkg/index/postinglist"
)

type indexPartition struct {
	data map[string]*postinglist.List
}

type postingEntry struct {
	docID  uint32
	offset uint32
}

func newIndexPartition() *indexPartition {
	ip := indexPartition{
		data: make(map[string]*postinglist.List),
	}

	return &ip
}

func (ip *indexPartition) add(term string, entry partitionEntry) {
	switch entry.(type) {
	case *postingEntry:
		e := entry.(*postingEntry)
		if _, ok := ip.data[term]; !ok {
			ip.data[term] = postinglist.NewList()
		}
		ip.data[term].Add(e.docID, e.offset)
	case *postinglist.List:
		ip.data[term] = entry.(*postinglist.List)
	}
}

func (ip *indexPartition) get(term string) (partitionEntry, bool) {
	if val, ok := ip.data[term]; ok {
		return val, true
	}
	return nil, false
}

func (ip *indexPartition) decode(buf *bytes.Buffer) partitionEntry {
	return postinglist.Decode(buf)
}

func (ip *indexPartition) empty() bool {
	return len(ip.data) == 0
}

func (ip *indexPartition) keys() []string {
	keys := make([]string, 0, len(ip.data))
	for k := range ip.data {
		keys = append(keys, k)
	}
	return keys
}

func (ip *indexPartition) clear() {
	ip.data = nil
}

func (pe *postingEntry) Bytes() *bytes.Buffer {
	var buf *bytes.Buffer
	binary.Write(buf, binary.LittleEndian, pe.docID)
	binary.Write(buf, binary.LittleEndian, pe.offset)
	return buf
}
