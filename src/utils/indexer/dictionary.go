package indexer

import "fmt"

type dictionary struct {
	path      string
	blockSize int
	entries   map[string]uint64
}

func newDictionary(root string, blockSize int) *dictionary {
	path := fmt.Sprintf("%v/index.dict", root)

	d := dictionary{
		path:      path,
		blockSize: blockSize,
		entries:   make(map[string]uint64),
	}

	return &d
}
