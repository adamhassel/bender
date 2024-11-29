package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	url2 "net/url"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	irc "github.com/thoj/go-ircevent"
	"github.com/valyala/fastjson"
	"mvdan.cc/xurls/v2"

	"github.com/adamhassel/bender/internal/helpers"
)

type service int

const (
	unknown service = iota
	bitly
	tinyurl
	cleanuri
	isgd
	shortio
)

func ParseService(s string) service {
	switch strings.ToLower(s) {
	case "bitly", "bit.ly":
		return bitly
	case "tinyurl", "tinyurl.com":
		return tinyurl
	case "clean", "cleanuri":
		return cleanuri
	case "is.gd", "isgd":
		return isgd
	case "short.io", "shortio":
		return shortio
	default:
		return unknown
	}
}

func authString(s service, token string) string {
	switch s {
	case shortio:
		return token
	}
	return "Bearer " + token
}

const bitlyAPIUrl = "https://api-ssl.bitly.com/v4/shorten"
const tinyurlAPIUrl = "https://api.tinyurl.com/create"
const cleanuriAPIUrl = "https://cleanuri.com/api/v1/shorten"
const isgdAPIUrl = "https://is.gd/create.php"
const shortioAPIUrl = "https://api.short.io/links/public"
const cleanparamfile = "plugins/urlshort/tracking.json"

var Matchers = []string{"UrlShort"}

var ErrNoCustomDomain = errors.New("custom domain undefined")

var apikey, customDomain string
var cleanup bool
var minlen int
var serv service
var cleanlist helpers.Set[string]

// UrlShort asks a shortener to shorten any link in `msg` longer than `minlen`
func UrlShort(msg string, e *irc.Event) (string, bool) {
	m := xurls.Strict()
	urls := m.FindAllString(msg, -1)
	shorts := make([]string, 0, len(urls))
	for _, url := range urls {
		if len(url) < minlen {
			continue
		}
		link, err := shortenUrl(url, serv)
		if err != nil {
			log.Infof("error looking up url %q: %s", url, err)
			continue
		}
		shorts = append(shorts, link)
	}
	return strings.Join(shorts, " "), false
}

// Configure is called by the bot on startup to configure the plugin
func Configure(c map[interface{}]interface{}) error {
	if cd, ok := c["custom_domain"]; ok {
		customDomain, ok = cd.(string)
		if !ok {
			return errors.New("invalid custom_domain format")
		}
	}
	key, ok := c["apikey"]
	if !ok {
		return errors.New("no apikey found")
	}
	apikey, ok = key.(string)
	if !ok {
		return errors.New("invalid apikey format")
	}
	len, ok := c["minlen"]
	if ok { // not configured? No min len
		minlen, ok = len.(int)
	}
	if !ok {
		return errors.New("invalid minlen format")
	}
	serv = cleanuri // default to cleanuri
	s, ok := c["service"]
	if ok {
		ss, ok := s.(string)
		if !ok {
			return errors.New("invalid service format")
		}
		serv = ParseService(ss)
	}
	cleanup = true
	clean, ok := c["cleanup"]
	if ok {
		var isok bool
		cleanup, isok = clean.(bool)
		if !isok {
			return errors.New("invalid cleanup format")
		}
	}
	// load the list
	if cleanup {
		if err := loadCleanParams(cleanparamfile); err != nil {
			return errors.New("error loading cleanup parameters file")
		}
	}
	return nil
}

func shortenUrl(url string, service service) (string, error) {
	var req *http.Request
	var resultKey []string
	if cleanup {
		var err error
		url, err = cleanURL(url)
		if err != nil {
			log.Errorf("error cleaning url: %s", err)
		}

	}
	switch service {
	case bitly:
		var err error
		req, err = bitlyShortUrl(url)
		if err != nil {
			return "", err
		}
		resultKey = []string{"link"}
	case tinyurl:
		var err error
		req, err = tinyURLShortUrl(url)
		if err != nil {
			return "", err
		}
		resultKey = []string{"data", "tiny_url"}
	case cleanuri:
		var err error
		req, err = cleanuriShortUrl(url)
		if err != nil {
			return "", err
		}
		resultKey = []string{"result_url"}
	case isgd:
		var err error
		req, err = isgdShortUrl(url)
		if err != nil {
			return "", err
		}
		resultKey = []string{"shorturl"}
	case shortio:
		var err error
		req, err = shortioShortUrl(url)
		if err != nil {
			return "", err
		}
		resultKey = []string{"shortURL"}
	default:
		return "", errors.New("unknown service")
	}
	client := http.Client{}
	if len(apikey) > 0 {

		req.Header.Set("Authorization", authString(service, apikey))
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if dout, err := httputil.DumpRequest(req, true); err == nil {
		log.Debugf("Request: %s", string(dout))
	} else {
		log.Debug(err)
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	var result []byte
	result, err = io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	log.Debug(string(result))
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("got error reply from upstream: %s, body %q", res.Status, string(result))
	}
	return fastjson.GetString(result, resultKey...), nil
}

func bitlyShortUrl(url string) (*http.Request, error) {
	cd := `"domain":"bit.ly",`
	if customDomain != "" {
		cd = `"domain":"` + customDomain + `",`
	}
	body := fmt.Sprintf(` {`+cd+`
		"long_url" : %q
		}`, url)
	return http.NewRequest("POST", bitlyAPIUrl, bytes.NewBufferString(body))
}

func shortioShortUrl(url string) (*http.Request, error) {
	if customDomain == "" {
		return nil, ErrNoCustomDomain
	}
	cd := `"domain":"` + customDomain + `",`
	body := fmt.Sprintf(` {`+cd+`
		"originalURL" : %q
		}`, url)
	return http.NewRequest("POST", shortioAPIUrl, bytes.NewBufferString(body))
}

func tinyURLShortUrl(url string) (*http.Request, error) {
	var cd string
	if customDomain != "" {
		cd = `"domain":"` + customDomain + `",`
	}
	body := fmt.Sprintf(` {`+cd+`
		"url" : %q
		}`, url)
	return http.NewRequest("POST", tinyurlAPIUrl, bytes.NewBufferString(body))
}

func cleanuriShortUrl(url string) (*http.Request, error) {
	body := fmt.Sprintf(` { "url" : %q }`, url)
	return http.NewRequest("POST", cleanuriAPIUrl, bytes.NewBufferString(body))
}

func isgdShortUrl(url string) (*http.Request, error) {
	body := fmt.Sprintf("format=json&url=%s", url2.QueryEscape(url))
	log.Info("yay")
	req, err := http.NewRequest("POST", isgdAPIUrl, bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

func loadCleanParams(filename string) error {
	jsonFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	raw, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}
	type cl struct {
		Name    string
		Company string
	}
	data := make([]cl, 0)
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}
	cleanlist = helpers.NewSet[string]()
	for _, d := range data {
		cleanlist.Add(d.Name)
	}
	return nil
}

func cleanURL(url string) (string, error) {
	u, err := url2.Parse(url)
	if err != nil {
		return "", err
	}

	v := u.Query()
	for k := range v {
		if cleanlist.Exists(k) {
			v.Del(k)
			log.Infof("removed tracking parameter %q from url", k)
		}
	}
	u.RawQuery = v.Encode()
	return u.String(), nil
}
