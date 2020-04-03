package irc

import (
	"strings"

	"github.com/adamhassel/bender/internal/factoids"
	irc "github.com/thoj/go-ircevent"
)

// HandleMessages is the function that intercepts channel (or private) messages and handles them
func HandleMessages(c *irc.Connection, e *irc.Event) {
	msg := e.Message()
	channel := e.Arguments[0]
	if strings.HasPrefix(msg, c.GetNick()) {
		c.Privmsg(channel, "You said "+msg)
	}

	if strings.HasPrefix(msg, "!! ") {
		reply := factoids.Store(msg)
		c.Privmsg(channel, reply)
	}

	if strings.HasPrefix(msg, "!? ") {
		reply, action := factoids.Lookup(msg)
		if action {
			c.Action(channel, reply)
			return
		}
		c.Privmsg(channel, reply)
	}
	if strings.HasPrefix(msg, "!listfacts ") {
		c.Privmsg(channel, "noone implemented this yet")
	}
}
