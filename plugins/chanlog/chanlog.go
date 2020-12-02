package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	irc "github.com/thoj/go-ircevent"
)

// Matchers lists exported matchers in the plugin
var Matchers = []string{"Chanlog"}

var loggers map[string]*log.Logger

// A logrus formatter
type IRCFormatter struct{}

// Format implements the logrus.Formatter interface
func (IRCFormatter) Format(entry *log.Entry) ([]byte, error) {
	// make sure all fields are present
	user, userok := entry.Data["user"]
	if !userok {
		return nil, errors.New("required fields missing")
	}
	msg := fmt.Sprintf("%s < %s> %s\n", entry.Time.Format("15:04:05"), user, entry.Message)
	return []byte(msg), nil
}

// Chanlog is called for every message, and logs it if it's supposed to
func Chanlog(msg string, e *irc.Event) (string, bool) {
	channel := e.Arguments[0]

	// Check if we're configured to log this channel
	logger, ok := loggers[channel]
	if !ok {
		return "", false
	}
	logger.WithFields(log.Fields{"user": e.Nick}).Info(msg)
	return "", false
}

// Configure configures the plugin
func Configure(c map[interface{}]interface{}) error {
	channels, ok := c["channels"]
	if !ok {
		return nil
	}
	chanlist, ok := channels.([]string)
	if !ok {
		return fmt.Errorf("channels directive found, but not []string. Instead %T", channels)
	}

	logroot, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return fmt.Errorf("error determining bot root: %w", err)
	}
	lr, ok := c["logroot"]
	if ok {
		logroot, ok = lr.(string)
		if !ok {
			return fmt.Errorf("expected string logroot, got %T", lr)
		}
	}

	for _, channel := range chanlist {
		if loggers == nil {
			loggers = make(map[string]*log.Logger, len(chanlist))
		}
		logger := log.New()
		logfile := filepath.Join(logroot, makeDatePath(), channel+".log")
		file, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("couldn't open logfile at %s: %w", channel, err)
		}
		logger.Out = file
		logger.SetFormatter(new(IRCFormatter))
		loggers[channel] = logger
	}
	return nil
}

// makedDatePath returns a path from the current date as "year/month/day"
func makeDatePath() string {
	now := time.Now()
	return filepath.Join(now.Format("2006"), now.Format("01"), now.Format("02"))
}

// rotate the logger's files. Close existing, open a new. Used e.g. on date changes
func rotate(logger *log.Logger) error {
	// TODO: implement :P
	return nil
}
