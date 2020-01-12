package db

import (
	"bufio"
	"bytes"
	"database/sql/driver"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"log"
	"net/url"
	"os"
	"strings"
)

//go:generate go get -u github.com/programmfabrik/esc
//go:generate esc -private -local-prefix-cwd -pkg=db -o=static.go standard_hosts.txt

var db *sqlx.DB
var providerBlacklist map[string]bool

func init() {
	var err error
	db, err = sqlx.Connect("sqlite3", "file:link.db")
	if err != nil {
		log.Fatalln(err)
	}

	// Init tables
	initSQL := `
CREATE TABLE IF NOT EXISTS links (
  short_url text PRIMARY KEY,
  long_url text,
  blacklisted boolean
);
CREATE TABLE IF NOT EXISTS hosts (
  name   text PRIMARY KEY
);`

	_, err = db.Exec(initSQL)
	if err != nil {
		log.Fatalln(err)
	}

	// Load std providers into db
	s := bufio.NewScanner(bytes.NewReader(_escFSMustByte(false, "/standard_hosts.txt")))
	for s.Scan() {
		err = AddHost(strings.ToLower(s.Text()))
		if err != nil {
			log.Fatalf("Could not add host '%s': %s ", s.Text(), err)
		}
	}

	loadProviderBlacklist()
}

func loadProviderBlacklist() {
	providerBlacklist = map[string]bool{}
	var s *bufio.Scanner
	if _, err := os.Stat("provider_blacklist.txt"); os.IsNotExist(err) {
		s = bufio.NewScanner(strings.NewReader(`google.com
linkedin.com
twitter.com
facebook.com
unsplash.com`))
	} else {
		r, err := os.Open("provider_blacklist.txt")
		if err != nil {
			panic(errors.Wrap(err, "Could not open blacklist.txt"))
		}
		s = bufio.NewScanner(r)
	}
	for s.Scan() {
		if s.Text() == "" {
			continue
		}
		providerBlacklist[s.Text()] = true
		providerBlacklist["www."+s.Text()] = true
	}
	err := s.Err()
	if err != nil {
		log.Fatal(err)
	}

}

type DUrl struct{ url.URL }

func (u DUrl) Value() (driver.Value, error) {
	return u.String(), nil
}

func (u DUrl) MarshalJSON() ([]byte, error) {
	return []byte(u.String()), nil
}

func (u *DUrl) UnmarshalJSON(b []byte) error {
	uu, err := u.Parse(string(b))
	if err != nil {
		return err
	}

	*u = DUrl{*uu}
	return nil
}

func (u *DUrl) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	uu, err := u.Parse(value.(string))
	if err != nil {
		return err
	}

	*u = DUrl{*uu}
	return nil
}

type UnShortUrl struct {
	ShortUrl    DUrl `json:"short_url" db:"short_url"`
	LongUrl     DUrl `json:"long_url" db:"long_url"`
	Blacklisted bool `json:"blacklisted" db:"blacklisted"`
}
type Host struct {
	Name string `db:"name"`
}

func GetUrlFromDB(shortUrl *url.URL) (*UnShortUrl, error) {
	u := &UnShortUrl{}
	err := db.Get(u, "SELECT * FROM links WHERE short_url = ? LIMIT 1", shortUrl.String())
	if err != nil {
		logrus.Errorf("Could not GET url: %s", err)
	}
	return u, err
}

func SaveUrlToDB(url UnShortUrl) error {
	_, err := db.Exec("INSERT INTO links (short_url, long_url, blacklisted) VALUES (?, ?, ?)",
		url.ShortUrl, url.LongUrl, url.Blacklisted)
	if err != nil {
		logrus.Errorf("Could not save new url: %s", err)
	}
	return err
}

func GetLinkCount() (int, error) {
	d := 0
	err := db.Get(&d, "SELECT COUNT(*) FROM links")
	return d, err
}

func AddHost(host string) error {
	res, err := db.Query("SELECT * FROM hosts where name = ?", host)
	if err != nil {
		return errors.Wrap(err, "Could not select from db")
	}
	if res.Next() {
		// The host is already in the db
		return nil
	}

	_, err = db.Exec("INSERT INTO hosts (name) VALUES (?)", host)
	return err
}

func GetHosts() ([]string, error) {
	h := []Host{}
	err := db.Select(&h, "SELECT * FROM hosts")
	if err != nil {
		return nil, err
	}

	u := make([]string, 0, len(h))
	for _, v := range h {
		if _, ok := providerBlacklist[v.Name]; !ok {
			u = append(u, v.Name)
		}
	}
	return u, err
}
