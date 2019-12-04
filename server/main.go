package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

//go:generate curl -o blacklist.txt https://hosts.ubuntu101.co.za/domains.list
//go:generate go get -u github.com/programmfabrik/esc
//go:generate esc -private -local-prefix-cwd -pkg=main -o=static.go static/ blacklist.txt standard_hosts.txt

var serveUrl, port string
var useLocal bool

func init() {
	flag.BoolVar(&useLocal, "local", false, "Use assets from local filesystem")
	flag.StringVar(&serveUrl, "url", "http://localhost:8080", "The server url this server runs on. (Required for the frontend)")
	flag.StringVar(&port, "port", "8080", "Port to run the server on")

	flag.Parse()
}

func main() {


	http.Handle("/static/", http.FileServer(_escFS(useLocal)))
	handler := func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "" || req.URL.Path == "/" || req.URL.Path == "/d/" || req.URL.Path == "/d" {
			handleIndex(rw)
			return
		}
		if strings.HasPrefix(req.URL.Path, "/providers") {
			rw.Header().Set("Access-Control-Allow-Origin", "*")
			rw.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			rw.Header().Set("Access-Control-Allow-Headers", "*")
			handleProviders(rw)
			return
		}
		if strings.HasPrefix(req.URL.Path, "/api/") {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api")
			handleUnShort(rw, req, false, true)
			return
		}
		if strings.HasPrefix(req.URL.Path, "/d/") {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/d")
			handleUnShort(rw, req, false, false)
			return
		}
		handleUnShort(rw, req, true, false)
	}
	http.HandleFunc("/", handler)

	logrus.Infof("Run server on port '%s', with url '%s' and local assets is set to '%t'", port, serveUrl, useLocal)

	logrus.Fatal(http.ListenAndServe(":"+port, nil))
}
