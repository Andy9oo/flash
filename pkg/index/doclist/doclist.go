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
	dir       string
	collector *partition.Collector
	totalDocs uint32
	avgLength float64
}

// NewList creates a new doclist
func NewList(indexpath string) *DocList {
	l := DocList{
		dir:       indexpath,
		collector: partition.NewCollector(indexpath, "doclist", NewPartition),
	}

	return &l
}

// Load loads a doclist for the given index
func Load(indexpath string) *DocList {
	l := NewList(indexpath)

	err := l.collector.Load()
	if err == nil {
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
	d.collector.Add(fmt.Sprint(doc.id), doc)
	d.addLength(int(length))
	d.totalDocs++
}

// Delete removes a document from the doclist
func (d *DocList) Delete(id string) {
	bufs, impls := d.collector.GetBuffers(id)
	totalLen := 0
	for i := range bufs {
		entry, _ := impls[i].Decode(bufs[i])
		if doc, ok := entry.(*Document); ok {
			totalLen += int(doc.length)
		}
	}

	if len(bufs) > 0 {
		d.removeLength(totalLen)
		d.totalDocs--
		d.collector.Delete(id)
	}
}

// GetIDs returns the docIDs of the matching docs
func (d *DocList) GetIDs(path string) []string {
	return d.collector.GetMatchingKeys(&Document{path: path})
}

// Fetch gets the document with the given id
func (d *DocList) Fetch(id uint64) (doc *Document, ok bool) {
	entries := d.collector.GetEntries(fmt.Sprint(id))
	if len(entries) == 1 {
		if d, ok := entries[0].(*Document); ok {
			return d, true
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
	d.collector.ClearMemory()
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
