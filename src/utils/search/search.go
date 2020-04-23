package search

import (
	"container/heap"
	"flash/src/utils/indexer"
	"sort"
	"strings"
)

// Engine is the search engine datastructure
type Engine struct {
	index *indexer.Index
}

// Result type
type Result struct {
	ID    uint32
	Score float64
}

// NewEngine creates a search engine for the given index
func NewEngine(index *indexer.Index) *Engine {
	e := Engine{
		index: index,
	}

	return &e
}

// Search query
func (e *Engine) Search(query string, k int) []*Result {
	results, terms := e.initHeaps(query, k)

	for terms[0].ok {
		doc := terms[0].nextDoc
		score := 0.0
		for terms[0].nextDoc == doc {
			t := terms[0].value
			score += e.index.Score(t, doc, 1.2, 0.75)

			terms[0].nextDoc, terms[0].ok = e.index.NextDoc(t, doc)
			heap.Fix(&terms, 0)
		}

		if score > results[0].Score {
			results[0].ID = doc
			results[0].Score = score
			heap.Fix(&results, 0)
		}
	}

	var finalResults []*Result
	for i := range results {
		if results[i].Score != 0 {
			finalResults = append(finalResults, &results[i])
		}
	}

	sort.Slice(finalResults, func(i, j int) bool { return finalResults[i].Score > finalResults[j].Score })

	return finalResults
}

func (e *Engine) initHeaps(query string, k int) (resultHeap, termHeap) {
	var results resultHeap
	for i := 0; i < k; i++ {
		heap.Push(&results, Result{
			ID:    0,
			Score: 0,
		})
	}

	terms := strings.Split(query, " ")

	var theap termHeap
	for _, t := range terms {
		doc, _, ok := e.index.First(t)

		heap.Push(&theap, term{
			value:   t,
			nextDoc: doc,
			ok:      ok,
		})
	}

	return results, theap
}
