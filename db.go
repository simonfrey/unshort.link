package main

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/prologic/bitcask"
	"log"
	"net/url"
)

var db *bitcask.Bitcask

func init() {
	var err error
	db, err = bitcask.Open("./db.bitcask")
	if err != nil {
		log.Fatal("Could not open bitcask db")
	}
}

type UnShortUrl struct {
	ShortUrl    url.URL
	LongUrl     url.URL
	Blacklisted bool
}

func GetUrlFromDB(shortUrl *url.URL) (*UnShortUrl, error) {
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

func SaveUrlToDB(url UnShortUrl) error {
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

func GetLinkCount() (int, error) {
	stats, err := db.Stats()
	if err != nil {
		return -1, errors.Wrap(err, "Could not get url from db")
	}
	return stats.Keys, nil
}
