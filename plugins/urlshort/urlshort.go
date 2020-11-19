package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	irc "github.com/thoj/go-ircevent"
	"github.com/valyala/fastjson"
	"mvdan.cc/xurls/v2"
)

const bitlyAPIUrl = "https://api-ssl.bitly.com/v4/shorten"

var Matchers = []string{"UrlShort"}
var apikey string
var minlen int

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
		link, err := bitlyShortUrl(url)
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
	if !ok {
		// not configured? No min len
		return nil
	}
	minlen, ok = len.(int)
	if !ok {
		return errors.New("invalid minlen format")
	}
	return nil
}

func bitlyShortUrl(url string) (string, error) {
	body := fmt.Sprintf(` { 
		"domain" :   "bit.ly",
		"long_url" : %q
		}`, url)

	client := http.Client{}
	req, err := http.NewRequest("POST", bitlyAPIUrl, bytes.NewBufferString(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apikey)
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
	return fastjson.GetString(result, "link"), nil
}
