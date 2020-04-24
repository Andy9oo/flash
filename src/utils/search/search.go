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
	results, terms, preaders := e.initQuery(query, k)

	for terms[0].ok {
		doc := terms[0].nextDoc
		score := 0.0
		for terms[0].nextDoc == doc {
			t := terms[0].value
			numDocs := preaders[t].NumDocs
			freq := terms[0].frequency

			score += e.index.Score(doc, numDocs, freq, 1.2, 0.75)

			terms[0].nextDoc, terms[0].frequency, _, terms[0].ok = preaders[t].NextPosting()
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

func (e *Engine) initQuery(query string, k int) (resultHeap, termHeap, map[string]*indexer.PostingReader) {
	var results resultHeap
	for i := 0; i < k; i++ {
		heap.Push(&results, Result{
			ID:    0,
			Score: 0,
		})
	}

	terms := strings.Split(query, " ")
	for i := range terms {
		terms[i] = strings.ToLower(terms[i])
	}

	preaders := make(map[string]*indexer.PostingReader)

	var theap termHeap
	for _, t := range terms {
		if _, ok := preaders[t]; !ok {
			if pr, ok := e.index.GetPostingReader(t); ok {
				preaders[t] = pr
				id, freq, _, ok := pr.NextPosting()

				heap.Push(&theap, term{
					value:     t,
					frequency: freq,
					nextDoc:   id,
					ok:        ok,
				})
			}
		}
	}

	return results, theap, preaders
}
