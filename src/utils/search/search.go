package search

import (
	"container/heap"
	"flash/src/utils/indexer"
	"math"
	"sort"
	"strings"
)

// Engine is the search engine datastructure
type Engine struct {
	index *indexer.Index
	info  *indexer.IndexInfo
}

// Result type
type Result struct {
	ID    uint32
	Score float64
}

const (
	k1 float64 = 1.2
	b  float64 = 0.75
)

// NewEngine creates a search engine for the given index
func NewEngine(index *indexer.Index) *Engine {
	e := Engine{
		index: index,
		info:  index.GetInfo(),
	}

	return &e
}

// Search query
func (e *Engine) Search(query string, k int) []*Result {
	results, terms, preaders := e.initQuery(query, k)
	var removedTerms []term
	var removedScore float64

	for terms[0].ok {
		doc := terms[0].nextDoc
		score := 0.0
		for terms[0].nextDoc == doc {
			t := terms[0].value
			numDocs := preaders[t].NumDocs
			freq := terms[0].frequency

			score += e.Score(doc, numDocs, freq, k1, b)
			score += e.calculateRemovedTermsScore(removedTerms, preaders, doc)

			terms[0].nextDoc, terms[0].frequency, _, terms[0].ok = preaders[t].NextPosting()
			heap.Fix(&terms, 0)
		}

		if score > results[0].Score {
			results[0].ID = doc
			results[0].Score = score
			heap.Fix(&results, 0)
		}

		// Remove terms which cannot contribute
		if results[0].Score > terms[0].maxScore+removedScore {
			removedTerms = append(removedTerms, terms[0])
			removedScore += terms[0].maxScore
			heap.Remove(&terms, 0)
			heap.Fix(&terms, 0)
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
	for i := range terms {
		if _, ok := preaders[terms[i]]; !ok {
			if pr, ok := e.index.GetPostingReader(terms[i]); ok {
				preaders[terms[i]] = pr
				id, freq, _, ok := pr.NextPosting()

				t := term{
					value:     terms[i],
					frequency: freq,
					nextDoc:   id,
					maxScore:  calculateMaxScore(e.info, pr.NumDocs),
					ok:        ok,
				}

				heap.Push(&theap, t)
			}
		}
	}

	return results, theap, preaders
}

// Score returns the score for a doc using the BM25 ranking function
func (e *Engine) Score(doc, numDocs, frequency uint32, k float64, b float64) float64 {
	N := float64(e.info.NumDocs)
	lavg := float64(e.info.TotalLength) / float64(e.info.NumDocs)

	if _, length, ok := e.index.GetDocInfo(doc); ok {
		Nt := float64(numDocs)
		f := float64(frequency)
		l := float64(length)

		TF := (f * (k + 1)) / (f + k*((1-b)+b*(l/lavg)))
		return math.Log(N/Nt) * TF
	}

	return 0
}

func (e *Engine) calculateRemovedTermsScore(terms []term, preaders map[string]*indexer.PostingReader, doc uint32) float64 {
	score := 0.0
	for t := range terms {
		reader := preaders[terms[t].value]

		for terms[t].nextDoc < doc && terms[t].ok {
			terms[t].nextDoc, terms[t].frequency, _, terms[t].ok = reader.NextPosting()
		}

		// If the doc contains the term
		if terms[t].nextDoc == doc {
			score += e.Score(doc, reader.NumDocs, terms[t].frequency, k1, b)
		}
	}
	return score
}
