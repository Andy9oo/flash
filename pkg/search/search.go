package search

import (
	"container/heap"
	"flash/pkg/index"
	"flash/tools/text"
	"math"
	"sort"
	"strings"
)

// Engine is the search engine datastructure
type Engine struct {
	index    *index.Index
	info     *index.Info
	seenDocs map[uint64]uint32
}

// Result type
type Result struct {
	ID    uint64
	Score float64
}

const (
	k1 float64 = 1.2
	b  float64 = 0.75
)

// NewEngine creates a search engine for the given index
func NewEngine(index *index.Index) *Engine {
	e := Engine{
		index:    index,
		info:     index.GetInfo(),
		seenDocs: make(map[uint64]uint32),
	}

	return &e
}

// Search query
func (e *Engine) Search(query string, n int) []*Result {
	results, terms, treaders := e.initQuery(query, n)
	var removedTerms []term
	var removedScore float64

	if len(terms) == 0 {
		return nil
	}

	for len(terms) > 0 && terms[0].ok {
		doc := terms[0].nextDoc
		score := 0.0
		for terms[0].nextDoc == doc && terms[0].ok {
			t := terms[0].value
			numDocs := treaders[t].numDocs
			freq := terms[0].frequency

			score += e.Score(doc, numDocs, freq, k1, b)
			score += e.calculateRemovedTermsScore(removedTerms, treaders, doc)

			treaders[t].advanceDoc()
			terms[0].ok = !treaders[t].done()
			terms[0].nextDoc = treaders[t].nextDoc
			terms[0].frequency = treaders[t].frequency

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

func (e *Engine) initQuery(query string, n int) (resultHeap, termHeap, map[string]*termReader) {
	var results resultHeap
	for i := 0; i < n; i++ {
		heap.Push(&results, Result{
			ID:    0,
			Score: 0,
		})
	}

	terms := strings.Fields(query)
	for i := range terms {
		terms[i] = text.Normalize(terms[i])
	}

	treaders := make(map[string]*termReader)

	var theap termHeap
	for i := range terms {
		prs := e.index.GetPostingReaders(terms[i])
		if len(prs) == 0 {
			continue
		}

		tr := newTermReader(prs)
		treaders[terms[i]] = tr

		t := term{
			value:     terms[i],
			frequency: tr.frequency,
			nextDoc:   tr.nextDoc,
			maxScore:  calculateMaxScore(e.info, tr.numDocs),
			ok:        !tr.done(),
		}

		heap.Push(&theap, t)
	}

	return results, theap, treaders
}

// Score returns the score for a doc using the BM25 ranking function
func (e *Engine) Score(doc uint64, numDocs, frequency uint32, k float64, b float64) float64 {
	var docLength uint32
	lavg := e.info.AvgLength
	N := float64(e.info.NumDocs)

	if len, ok := e.seenDocs[doc]; ok {
		docLength = len
	} else if _, len, ok := e.index.GetDocInfo(doc); ok {
		docLength = len
		e.seenDocs[doc] = len
	} else {
		return 0
	}

	Nt := float64(numDocs)
	f := float64(frequency)
	l := float64(docLength)

	TF := (f * (k + 1)) / (f + k*((1-b)+b*(l/lavg)))
	return math.Log(N/Nt) * TF
}

func (e *Engine) calculateRemovedTermsScore(terms []term, treaders map[string]*termReader, doc uint64) float64 {
	score := 0.0
	for t := range terms {
		reader := treaders[terms[t].value]

		for terms[t].nextDoc < doc && terms[t].ok {
			reader.advanceDoc()

			terms[t].ok = !reader.done()
			terms[t].nextDoc = reader.nextDoc
			terms[t].frequency = reader.frequency
		}

		if terms[t].nextDoc == doc {
			score += e.Score(doc, reader.numDocs, terms[t].frequency, k1, b)
		}
	}
	return score
}
