package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"

	log "github.com/sirupsen/logrus"
	irc "github.com/thoj/go-ircevent"
	"github.com/valyala/fastjson"
	"mvdan.cc/xurls/v2"
)

type service int

const (
	unknown service = iota
	bitly
	tinyurl
)

func ParseService(s string) service {
	switch strings.ToLower(s) {
	case "bitly", "bit.ly":
		return bitly
	case "tinyurl", "tinyurl.com":
		return tinyurl
	default:
		return unknown
	}
}

const bitlyAPIUrl = "https://api-ssl.bitly.com/v4/shorten"
const tinyurlAPIUrl = "https://api.tinyurl.com/create"

var Matchers = []string{"UrlShort"}
var apikey string
var minlen int
var serv service

// UrlShort asks bit.ly to shorten any link in `msg` longer than `minlen`
func UrlShort(msg string, e *irc.Event) (string, bool) {
	if apikey == "" {
		return "bitly api key not set", false
	}
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
	serv = tinyurl // default to tinyurl
	s, ok := c["service"]
	if ok {
		ss, ok := s.(string)
		if !ok {
			return errors.New("invalid service format")
		}
		serv = ParseService(ss)
	}
	return nil
}

func shortenUrl(url string, service service) (string, error) {
	var req *http.Request
	var resultKey string
	switch service {
	case bitly:
		var err error
		req, err = bitlyShortUrl(url)
		if err != nil {
			return "", err
		}
		resultKey = "link"
	case tinyurl:
		var err error
		req, err = tinyURLShortUrl(url)
		if err != nil {
			return "", err
		}
		resultKey = "data.tiny_url"
	default:
		return "", errors.New("unknown service")
	}
	client := http.Client{}
	req.Header.Set("Authorization", "Bearer "+apikey)
	if dout, err := httputil.DumpRequest(req, true); err != nil {
		log.Debugf("Request: %s", string(dout))
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("got error reply from upstream: %s", res.Status)
	}
	var result []byte
	result, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return fastjson.GetString(result, resultKey), nil
}

func bitlyShortUrl(url string) (*http.Request, error) {
	body := fmt.Sprintf(` { 
		"domain" :   "bit.ly",
		"long_url" : %q
		}`, url)
	return http.NewRequest("POST", bitlyAPIUrl, bytes.NewBufferString(body))
}

func tinyURLShortUrl(url string) (*http.Request, error) {
	body := fmt.Sprintf(` { 
		"domain" :   "tiny.one",
		"url" : %q
		}`, url)
	return http.NewRequest("POST", tinyurlAPIUrl, bytes.NewBufferString(body))
}
