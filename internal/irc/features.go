package irc

import (
	"context"
	"strings"

	"github.com/adamhassel/bender/internal/factoids"
	irc "github.com/thoj/go-ircevent"
)

// HandleMessages is the function that intercepts channel (or private) messages and handles them
func HandleMessages(ctx context.Context, c *irc.Connection, e *irc.Event) {
	msg := e.Message()
	channel := e.Arguments[0]
	if strings.HasPrefix(msg, c.GetNick()) {
		c.Privmsg(channel, "You said "+msg)
	}

	if strings.HasPrefix(msg, "!! ") {
		reply := factoids.Store(msg, e.Nick)
		c.Privmsg(channel, reply)
	}

	if strings.HasPrefix(msg, "!? ") {
		reply, action := factoids.Lookup(msg)
		SendReply(c, channel, reply, action)
	}
	if msg == "!random" {
		reply, action := factoids.Lookup(factoids.RandomKey())
		SendReply(c, channel, reply, action)
	}
}

func SendReply(c *irc.Connection, ch string, msg string, action bool) {
	if action {
		c.Action(ch, msg)
		return
	}
	c.Privmsg(ch, msg)
}
