package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	hClient           http.Client
	badParams         []string
	metaRedirectRegex *regexp.Regexp
)

func init() {
	hClient = http.Client{
		Timeout: 3 * time.Second,
	}

	metaRedirectRegex = regexp.MustCompile(`<.*?(?:(?:http-equiv="refresh".*?content=".*?(?:url|URL)=(.*?)")|(?:content=".*?(?:url|URL)=(.*?)".*?http-equiv="refresh")).*?>`)
	badParams = []string{"utm_source", "utm_medium", "utm_term", "utm_content", "utm_campaign", "utm_reader", "utm_place", "utm_userid", "utm_cid", "utm_name", "utm_pubreferrer", "utm_swu", "utm_viz_id", "ga_source", "ga_medium", "ga_term", "ga_content", "ga_campaign", "ga_place", "yclid", "_openstat", "fb_action_ids", "fb_action_types", "fb_ref", "fb_source", "action_object_map", "action_type_map", "action_ref_map", "gs_l", "pd_rd_@amazon.", "_encoding@amazon.", "psc@amazon.", "ved@google.", "ei@google.", "sei@google.", "gws_rd@google.", "cvid@bing.com", "form@bing.com", "sk@bing.com", "sp@bing.com", "sc@bing.com", "qs@bing.com", "pq@bing.com", "feature@youtube.com", "gclid@youtube.com", "kw@youtube.com", "$/ref@amazon.", "_hsenc", "mkt_tok", "hmb_campaign", "hmb_medium", "hmb_source", "source@sourceforge.net", "position@sourceforge.net", "callback@bilibili.com", "elqTrackId", "elqTrack", "assetType", "assetId", "recipientId", "campaignId", "siteId", "tag@amazon.", "ref_@amazon.", "pf_rd_@amazon.", "spm@.aliexpress.com", "scm@.aliexpress.com", "aff_platform", "aff_trace_key", "terminal_id", "_hsmi", "fbclid", "spReportId", "spJobID", "spUserID", "spMailingID", "utm_mailing", "utm_brand", "CNDID", "mbid", "trk", "trkCampaign", "sc_campaign", "sc_channel", "sc_content", "sc_medium", "sc_outcome", "sc_geo", "sc_country"}
}

func getUrl(inUrl *url.URL) (*UnShortUrl, error) {
	if !strings.HasPrefix(inUrl.Scheme, "http") {
		inUrl.Scheme = "http"
	}

	resp, baseBody, err := getWithRedirects(inUrl, 10)
	if err != nil {
		return nil, errors.Wrap(err, "Could not get base")
	}

	if resp.Request.URL.Host != inUrl.Host {
		// Redirect happened
		err = addHost(inUrl.Host)
		if err != nil {
			return nil, errors.Wrap(err, "Could not add new redirect host")
		}
	}

	queryParams := make([]string, 0)
	for k, _ := range resp.Request.URL.Query() {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", k, resp.Request.URL.Query().Get(k)))
	}

	// Remove known tracking parameter e.g. utm_source
	queryParams = removeKnownBadParams(queryParams)

	queryParamSet := combinations(queryParams)

	wg := sync.WaitGroup{}
	foundChan := make(chan string)
	breakCtx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	rateLimitChan := make(chan bool, 2)
	for k, parameters := range queryParamSet {
		if k >= 20 {
			break
		}

		rateLimitChan <- true
		wg.Add(1)
		go func(v []string) {
			defer func() {
				wg.Done()
				<-rateLimitChan
			}()
			tmQuery := ""
			for _, v := range v {
				if tmQuery == "" {
					tmQuery = tmQuery + v
					continue
				}
				tmQuery = tmQuery + "&" + v
			}

			tmpUrl := *inUrl
			tmpUrl.RawQuery = tmQuery

			req, err := http.NewRequest("GET", tmpUrl.String(), nil)
			if err != nil {
				logrus.Error(errors.Wrapf(err, "Could not create new request for url '%s'", tmpUrl.String()))
				return
			}
			req = req.WithContext(breakCtx)
			tmpResp, err := hClient.Do(req)
			if err != nil {
				logrus.Error(errors.Wrapf(err, "Could not get tmp url '%s'", tmpUrl.String()))
				return
			}

			tmpBody, err := ioutil.ReadAll(tmpResp.Body)
			if err != nil {
				logrus.Error(errors.Wrapf(err, "Could not read tmp body for url '%s'", tmpUrl.String()))
				return
			}

			if textEquality(string(baseBody), string(tmpBody)) > 0.75 {
				foundChan <- tmQuery
			}
		}(parameters)
	}

	waitChan := make(chan bool)
	go func() {
		wg.Wait()
		waitChan <- true
	}()

	rawQuery := ""
	select {
	case q := <-foundChan:
		rawQuery = q
	case <-waitChan:
	}
	cancelFunc()

	if rawQuery == "" {
		for _, v := range queryParams {
			if rawQuery == "" {
				rawQuery = rawQuery + v
				continue
			}

			rawQuery = rawQuery + "&" + v
		}
	}
	resp.Request.URL.RawQuery = rawQuery

	return &UnShortUrl{
		ShortUrl:    *inUrl,
		LongUrl:     *resp.Request.URL,
		Blacklisted: false,
	}, nil
}

func removeKnownBadParams(set []string) []string {
	cleaned := make([]string, 0, len(set))

	for _, v := range set {
		bad := false
		for _, reg := range badParams {
			if strings.Contains(v, reg) {
				bad = true
				break
			}
		}

		if !bad {
			cleaned = append(cleaned, v)
		}
	}

	return cleaned
}

//combinations is based on https://github.com/mxschmitt/golang-combinations/blob/master/combinations.go
func combinations(set []string) (subsets subsets) {
	length := uint(len(set))

	// Go through all possible combinations of objects
	// from 1 (only first object in subset) to 2^length (all objects in subset)
	for subsetBits := 1; subsetBits < (1 << length); subsetBits++ {
		var subset []string

		for object := uint(0); object < length; object++ {
			// checks if object is contained in subset
			// by checking if bit 'object' is set in subsetBits
			if (subsetBits>>object)&1 == 1 {
				// add object to subset
				subset = append(subset, set[object])
			}
		}
		// add subset to subsets
		subsets = append(subsets, subset)
	}

	subsets = append(subsets, []string{})
	sort.Sort(subsets)
	return subsets
}

type subsets [][]string

func (s subsets) Len() int           { return len(s) }
func (s subsets) Less(i, j int) bool { return len(s[i]) < len(s[j]) }
func (s subsets) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func getWithRedirects(inUrl *url.URL, maxTries int) (res *http.Response, body []byte, err error) {
	req, err := http.NewRequest("GET", inUrl.String(), nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Could not create http request")
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:69.0) Gecko/20100101 Firefox/69.0")

	resp, err := hClient.Do(req)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Could not get original url")
	}

	baseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Could not read original body")
	}

	m := metaRedirectRegex.FindSubmatch(baseBody)
	if len(m) == 3 {
		d := string(m[1])
		if d == "" {
			d = string(m[2])
		}
		u, err := url.ParseRequestURI(d)
		if err == nil && u.Scheme != "" && u.Host != "" && maxTries > 0 {
			return getWithRedirects(u, maxTries-1)
		}
	}

	return resp, baseBody, nil
}
