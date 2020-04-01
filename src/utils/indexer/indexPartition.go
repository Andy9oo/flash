package indexer

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flash/src/utils/importer"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

type indexPartition struct {
	root            string
	partitionNumber uint32
	dictionary      map[string]*postingList
	memoryUsage     uint32
}

type partitionReader struct {
	file        *os.File
	scanner     *bufio.Scanner
	currentTerm string
	done        bool
}

// newIndexPartition creates a new index partition with the given partition number
func newIndexPartition(root string, partitionNumber uint32) *indexPartition {
	partition := indexPartition{
		root:            root,
		partitionNumber: partitionNumber,
		dictionary:      make(map[string]*postingList),
		memoryUsage:     0,
	}

	return &partition
}

func (p *indexPartition) addDoc(path string, docID uint32) {
	textChannel := importer.GetTextChannel(path)
	var offset uint32

	for term := range textChannel {
		term = strings.ToLower(term)

		pl := p.getPostingList(term)
		pl.add(docID, offset)

		p.memoryUsage += postingSize
		offset++

		if p.memoryUsage > partitionSizeLimit {
			p.writeToFile()
			p.reset(p.partitionNumber + 1)
		}
	}
}

func (p *indexPartition) getPostingList(term string) *postingList {
	var pl *postingList

	if val, ok := p.dictionary[term]; ok {
		pl = val
	} else {
		pl = new(postingList)
		p.dictionary[term] = pl
	}

	return pl
}

func (p *indexPartition) writeToFile() {
	path := fmt.Sprintf("%v/.index/p%d.index", p.root, p.partitionNumber)

	f, err := os.Create(path)
	if err != nil {
		log.Fatal("Could not create index partition")
	}
	defer f.Close()

	keys := make([]string, len(p.dictionary))

	count := 0
	for k := range p.dictionary {
		keys[count] = k
		count++
	}
	sort.Strings(keys)

	buf := new(bytes.Buffer)
	for _, key := range keys {
		binary.Write(buf, binary.LittleEndian, []byte(key))
		binary.Write(buf, binary.LittleEndian, []byte("\n"))
		binary.Write(buf, binary.LittleEndian, p.dictionary[key].Bytes())
		binary.Write(buf, binary.LittleEndian, []byte("\n"))
	}

	buf.WriteTo(f)
}

func (p *indexPartition) reset(partitionNumber uint32) {
	*p = *newIndexPartition(p.root, partitionNumber)
}

func mergePartitions(root string, numPartitions uint32) {
	path := fmt.Sprintf("%v/.index/final.index", root)

	f, err := os.Create(path)
	if err != nil {
		log.Fatal("Could not create index partition")
	}
	defer f.Close()

	readers := openPartitionReaders(root, numPartitions)

	currentIndex := readers[0]
	var currentTerm string
	var prevTerm string
	for currentIndex != nil {
		currentIndex = nil
		for i := uint32(0); i < numPartitions; i++ {
			if !readers[i].done {
				if currentIndex == nil || strings.Compare(readers[i].currentTerm, currentTerm) == -1 {
					currentIndex = readers[i]
					currentTerm = readers[i].currentTerm
				}
			}
		}

		if currentIndex != nil {
			currentIndex.scanner.Scan()
			postings := currentIndex.scanner.Bytes()

			buf := new(bytes.Buffer)

			if currentTerm != prevTerm {
				binary.Write(buf, binary.LittleEndian, []byte("\n"))
				binary.Write(buf, binary.LittleEndian, []byte(currentTerm))
				binary.Write(buf, binary.LittleEndian, []byte("\n"))
			}

			binary.Write(buf, binary.LittleEndian, postings)
			buf.WriteTo(f)

			if currentIndex.scanner.Scan() {
				currentIndex.currentTerm = currentIndex.scanner.Text()
			} else {
				currentIndex.done = true
				currentIndex.file.Close()
			}
		}
	}

	deletePartitionFiles(root, numPartitions)
}

func openPartitionReaders(root string, numPartitions uint32) []*partitionReader {
	readers := make([]*partitionReader, numPartitions)
	for i := uint32(0); i < numPartitions; i++ {
		partitionPath := fmt.Sprintf("%v/.index/p%d.index", root, i)

		temp, err := os.Open(partitionPath)
		if err != nil {
			fmt.Printf("Could not open file: %v\n", partitionPath)
			continue
		}

		scanner := bufio.NewScanner(temp)
		scanner.Split(bufio.ScanLines)
		scanner.Scan()

		readers[i] = &partitionReader{
			file:        temp,
			scanner:     scanner,
			currentTerm: scanner.Text(),
			done:        false,
		}
	}

	return readers
}

func deletePartitionFiles(root string, numPartitions uint32) {
	for i := uint32(0); i < numPartitions; i++ {
		partitionPath := fmt.Sprintf("%v/.index/p%d.index", root, i)
		os.Remove(partitionPath)
	}
}
