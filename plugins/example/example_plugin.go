package main

import (
	"fmt"
	"strings"

	irc "github.com/thoj/go-ircevent"
	"mvdan.cc/xurls/v2"
)

var Matchers = []string{"ExampleMatcher"}

func Example(args []string, e *irc.Event) (string, bool) {
	return fmt.Sprintf("caller: %s, ags: %s", e.Nick, strings.Join(args, ",")), false
}

func ExampleMatcher(msg string, e *irc.Event) (string, bool) {
	m := xurls.Strict()
	urls := m.FindAllString(msg, -1)
	if len(urls) == 0 {
		return "", false
	}
	return fmt.Sprintf("I found web address: %s", strings.Join(urls, ",")), false
}
