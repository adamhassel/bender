package irc

import (
	"crypto/tls"
	"fmt"
	"sync"

	"github.com/adamhassel/bender/internal/config"

	irc "github.com/thoj/go-ircevent"
)

const channel = "#cafeen"

func InitBot(conf config.Config) error {
	var wg sync.WaitGroup
	for server, sconf := range conf.Servers {
		irccon := irc.IRC(sconf.Identity.Nick, sconf.Identity.Name)
		irccon.VerboseCallbackHandler = true
		irccon.Debug = true
		irccon.UseTLS = sconf.SSL
		irccon.Password = sconf.Password
		irccon.TLSConfig = &tls.Config{InsecureSkipVerify: sconf.SkipInsecureVerify}

		// Join configured channels
		irccon.AddCallback("001", func(e *irc.Event) {
			for _, channel:= range sconf.Channels {
				irccon.Join(channel)
			}
		})

		// Have the bot parse any messages in a channel to see if it should act
		irccon.AddCallback("PRIVMSG", func(e *irc.Event) {
			go HandleMessages(irccon, e)
		})

		err := irccon.Connect(conf.ServerPort(server))
		if err != nil {
			return fmt.Errorf("error connecting to IRC server %q: %w", server, err)
			continue
		}
		wg.Add(1)
		go func (i *irc.Connection) {
			i.Loop()
			wg.Done()
		}(irccon)
	}
	wg.Wait()
	return nil
}