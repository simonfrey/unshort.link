package main

import (
	"bufio"
	"bytes"
	"log"
	"sort"
)

var blacklist []string

func init() {
	//Load blacklist
	loadBlacklist()
	//Sort blacklist
	sort.Strings(blacklist)
}

func HostIsInBlacklist(host string) bool {
	i := sort.Search(len(blacklist), func(i int) bool { return host <= blacklist[i] })
	if i < len(blacklist) && blacklist[i] == host {
		return true
	}
	return false
}

func loadBlacklist() {
	blacklist = make([]string, 0)
	s := bufio.NewScanner(bytes.NewReader(_escFSMustByte(useLocal, "/blacklist.txt")))
	for s.Scan() {
		blacklist = append(blacklist, s.Text())
	}
	err := s.Err()
	if err != nil {
		log.Fatal(err)
	}

}
