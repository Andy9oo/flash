package doclist

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flash/pkg/index/partition"
	"flash/tools/readers"
	"fmt"
	"os"
)

// DocList type
type DocList struct {
	dir          string
	docCollector *partition.Collector
	idCollector  *partition.Collector
	totalDocs    uint32
	avgLength    float64
}

// NewList creates a new doclist
func NewList(indexpath string) *DocList {
	l := DocList{
		dir:          indexpath,
		docCollector: partition.NewCollector(indexpath, "doclist", NewDocPartition),
		idCollector:  partition.NewCollector(indexpath, "doclist.ids", NewIDPartition),
	}

	return &l
}

// Load loads a doclist for the given index
func Load(indexpath string) *DocList {
	l := NewList(indexpath)

	if l.docCollector.Load() == nil && l.idCollector.Load() == nil {
		l.loadStats()
	}

	return l
}

// Add adds the given file to the doclist
func (d *DocList) Add(id uint64, file string, length uint32) {
	doc := &Document{
		id:     id,
		path:   file,
		length: length,
	}
	d.docCollector.Add(fmt.Sprint(doc.id), doc)
	d.idCollector.Add(file, &ID{id})
	d.addLength(int(length))
	d.totalDocs++
}

// Delete removes a document from the doclist
func (d *DocList) Delete(id string, path string) {
	bufs, impls := d.docCollector.GetBuffers(id)
	totalLen := 0
	for i := range bufs {
		entry, _ := impls[i].Decode(id, bufs[i])
		if doc, ok := entry.(*Document); ok {
			totalLen += int(doc.length)
		}
	}

	if len(bufs) > 0 {
		d.removeLength(totalLen)
		d.totalDocs--
		d.docCollector.Delete(id)
	}
	d.idCollector.Delete(path)
}

// GetIDs returns the docIDs of the matching docs
func (d *DocList) GetIDs(path string) []*ID {
	var ids []*ID
	for _, entry := range d.idCollector.GetMatching(path) {
		if id, ok := entry.(*ID); ok {
			ids = append(ids, id)
		}
	}
	return ids
}

// FetchID gets the document with the given id
func (d *DocList) FetchID(id uint64) (doc *Document, ok bool) {
	entries := d.docCollector.GetEntries(fmt.Sprint(id))
	if len(entries) == 1 {
		if d, ok := entries[0].(*Document); ok {
			return d, true
		}
	}
	return nil, false
}

// FetchPath fetches a document using it's path
func (d *DocList) FetchPath(path string) (doc *Document, ok bool) {
	entries := d.idCollector.GetEntries(path)
	if len(entries) == 1 {
		if id, ok := entries[0].(*ID); ok {
			return d.FetchID(id.uint64)
		}
	}
	return nil, false
}

// AvgLength returns the average length of all the documents added to the doclist
func (d *DocList) AvgLength() float64 {
	return d.avgLength
}

func (d *DocList) addLength(val int) {
	d.avgLength = d.avgLength + (float64(val)-d.avgLength)/float64(d.totalDocs+1)
}

func (d *DocList) removeLength(val int) {
	d.avgLength = (d.avgLength*float64(d.totalDocs) - float64(val)) / float64(d.totalDocs-1)
}

// NumDocs returns the total number of documents added to the doclist
func (d *DocList) NumDocs() uint32 {
	return d.totalDocs
}

// ClearMemory writes any remaining partitions to disk
func (d *DocList) ClearMemory() {
	d.docCollector.ClearMemory()
	d.idCollector.ClearMemory()
	d.dumpStats()
}

func (d *DocList) dumpStats() {
	f, err := os.Create(fmt.Sprintf("%v/doclist.stats", d.dir))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, d.totalDocs)
	binary.Write(buf, binary.LittleEndian, d.avgLength)
	buf.WriteTo(f)
}

func (d *DocList) loadStats() {
	f, err := os.Open(fmt.Sprintf("%v/doclist.stats", d.dir))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	r := bufio.NewReader(f)
	d.totalDocs = readers.ReadUint32(r)
	d.avgLength = readers.ReadFloat64(r)
}
