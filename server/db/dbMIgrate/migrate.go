package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/prologic/bitcask"
	"html"
	"log"
	db2 "unshort.link/db"
)

var db *bitcask.Bitcask
var providerBlacklist map[string]bool

func main() {
	var err error
	db, err = bitcask.Open("../db.bitcask")
	if err != nil {
		log.Fatal("Could not open bitcask db: ", err)
	}

	links,err := getLinks()
	if err != nil {
		log.Fatal("Could not get hosts: ", err)
	}

	fmt.Printf("INSERT INTO links (short_url, long_url, blacklisted) VALUES")
	for k, v := range links{
		if k != 0{
			fmt.Printf(",")
		}
		fmt.Printf(`("%s","%s",%t)`,html.EscapeString(v.ShortUrl.String()),html.EscapeString(v.LongUrl.String()),v.Blacklisted)
	}
	fmt.Printf(";")
}

func getLinks() ([]db2.UnShortUrl, error) {
	uu := []db2.UnShortUrl{}
	err := db.Scan([]byte("http"), func(key []byte) error {
		val, err := db.Get(key)
		if err != nil {
			return errors.Wrap(err, "Could not get val")
		}

		un := db2.UnShortUrl{}
		err = json.Unmarshal(val, &un)
		if err != nil {
			return errors.Wrap(err, "Could not unmarshal value from db")
		}

		uu = append(uu, un)
		return nil
	})
	return uu, err
}
