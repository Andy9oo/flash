package indexer

import (
	"bytes"
	"encoding/binary"
	"flash/src/utils/importer"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

type partition struct {
	root            string
	partitionNumber uint32
	postings        map[string]*postingList
	memoryUsage     uint32
}

func newPartition(root string, partitionNumber uint32) *partition {
	p := partition{
		root:            root,
		partitionNumber: partitionNumber,
		postings:        make(map[string]*postingList),
	}

	return &p
}

func (p *partition) add(file string, docID uint32) {
	textChannel := importer.GetTextChannel(file)
	var offset uint32

	for term := range textChannel {
		term = strings.ToLower(term)

		pl := p.getPostingList(term)
		pl.add(docID, offset)

		p.memoryUsage += postingSize

		if p.memoryUsage >= partitionSizeLimit {
			p.dump()
			p.reset(p.partitionNumber + 1)
		}

		offset++
	}
}

func (p *partition) getPostingList(term string) *postingList {
	var pl *postingList

	if val, ok := p.postings[term]; ok {
		pl = val
	} else {
		pl = new(postingList)
		p.postings[term] = pl
	}

	return pl
}

func (p *partition) dump() {
	if p.memoryUsage == 0 {
		return
	}

	f, err := os.Create(p.getPath())
	if err != nil {
		log.Fatal("Could not create index partition")
	}

	defer f.Close()

	terms := make([]string, len(p.postings))

	count := 0
	for t := range p.postings {
		terms[count] = t
		count++
	}

	sort.Strings(terms)

	buf := new(bytes.Buffer)
	for _, term := range terms {
		binary.Write(buf, binary.LittleEndian, []byte(term))
		binary.Write(buf, binary.LittleEndian, []byte("\n"))
		binary.Write(buf, binary.LittleEndian, p.postings[term].Bytes())
		binary.Write(buf, binary.LittleEndian, []byte("\n"))
	}

	buf.WriteTo(f)
}

func (p *partition) reset(partitionNumber uint32) {
	p.partitionNumber = partitionNumber
	p.postings = make(map[string]*postingList)
	p.memoryUsage = 0
}

func (p *partition) getPath() string {
	return fmt.Sprintf("%v/.index/p%d.part", p.root, p.partitionNumber)
}
