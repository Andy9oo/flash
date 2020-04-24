package search

import (
	"flash/src/utils/indexer"
	"math"
)

type termHeap []term

type term struct {
	value     string
	frequency uint32
	nextDoc   uint32
	maxScore  float64
	ok        bool
}

func (h termHeap) Len() int           { return len(h) }
func (h termHeap) Less(i, j int) bool { return h[i].nextDoc < h[j].nextDoc }
func (h termHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *termHeap) Push(x interface{}) {
	*h = append(*h, x.(term))
}

func (h *termHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func calculateMaxScore(info *indexer.IndexInfo, numDocs uint32) float64 {
	return (k1 + 1) * math.Log(float64(info.NumDocs)/float64(numDocs))
}
