package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/prologic/bitcask"
	"log"
	"net/url"
	"strings"
)

var db *bitcask.Bitcask

func init() {
	var err error
	db, err = bitcask.Open("./db.bitcask")
	if err != nil {
		log.Fatal("Could not open bitcask db: ", err)
	}

	// Load std providers into db
	s := bufio.NewScanner(bytes.NewReader(_escFSMustByte(useLocal, "/standard_hosts.txt")))
	for s.Scan() {
		err = addHost(strings.ToLower(s.Text()))
		if err != nil {
			log.Fatalf("Could not add host '%s': %s ", s.Text(), err)
		}
	}
}

type UnShortUrl struct {
	ShortUrl    url.URL `json:"short_url"`
	LongUrl     url.URL `json:"long_url"`
	Blacklisted bool    `json:"blacklisted"`
}

func getUrlFromDB(shortUrl *url.URL) (*UnShortUrl, error) {
	val, err := db.Get([]byte(shortUrl.String()))
	if err != nil {
		return nil, errors.Wrap(err, "Could not get url from db")
	}
	un := &UnShortUrl{}
	err = json.Unmarshal(val, un)
	if err != nil {
		return nil, errors.Wrap(err, "Could not unmarshal value from db")
	}
	return un, nil
}

func saveUrlToDB(url UnShortUrl) error {
	urlJson, err := json.Marshal(url)
	if err != nil {
		return errors.Wrap(err, "Could not marshal url to json")
	}
	err = db.Put([]byte(url.ShortUrl.String()), urlJson)
	if err != nil {
		return errors.Wrap(err, "Could not put url into db")
	}

	return nil
}

func getLinkCount() (int, error) {
	stats, err := db.Stats()
	if err != nil {
		return -1, errors.Wrap(err, "Could not get url from db")
	}
	return stats.Keys, nil
}

func addHost(host string) error {
	_, err := db.Get([]byte(host))
	if err == nil {
		return nil
	}

	err = db.Put([]byte(fmt.Sprintf("host_%s", host)), []byte(host))
	if err != nil {
		return errors.Wrap(err, "Could not put host into db")
	}

	return nil
}

func getHosts() ([]string, error) {
	hosts := make([]string, 0)
	err := db.Scan([]byte("host_"), func(key []byte) error {
		h := strings.TrimPrefix(string(key), "host_")
		if h == "" {
			return nil
		}
		hosts = append(hosts, h)
		return nil
	})
	return hosts, err
}
