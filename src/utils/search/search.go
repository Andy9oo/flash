package search

import (
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

// Search uses the engine to search an index
func (e *Engine) Search(query string, k int) []*Result {
	terms := strings.Split(query, " ")

	var results []*Result
	doc, ok := e.getNextDoc(terms, 0, true)
	for ok {
		results = append(results, &Result{
			ID:    doc,
			Score: e.index.Score(terms, doc, 1.2, 0.75),
		})

		doc, ok = e.getNextDoc(terms, doc, false)
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })

	if len(results) < k {
		return results
	}

	return results[:k]
}

func (e *Engine) getNextDoc(terms []string, doc uint32, first bool) (selected uint32, ok bool) {
	docSelected := false
	var d uint32
	for i := 0; i < len(terms); i++ {
		if first {
			d, _, ok = e.index.First(terms[i])
		} else {
			d, ok = e.index.NextDoc(terms[i], doc)
		}

		if ok && (d < selected || !docSelected) {
			selected = d
			docSelected = true
		}
	}
	return selected, docSelected
}
