package main

import (
	"fmt"
	"github.com/sergi/go-diff/diffmatchpatch"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestGoDiff(t *testing.T) {
	r1,_ := http.Get("https://duckduckgo.com/?q=dda&ia=web&t=ffab")
	r2,_ := http.Get("https://duckduckgo.com/?q=dda")

	s1, _ := ioutil.ReadAll(r1.Body)
	s2, _ := ioutil.ReadAll(r2.Body)

	dmp := diffmatchpatch.New()

	diffs := dmp.DiffMain(string(s1), string(s2), false)

	d := dmp.DiffLevenshtein(diffs)
	fmt.Println(1-(float64(d)/float64(len(s1))))
}
