package irc

import (
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/adamhassel/bender/internal/config"
	"github.com/sirupsen/logrus"

	irc "github.com/thoj/go-ircevent"
)

func confLogger(conf config.Main) *io.PipeWriter {
	logger := logrus.New()
	level, err := logrus.ParseLevel(conf.LogLevel)
	if err != nil {
		fmt.Printf("unknown loglevel %q, defaulting to 'debug'")
		level = logrus.DebugLevel
	}
	logger.Out = os.Stderr
	file, err := os.OpenFile(conf.Logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logger.Out = file
	} else {
		logger.Info("Failed to log to file, using default stderr")
	}

	logger.SetLevel(level)
	logger.Formatter = &logrus.TextFormatter{}
	return logger.Writer()
}

func InitBot(conf config.Config) error {
	var wg sync.WaitGroup
	logwriter := confLogger(conf.Main)
	for server, sconf := range conf.Servers {
		irccon := irc.IRC(sconf.Identity.Nick, sconf.Identity.Name)
		irccon.Log.SetOutput(logwriter)
		irccon.VerboseCallbackHandler = true
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
			go HandleMessages(irccon, e)
		})

		err := irccon.Connect(conf.ServerPort(server))
		if err != nil {
			return fmt.Errorf("error connecting to IRC server %q: %w", server, err)
			continue
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
