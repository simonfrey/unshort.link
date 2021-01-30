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
	ServerUrl        string
	ShortUrl         string
	FeedbackBody     string
	LongUrl          string
	Error            string
	DirectRedirect   bool
	BrowserExtension bool
	LinkCount        int
}

type blacklistSource interface {
	IsBlacklisted(url string) bool
}

func handleAbout(rw http.ResponseWriter, browserExtension bool) {
	renderLoading(rw)
	err := renderTemplate(rw,
		append(
			_escFSMustByte(useLocal, "/static/about.html"),
			_escFSMustByte(useLocal, "/static/main.html")...,
		),
		TemplateVars{ServerUrl: serveUrl, BrowserExtension: browserExtension},
	)
	if err != nil {
		handleError(rw, errors.Wrap(err, "Could not render template"), false)
		return
	}
}

func handleIndex(rw http.ResponseWriter, renderLoadingHTML bool) {
	if renderLoadingHTML {
		renderLoading(rw)
	}

	linkCount, err := db.GetLinkCount()
	if err != nil {
		handleError(rw, errors.Wrap(err, "Could not get link count"), false)
		return
	}
	linkCount = linkCount - (linkCount % 500)

	err = renderTemplate(rw,
		append(
			_escFSMustByte(useLocal, "/static/index.html"),
			_escFSMustByte(useLocal, "/static/main.html")...,
		),
		TemplateVars{ServerUrl: serveUrl, LinkCount: linkCount},
	)
	if err != nil {
		handleError(rw, errors.Wrap(err, "Could not render template"), false)
		return
	}
}

func handleShowRedirectPage(rw http.ResponseWriter, u *db.UnShortUrl, renderLoadingHTML, directRedirect bool) {
	if renderLoadingHTML {
		renderLoading(rw)
	}

	err := renderTemplate(rw,
		append(
			_escFSMustByte(useLocal, "/static/show.html"),
			_escFSMustByte(useLocal, "/static/main.html")...,
		),
		TemplateVars{DirectRedirect: directRedirect,
			LongUrl:      u.LongUrl.String(),
			ShortUrl:     u.ShortUrl.String(),
			FeedbackBody: fmt.Sprintf("\n\n\n-----\nShort Url: %s\nLong Url: %s", u.ShortUrl.String(), u.LongUrl.String())},
	)
	if err != nil {
		handleError(rw, err, false)
		return
	}
}
func handleShowBlacklistPage(rw http.ResponseWriter, url *db.UnShortUrl, renderLoadingHTML bool) {
	if renderLoadingHTML {
		renderLoading(rw)
	}

	err := renderTemplate(rw,
		append(
			_escFSMustByte(useLocal, "/static/blacklist.html"),
			_escFSMustByte(useLocal, "/static/main.html")...,
		),
		TemplateVars{LongUrl: url.LongUrl.String(), ShortUrl: url.ShortUrl.String()},
	)
	if err != nil {
		handleError(rw, err, false)
		return
	}
}

func renderLoading(rw http.ResponseWriter) {
	_, _ = io.Copy(rw, bytes.NewReader(_escFSMustByte(useLocal, "/static/loading.html")))
	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}
}

func handleApiError(rw http.ResponseWriter, err error) {
	rw.WriteHeader(http.StatusInternalServerError)
	_, _ = io.Copy(rw, strings.NewReader(err.Error()))
}

func handleError(rw http.ResponseWriter, err error, renderLoadingHTML bool) {
	if renderLoadingHTML {
		renderLoading(rw)
	}

	nErr := renderTemplate(rw,
		append(
			_escFSMustByte(useLocal, "/static/error.html"),
			_escFSMustByte(useLocal, "/static/main.html")...,
		),
		TemplateVars{Error: err.Error()},
	)
	if nErr != nil {
		_, _ = fmt.Fprintf(rw, "An error occured: %s", err)
	}
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

func handleUnShort(rw http.ResponseWriter, req *http.Request, redirect, api, checkBlacklist bool, blacklistSource blacklistSource) {
	if !api {
		renderLoading(rw)
	}

	baseUrl := strings.TrimPrefix(req.URL.String(), serveUrl)
	baseUrl = schemeReplacer.Replace(baseUrl)
	baseUrl = strings.TrimPrefix(baseUrl, "/")

	myUrl, err := url.Parse(baseUrl)
	if err != nil {
		handleError(rw, err, api)
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
			handleError(rw, err, api)
			return
		}

		// Save to db
		err = db.SaveUrlToDB(*endUrl)
		if err != nil {
			handleError(rw, err, api)
			return
		}
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
			handleApiError(rw, errors.Wrap(err, "Could not marshal json"))
			return
		}
		_, _ = io.Copy(rw, bytes.NewReader(jsoRes))
		return
	}

	if checkBlacklist && endUrl.Blacklisted {
		handleShowBlacklistPage(rw, endUrl, api)
		return
	}

	handleShowRedirectPage(rw, endUrl, api, redirect)
}

func handleProviders(rw http.ResponseWriter) {
	providers, err := db.GetHosts()
	if err != nil {
		handleError(rw, errors.Wrap(err, "Could not get hosts from db"), true)
	}
	providersJSON, err := json.MarshalIndent(providers, "", " ")
	if err != nil {
		handleError(rw, errors.Wrap(err, "Could not unmarshal standard hosts"), true)
	}
	_, _ = io.Copy(rw, bytes.NewReader(providersJSON))
}
