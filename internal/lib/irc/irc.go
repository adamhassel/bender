package irc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/adamhassel/bender/internal/config"
	log "github.com/sirupsen/logrus"

	irc "github.com/thoj/go-ircevent"
)

func InitBot(ctx context.Context) error {
	conf := config.FromContext(ctx)
	var wg sync.WaitGroup
	for server, sconf := range conf.Servers {
		irccon := irc.IRC(sconf.Identity.Nick, sconf.Identity.Name)
		irccon.Log.SetOutput(conf.Main.LogWriter)
		irccon.VerboseCallbackHandler = conf.Main.LogLevel == "debug"
		irccon.Debug = conf.Main.LogLevel == "debug"
		irccon.UseTLS = sconf.SSL
		irccon.Password = sconf.Password
		irccon.TLSConfig = &tls.Config{InsecureSkipVerify: sconf.SkipInsecureVerify}

		// Join configured channels
		irccon.AddCallback("001", func(e *irc.Event) {
			for _, channel := range sconf.Channels {
				irccon.Join(channel)
			}
		})

		// Have the bot parse any messages in a channel to see if it should act
		irccon.AddCallback("PRIVMSG", func(e *irc.Event) {
			if stringInSlice(e.Nick, sconf.Ignore) {
				irccon.Log.Printf("Ignoring %q", e.Nick)
				return
			}
			go HandleMessages(ctx, irccon, e)
		})

		err := irccon.Connect(conf.ServerPort(server))
		if err != nil {
			return fmt.Errorf("error connecting to IRC server %q: %w", server, err)
		}
		wg.Add(1)
		go func(i *irc.Connection) {
			i.Loop()
			wg.Done()
		}(irccon)
	}
	wg.Wait()
	return nil
}

func stringInSlice(s string, sl []string) bool {
	for _, v := range sl {
		if s == v {
			return true
		}
	}
	return false
}

// RequestReply abstracts sending a command to the IRC server and listening for a reply. Eventcodes must match the
// commands. See https://www.alien.net.au/irc/irc2numerics.html
func RequestReply(c *irc.Connection, eventcode, command string) (string, error) {
	reply := make(chan string, 1)
	id := c.AddCallback(eventcode, func(e *irc.Event) {
		reply <- e.Message()
		defer close(reply)
	})
	defer c.RemoveCallback(eventcode, id)
	c.SendRaw(command)
	var r string
	select {
	case r = <-reply:
	case <-time.NewTicker(5 * time.Second).C:
		return "", errors.New("timeout getting command response")
	}
	log.Debugf("%+v", r)
	return r, nil
}
