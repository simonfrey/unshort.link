package main

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

var (
	hClient   http.Client
	badParams []string
)

func init() {
	hClient = http.Client{
		Timeout: 10 * time.Second,
	}

	badParams = []string{"utm_source", "utm_medium", "utm_term", "utm_content", "utm_campaign", "utm_reader", "utm_place", "utm_userid", "utm_cid", "utm_name", "utm_pubreferrer", "utm_swu", "utm_viz_id", "ga_source", "ga_medium", "ga_term", "ga_content", "ga_campaign", "ga_place", "yclid", "_openstat", "fb_action_ids", "fb_action_types", "fb_ref", "fb_source", "action_object_map", "action_type_map", "action_ref_map", "gs_l", "pd_rd_@amazon.", "_encoding@amazon.", "psc@amazon.", "ved@google.", "ei@google.", "sei@google.", "gws_rd@google.", "cvid@bing.com", "form@bing.com", "sk@bing.com", "sp@bing.com", "sc@bing.com", "qs@bing.com", "pq@bing.com", "feature@youtube.com", "gclid@youtube.com", "kw@youtube.com", "$/ref@amazon.", "_hsenc", "mkt_tok", "hmb_campaign", "hmb_medium", "hmb_source", "source@sourceforge.net", "position@sourceforge.net", "callback@bilibili.com", "elqTrackId", "elqTrack", "assetType", "assetId", "recipientId", "campaignId", "siteId", "tag@amazon.", "ref_@amazon.", "pf_rd_@amazon.", "spm@.aliexpress.com", "scm@.aliexpress.com", "aff_platform", "aff_trace_key", "terminal_id", "_hsmi", "fbclid", "spReportId", "spJobID", "spUserID", "spMailingID", "utm_mailing", "utm_brand", "CNDID", "mbid", "trk", "trkCampaign", "sc_campaign", "sc_channel", "sc_content", "sc_medium", "sc_outcome", "sc_geo", "sc_country"}
}

func GetUrl(inUrl *url.URL) (*UnShortUrl, error) {
	if !strings.HasPrefix(inUrl.Scheme, "http") {
		inUrl.Scheme = "http"
	}

	resp, err := hClient.Get(inUrl.String())
	if err != nil {
		return nil, errors.Wrap(err, "Could not get original url")
	}

	queryParams := make([]string, 0)
	for k, _ := range resp.Request.URL.Query() {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", k, resp.Request.URL.Query().Get(k)))
	}

	// Remove known tracking parameter e.g. utm_source
	queryParams = RemoveKnownBadParams(queryParams)

	baseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Could not read original body")
	}
	queryParamSet := Combinations(queryParams)

	rawQuery := ""
	found := false
	for _, v := range queryParamSet {
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

		tmpResp, err := hClient.Get(tmpUrl.String())
		if err != nil {
			return nil, errors.Wrap(err, "Could not get tmp url")
		}

		tmpBody, err := ioutil.ReadAll(tmpResp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "Could not read tmp body")
		}

		if TextEquality(string(baseBody), string(tmpBody)) > 0.6 {
			rawQuery = tmQuery
			found = true
			break
		}
	}

	if !found{
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

func RemoveKnownBadParams(set []string) []string {
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

//Combinations is based on https://github.com/mxschmitt/golang-combinations/blob/master/combinations.go
func Combinations(set []string) (subsets Subsets) {
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

	subsets = append(subsets,[]string{})
	sort.Sort(subsets)
	return subsets
}

type Subsets [][]string

func (s Subsets) Len() int           { return len(s) }
func (s Subsets) Less(i, j int) bool { return len(s[i]) < len(s[j]) }
func (s Subsets) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
