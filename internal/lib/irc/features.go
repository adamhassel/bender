package irc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	irc "github.com/thoj/go-ircevent"

	"github.com/adamhassel/bender/internal/factoids"
	"github.com/adamhassel/bender/internal/helpers"
	"github.com/adamhassel/bender/internal/lib/plugins"
)

// HandleMessages is the function that intercepts channel (or private) messages and handles them
func HandleMessages(ctx context.Context, c *irc.Connection, e *irc.Event) {
	msg := e.Message()
	channel := e.Arguments[0]

	factoidconf, err := factoids.ParseConfFile(factoids.DefaultConfFile)
	if err != nil {
		log.Error(err)
	}
	ctx = factoidconf.Context(ctx)

	// TODO: this structure is ugly
	command, err := ParseCommand(ctx, msg)
	if err != nil {
		if errors.Is(err, ErrNotCommand) {
			replies, err := plugins.Matchers(msg, e)
			if err != nil {
				log.Error(err)
				return
			}
			for _, r := range replies {
				SendReply(c, channel, r.Message, r.Action)
			}
		}
		return
	}

	switch command.Command {
	case "!":
		reply := factoids.Store(command.Argument, e.Nick)
		c.Privmsg(channel, reply)
	case "?":
		reply, action := factoids.Lookup(ctx, e.Nick, command.Argument)
		SendReply(c, channel, reply, action)
	case "random":
		reply, action := factoids.Lookup(ctx, e.Nick, factoids.RandomKey())
		SendReply(c, channel, reply, action)
	case "finfo":
		SendReply(c, channel, factoids.Lastfact().Info(), false)
	case "list":
		if command.Argument == "" {
			SendReply(c, channel, "You gotta tell me what to look for, bub", false)
			return
		}
		results, err := factoids.List(command.Argument)
		if err != nil {
			SendReply(c, channel, err.Error(), false)
			return
		}
		SendReply(c, channel, results, false)
	case "search":
		if command.Argument == "" {
			SendReply(c, channel, "You gotta tell me what to look for, bub", false)
			return
		}
		results, err := factoids.Search(command.Argument, 5)
		if err != nil {
			SendReply(c, channel, err.Error(), false)
			return
		}
		for _, s := range results {
			SendReply(c, channel, s, false)
			time.Sleep(200 * time.Millisecond)
		}
	case "coffee":
		reply, action := fmt.Sprintf("pours %s a cup of hot coffee, straight from the pot", e.Nick), true
		SendReply(c, channel, reply, action)
	case "buy": // this is the most used !bar feature from old bender, so it's implemented on its own.
		nick, item := splitBySpace(command.Argument)
		reply := fmt.Sprintf("gives %s a %s, \"Compliments of %s!\"", nick, item, e.Nick)
		SendReply(c, channel, reply, true)
	case "beatme":
		l, err := RequestReply(c, "353", "NAMES "+channel)
		if err != nil {
			SendReply(c, channel, fmt.Sprintf("Error getting user list: %s", err), false)
			return
		}
		users := helpers.NewStringSet(strings.Split(strings.TrimSpace(l), " ")...)

		// Can we kick anyone?
		if !users.Exists("@" + c.GetNick()) {
			SendReply(c, channel, "I am not a channel operator", false)
			return
		}

		// If it ain't friday in either CA or DK, kick whoever beatme'd.
		CA, _ := time.LoadLocation("America/Los_Angeles")
		DK, _ := time.LoadLocation("Europe/Copenhagen")
		if time.Now().In(CA).Weekday() != time.Friday && time.Now().In(DK).Weekday() != time.Friday {
			c.Kick(e.Nick, channel, "Det er ikke fredag, tåbe.")
			return
		}

		// Let's not kick ourself or the channel. Also, we have a '@' now, cause we're channel operator
		users.Delete("@" + c.GetNick())

		kickme := users.Slice().Random()
		if command.Argument == "" {
			command.Argument = "Det har du sikkert fortjent"
		}
		//SendReply(c, channel, fmt.Sprintf("I would have kicked %s if I were mean, while yelling %q", kickme, command.Argument), false)
		c.Kick(kickme, channel, command.Argument)
	default: // Check plugins
		r, err := plugins.Execute(command.Command, strings.Split(command.Argument, " "), e)
		if err != nil {
			log.Error(err)
			return
		}
		SendReply(c, channel, r.Message, r.Action)
	}
}

func SendReply(c *irc.Connection, ch string, msg string, action bool) {
	if action {
		c.Action(ch, msg)
		return
	}
	c.Privmsg(ch, msg)
}
