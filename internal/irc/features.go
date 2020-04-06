package irc

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/adamhassel/bender/internal/factoids"
	"github.com/adamhassel/bender/internal/helpers"
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
	case "!beatme":
		// TODO: also do custom reason
		// TODO: detect if I have +o
		// TODO: refactor this a bit, it's kinda hacky
		list := make(chan string, 1)
		id := c.AddCallback("353", func(e *irc.Event) {
			list <- e.Message()
			defer close(list)
		})
		defer c.RemoveCallback("353", id)
		c.SendRaw("NAMES " + channel)
		var l string
		select {
		case l = <-list:
		case <-time.NewTicker(5 * time.Second).C:
			SendReply(c, channel, "didn't receive user list in time", false)
		}
		users := helpers.NewStringSet(strings.Split(strings.TrimSpace(l), " ")...)
		users.Delete(c.GetNick())
		kickme := users.Slice().Random()
		SendReply(c, channel, "I would have kicked "+kickme+" if I was mean", false)
		//c.Kick(kickme, channel, "Det har du sikkert fortjent.")
	}
}

func SendReply(c *irc.Connection, ch string, msg string, action bool) {
	if action {
		c.Action(ch, msg)
		return
	}
	c.Privmsg(ch, msg)
}
