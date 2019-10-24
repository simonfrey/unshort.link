package main

import (
	"bufio"
	"log"
	"os"
	"sort"
)


var blacklist []string

func init(){
	//Load blacklist
	loadBlacklist("blackweb.txt")
	//Sort blacklist
	sort.Strings(blacklist)
}


func HostIsInBlacklist(host string)bool{
	i := sort.Search(len(blacklist), func(i int) bool { return host <= blacklist[i] })
	if i < len(blacklist) && blacklist[i] == host {
		return true
	}
	return false
}

func loadBlacklist(filename string){
	blacklist = make([]string,0)
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	s := bufio.NewScanner(f)
	for s.Scan() {
		blacklist = append(blacklist,s.Text())
	}
	err = s.Err()
	if err != nil {
		log.Fatal(err)
	}

}
