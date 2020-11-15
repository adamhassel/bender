package main

import (
	"fmt"
	"strings"

	irc "github.com/thoj/go-ircevent"
)

func Example(args []string, e *irc.Event) (string, bool) {
	return fmt.Sprintf("caller: %s, ags: %s", e.Nick, strings.Join(args, ",")), false
}
