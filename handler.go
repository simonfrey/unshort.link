package main

import (
	"fmt"
	"github.com/pkg/errors"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var schemeReplacer *strings.Replacer

func init() {
	schemeReplacer = strings.NewReplacer("https:/", "https://", "http:/", "http://")
}

type TemplateVars struct {
	ServerUrl string
	ShortUrl  string
	LongUrl   string
	Error     string
}

func HandleIndex(rw http.ResponseWriter, req *http.Request) {
	err := renderTemplate(rw,
		append(
			_escFSMustByte(useLocal, "/static/index.html"),
			_escFSMustByte(useLocal, "/static/main.html")...,
		),
		TemplateVars{ServerUrl: serveUrl},
	)
	if err != nil {
		HandleError(rw, err)
		return
	}
}

func HandleShowRedirectPage(rw http.ResponseWriter, req *http.Request, url *UnShortUrl) {
	err := renderTemplate(rw,
		append(
			_escFSMustByte(useLocal, "/static/show.html"),
			_escFSMustByte(useLocal, "/static/main.html")...,
		),
		TemplateVars{LongUrl: url.LongUrl.String(), ShortUrl: url.ShortUrl.String()},
	)
	if err != nil {
		HandleError(rw, err)
		return
	}
}
func HandleShowBlacklistPage(rw http.ResponseWriter, req *http.Request, url *UnShortUrl) {
	err := renderTemplate(rw,
		append(
			_escFSMustByte(useLocal, "/static/blacklist.html"),
			_escFSMustByte(useLocal, "/static/main.html")...,
		),
		TemplateVars{LongUrl: url.LongUrl.String(), ShortUrl: url.ShortUrl.String()},
	)
	if err != nil {
		HandleError(rw, err)
		return
	}
}

func HandleError(rw http.ResponseWriter, err error) {
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

func HandleUnShort(rw http.ResponseWriter, req *http.Request, redirect bool) {
	baseUrl := strings.TrimPrefix(req.URL.String(), serveUrl)
	baseUrl = schemeReplacer.Replace(baseUrl)
	baseUrl = strings.TrimPrefix(baseUrl, "/")

	myUrl, err := url.Parse(baseUrl)
	if err != nil {
		HandleError(rw, err)
		return
	}

	//Check in DB
	endUrl, err := GetUrlFromDB(myUrl)
	if err != nil {
		endUrl, err = GetUrl(myUrl)
		if err != nil {
			HandleError(rw, err)
			return
		}

		// Check for blacklist
		if HostIsInBlacklist(endUrl.LongUrl.Host) {
			endUrl.Blacklisted = true
		}

		// Save to db
		err = SaveUrlToDB(*endUrl)
	}

	if endUrl.Blacklisted {
		HandleShowBlacklistPage(rw, req, endUrl)
		return
	}

	if !redirect {
		HandleShowRedirectPage(rw, req, endUrl)
		return
	}

	http.Redirect(rw, req, endUrl.LongUrl.String(), http.StatusPermanentRedirect)
}
