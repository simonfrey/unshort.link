package main

import (
	"log"
	"net/http"
	"strings"
)

var serveUrl string
var useLocal = true
func init() {
	serveUrl = "http://localhost:8080/"
}

func main() {
	handler := func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			HandleIndex(rw, req)
			return
		}
		if strings.HasPrefix(req.URL.Path, "/d/") {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/d")
			HandleUnShort(rw, req, false)
			return
		}
		HandleUnShort(rw, req, true)
	}

	http.Handle("/static/", http.FileServer(_escFS(useLocal)))
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
