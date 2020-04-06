package irc

import (
	"context"
	"fmt"
	"log"
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

	factoidconf, err := factoids.ParseConfFile(factoids.DefaultConfFile)
	if err != nil {
		log.Print(err)
	}
	ctx = factoidconf.Context(ctx)

	if strings.HasPrefix(msg, "!! ") {
		reply := factoids.Store(msg, e.Nick)
		c.Privmsg(channel, reply)
	}

	if strings.HasPrefix(msg, "!? ") {
		reply, action := factoids.Lookup(ctx, msg)
		SendReply(c, channel, reply, action)
	}

	switch msg {
	case "!random":
		reply, action := factoids.Lookup(ctx, factoids.RandomKey())
		SendReply(c, channel, reply, action)
	case "!coffee":
		reply, action := fmt.Sprintf("pours %s a cup of coffee, straight from the pot", e.Nick), true
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
