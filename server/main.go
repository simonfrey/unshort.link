package main

import (
	"database/sql"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
	"unshort.link/blacklist"
)

//go:generate go get -u github.com/programmfabrik/esc

var serveUrl, port string
var useLocal bool
var blacklistSyncInterval time.Duration
var blacklistUrls []string

func init() {
	flag.BoolVar(&useLocal, "local", false, "Use assets from local filesystem")
	flag.StringVar(&serveUrl, "url", "http://localhost:8080", "The server url this server runs on. (Required for the frontend)")
	flag.StringVar(&port, "port", "8080", "Port to run the server on")
	flag.DurationVar(&blacklistSyncInterval, "sync", time.Hour, "Blacklist synchronization interval")
	rawBlacklistUrls := flag.String("blacklist-sources", "https://hosts.ubuntu101.co.za/domains.list","Comma separated list of blacklist urls to periodically sync")
	flag.Parse()
	blacklistUrls = strings.Split(*rawBlacklistUrls, ",")
}

func main() {
	db, err := sql.Open("sqlite3", "file:blacklist.db")
	if err != nil {
		panic("Couldn't create database for host blacklist")
	}
	blacklistSource := blacklist.NewSqliteRepository(db)
	go blacklist.NewLoader(blacklistUrls, blacklistSource, blacklistSyncInterval).StartSync()


	handler := func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "" || req.URL.Path == "/" || req.URL.Path == "/d/" || req.URL.Path == "/d" {
			handleIndex(rw)
			return
		}
		if strings.HasPrefix(req.URL.Path, "favicon.ico") {
			rw.WriteHeader(http.StatusNotFound)
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
			handleUnShort(rw, req, false, true, blacklistSource)
			return
		}
		if strings.HasPrefix(req.URL.Path, "/d/") {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/d")
			handleUnShort(rw, req, false, false, blacklistSource)
			return
		}
		handleUnShort(rw, req, true, false, blacklistSource)
	}

	http.Handle("/static/", http.FileServer(_escFS(useLocal)))
	http.HandleFunc("/", handler)

	log.Infof("Run server on port '%s', with url '%s' and local assets is set to '%t'", port, serveUrl, useLocal)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
