package main

import (
	"github.com/sergi/go-diff/diffmatchpatch"
)

// TextEquality gives percent of text equality
func TextEquality(doc1, doc2 string) float64 {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(doc1, doc2, false)
	d := dmp.DiffLevenshtein(diffs)
	return 1 - (float64(d) / float64(len(doc1)))
}
