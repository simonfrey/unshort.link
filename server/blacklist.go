package main

import (
	"bufio"
	"bytes"
	"github.com/pkg/errors"
	"log"
	"os"
	"sort"
)

var blacklist []string

func init() {
	//Load blacklist
	loadBlacklist()
	//Sort blacklist
	sort.Strings(blacklist)
}

func hostIsInBlacklist(host string) bool {
	i := sort.Search(len(blacklist), func(i int) bool { return host <= blacklist[i] })
	if i < len(blacklist) && blacklist[i] == host {
		return true
	}
	return false
}

func loadBlacklist() {
	blacklist = make([]string, 0)
	var s *bufio.Scanner
	if _, err := os.Stat("blacklist.txt"); os.IsNotExist(err) {
		s = bufio.NewScanner(bytes.NewReader(_escFSMustByte(useLocal, "/blacklist.txt")))
	} else {
		r, err := os.Open("blacklist.txt")
		if err != nil {
			panic(errors.Wrap(err, "Could not open blacklist.txt"))
		}
		s = bufio.NewScanner(r)
	}
	for s.Scan() {
		if s.Text() == "" {
			continue
		}
		blacklist = append(blacklist, s.Text())
	}
	err := s.Err()
	if err != nil {
		log.Fatal(err)
	}

}
