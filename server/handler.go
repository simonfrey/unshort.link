package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unshort.link/db"
)

var schemeReplacer *strings.Replacer

func init() {
	schemeReplacer = strings.NewReplacer("https://", "https://", "http://", "http://", "https:/", "https://", "http:/", "http://")
}

type TemplateVars struct {
	ServerUrl    string
	ShortUrl     string
	FeedbackBody string
	LongUrl      string
	Error        string

	LinkCount int
}

type blacklistSource interface {
	IsBlacklisted(url string) bool
}

func handleIndex(rw http.ResponseWriter) {
	linkCount, err := db.GetLinkCount()
	if err != nil {
		handleError(rw, errors.Wrap(err, "Could not get link count"))
		return
	}

	err = renderTemplate(rw,
		append(
			_escFSMustByte(useLocal, "/static/index.html"),
			_escFSMustByte(useLocal, "/static/main.html")...,
		),
		TemplateVars{ServerUrl: serveUrl, LinkCount: linkCount},
	)
	if err != nil {
		handleError(rw, errors.Wrap(err, "Could not render template"))
		return
	}
}

func handleShowRedirectPage(rw http.ResponseWriter, u *db.UnShortUrl) {
	err := renderTemplate(rw,
		append(
			_escFSMustByte(useLocal, "/static/show.html"),
			_escFSMustByte(useLocal, "/static/main.html")...,
		),
		TemplateVars{LongUrl: u.LongUrl.String(),
			ShortUrl:     u.ShortUrl.String(),
			FeedbackBody: fmt.Sprintf("\n\n\n-----\nShort Url: %s\nLong Url: %s", u.ShortUrl.String(), u.LongUrl.String())},
	)
	if err != nil {
		handleError(rw, err)
		return
	}
}
func handleShowBlacklistPage(rw http.ResponseWriter, url *db.UnShortUrl) {
	err := renderTemplate(rw,
		append(
			_escFSMustByte(useLocal, "/static/blacklist.html"),
			_escFSMustByte(useLocal, "/static/main.html")...,
		),
		TemplateVars{LongUrl: url.LongUrl.String(), ShortUrl: url.ShortUrl.String()},
	)
	if err != nil {
		handleError(rw, err)
		return
	}
}

func handleError(rw http.ResponseWriter, err error) {
	rw.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(rw, "An error occured: %s", err)
}

func renderTemplate(rw io.Writer, templateBytes []byte, vars TemplateVars) error {
	var err error
	mainTemplate := template.New("main")
	mainTemplate, err = mainTemplate.Parse(string(templateBytes))
	if err != nil {
		return errors.Wrap(err, "Could not parse tempalte")
	}

	err = mainTemplate.Execute(rw, vars)
	if err != nil {
		return errors.Wrap(err, "Could not execute tempalte")
	}
	return nil
}

func handleUnShort(rw http.ResponseWriter, req *http.Request, redirect, api bool, blacklistSource blacklistSource) {
	baseUrl := strings.TrimPrefix(req.URL.String(), serveUrl)
	baseUrl = schemeReplacer.Replace(baseUrl)
	baseUrl = strings.TrimPrefix(baseUrl, "/")

	myUrl, err := url.Parse(baseUrl)
	if err != nil {
		handleError(rw, err)
		return
	}

	if myUrl.Scheme == "" {
		myUrl.Scheme = "http"
	}

	//Check in DB
	endUrl, err := db.GetUrlFromDB(myUrl)
	if err != nil {
		logrus.Infof("Get new url from short link: '%s'", myUrl.String())

		endUrl, err = getUrl(myUrl)
		if err != nil {
			handleError(rw, err)
			return
		}

		// Save to db
		err = db.SaveUrlToDB(*endUrl)
	}

	endUrl.Blacklisted = blacklistSource.IsBlacklisted(endUrl.LongUrl.Host)

	logrus.Infof("Access url: '%v'", endUrl)

	if api {
		jsoRes, err := json.Marshal(struct {
			ShortLink   string `json:"short_link"`
			LongLink    string `json:"long_link"`
			Blacklisted bool   `json:"blacklisted"`
		}{
			ShortLink:   endUrl.ShortUrl.String(),
			LongLink:    endUrl.LongUrl.String(),
			Blacklisted: endUrl.Blacklisted,
		})
		if err != nil {
			handleError(rw, errors.Wrap(err, "Could not marshal json"))
			return
		}
		_, _ = io.Copy(rw, bytes.NewReader(jsoRes))
		return
	}

	if endUrl.Blacklisted {
		handleShowBlacklistPage(rw, endUrl)
		return
	}

	if !redirect || endUrl.ShortUrl.String() == endUrl.LongUrl.String() {
		handleShowRedirectPage(rw, endUrl)
		return
	}

	http.Redirect(rw, req, endUrl.LongUrl.String(), http.StatusPermanentRedirect)
}

func handleProviders(rw http.ResponseWriter) {
	providers, err := db.GetHosts()
	if err != nil {
		handleError(rw, errors.Wrap(err, "Could not get hosts from db"))
	}
	providersJSON, err := json.MarshalIndent(providers, "", " ")
	if err != nil {
		handleError(rw, errors.Wrap(err, "Could not unmarshal standard hosts"))
	}
	_, _ = io.Copy(rw, bytes.NewReader(providersJSON))
}
