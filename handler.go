package main

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"strings"
)

var schemeReplacer *strings.Replacer
var indexTemplate, errorTemplate, blacklistTemplate, showTemplate []byte

func init() {
	schemeReplacer = strings.NewReplacer("https:/", "https://", "http:/", "http://")
}

func HandleIndex(rw http.ResponseWriter, req *http.Request) {
	_, _ = fmt.Fprintf(rw, "%s", "index.html")
}

func HandleShowRedirectPage(rw http.ResponseWriter, req *http.Request, url *UnShortUrl) {
	_, _ = fmt.Fprintf(rw, "%s", "redirect.html")
}
func HandleShowBlacklistPage(rw http.ResponseWriter, req *http.Request, url *UnShortUrl) {
	_, _ = fmt.Fprintf(rw, "%s", "blacklist.html")
}

func HandleUnShort(rw http.ResponseWriter, req *http.Request, redirect bool) {
	baseUrl := strings.TrimPrefix(req.URL.String(), serveUrl)
	baseUrl = schemeReplacer.Replace(baseUrl)
	baseUrl = strings.TrimPrefix(baseUrl, "/")

	myUrl, err := url.Parse(baseUrl)
	if err != nil {
		_, _ = fmt.Fprintf(rw, "%s", errors.Wrapf(err, "Could not parse given url '%s'", baseUrl))
		return
	}

	//Check in DB
	endUrl, err := GetUrlFromDB(myUrl)
	if err != nil {
		endUrl, err = GetUrl(myUrl)
		if err != nil {
			_, _ = fmt.Fprintf(rw, "%s", errors.Wrapf(err, "Could not access url '%s'", baseUrl))
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
