package indexer

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type partitionReader struct {
	file        *os.File
	scanner     *bufio.Scanner
	currentTerm string
	done        bool
}

func newPartitionReader(path string) *partitionReader {
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("Could not open file: %v\n", path)
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	scanner.Scan()

	return &partitionReader{
		file:        f,
		scanner:     scanner,
		currentTerm: scanner.Text(),
		done:        false,
	}
}

func (pr *partitionReader) compare(s string) int {
	return strings.Compare(pr.currentTerm, s)
}

func (pr *partitionReader) getPostings() []byte {
	pr.scanner.Scan()
	return pr.scanner.Bytes()
}

func (pr *partitionReader) advanceCurrentTerm() {
	if pr.scanner.Scan() {
		pr.currentTerm = pr.scanner.Text()
	} else {
		pr.done = true
		pr.file.Close()
	}
}
