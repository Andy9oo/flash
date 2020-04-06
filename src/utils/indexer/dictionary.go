package indexer

import (
	"fmt"
	"os"
)

type dictionary struct {
	path      string
	blockSize int64
	entries   map[string]int64
}

func loadDictionary(root string, blockSize int64) *dictionary {
	path := fmt.Sprintf("%v/index.dict", root)

	d := dictionary{
		path:      path,
		blockSize: blockSize,
		entries:   make(map[string]int64),
	}

	// d.fecthOffsets(root)
	d.dump()
	return &d
}

// func (d *dictionary) fecthOffsets(root string) {
// 	postingsFile := fmt.Sprintf("%v/index.postings", root)

// 	f, err := os.Open(postingsFile)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	defer f.Close()

// 	reader := bufio.NewReader(f)
// 	// scanner.Split(bufio.ScanLines)

// 	var remainingBytes int64
// 	var offset int64
// 	for {
// 		term, err := reader.ReadString(';')
// 		if err != nil {
// 			fmt.Println(err)
// 		}
// 		// term := scanner.Text()

// 		scanner.Scan()
// 		postings := scanner.Bytes()

// 		numBytes := int64(len(term) + len(postings))
// 		remainingBytes -= numBytes
// 		if remainingBytes <= 0 {
// 			d.entries[term] = offset
// 			remainingBytes = d.blockSize
// 		}
// 		offset += numBytes
// 	}
// }

func (d *dictionary) dump() {
	f, err := os.Create(d.path)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	for key, val := range d.entries {
		f.WriteString(key)
		f.WriteString(fmt.Sprintf(":%v\n", val))
	}

}
