package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	irc "github.com/thoj/go-ircevent"
)

// Matchers lists exported matchers in the plugin
var Matchers = []string{"Chanlog"}

var loggers map[string]*log.Logger
var lm sync.Mutex
var activeRotator *gocron.Scheduler
var logroot string

// A logrus formatter
type IRCFormatter struct{}
type IRCSystemFormatter struct{}

// Format implements the logrus.Formatter interface
func (*IRCFormatter) Format(entry *log.Entry) ([]byte, error) {
	// make sure all fields are present
	user, userok := entry.Data["user"]
	if !userok {
		return nil, errors.New("required fields missing")
	}
	var msg string
	if action := entry.Data["action"].(bool); action {
		msg = fmt.Sprintf("%s *** %s %s\n", entry.Time.Format("15:04:05"), user, entry.Message)
	} else {
		msg = fmt.Sprintf("%s < %s> %s\n", entry.Time.Format("15:04:05"), user, entry.Message)
	}
	return []byte(msg), nil
}

// Format implements the logrus.Formatter interface
func (*IRCSystemFormatter) Format(entry *log.Entry) ([]byte, error) {
	return []byte(fmt.Sprintf("%s  *****  %s", entry.Time.Format("15:04:05"), entry.Message)), nil
}

// Chanlog is called for every message, and logs it if it's supposed to
func Chanlog(msg string, e *irc.Event) (string, bool) {
	channel := e.Arguments[0]

	// Check if we're configured to log this channel
	lm.Lock()
	defer lm.Unlock()
	logger, ok := loggers[channel]
	if !ok {
		return "", false
	}
	fields := log.Fields{"user": e.Nick}
	if e.Code == "CTCP_ACTION" {
		fields["action"] = true
	}
	logger.WithFields(fields).Info(msg)
	return "", false
}

// Configure configures the plugin
func Configure(c map[interface{}]interface{}) error {
	channels, ok := c["channels"]
	if !ok {
		return nil
	}
	var chanlist []string
	for _, ci := range channels.([]interface{}) {
		c, ok := ci.(string)
		if !ok {
			return fmt.Errorf("channels directive found, but not []string. Instead %T", channels)
		}
		if chanlist == nil {
			chanlist = make([]string, 0, len(channels.([]interface{})))
		}
		chanlist = append(chanlist, c)
	}

	var err error
	logroot, err = filepath.Abs(filepath.Dir(os.Args[0]))
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
		if err := configureLogging(channel, logger); err != nil {
			return err
		}
	}
	configureRotator()
	dateChangeLogger()
	return nil
}

// makeDatePath returns a path from the current date as "year/month"
func makeDatePath() string {
	now := time.Now()
	return filepath.Join(now.Format("2006"), now.Format("01"))
}

// getLogFileHandle will return a handle to a file named and located for logging
func getLogfilehandle(logroot, channel string) (*os.File, error) {
	logfile := filepath.Join(logroot, makeDatePath(), channel+".log")
	os.MkdirAll(path.Dir(logfile), 0700)
	return os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
}

func rotate(channel string) error {
	lm.Lock()
	logger, ok := loggers[channel]
	if !ok || logger == nil {
		log.WithField("channel", channel).Info("no logger defined for channel, so not stopping")
		lm.Unlock()
		return configureLogging(channel, logger)
	}
	if closer, ok := logger.Out.(io.Closer); ok {
		closer.Close()
	}
	lm.Unlock()
	// Replace the destination
	return configureLogging(channel, logger)
}

// configure logging for a channel.
func configureLogging(channel string, logger *log.Logger) error {
	if logroot == "" {
		return errors.New("logroot undefined. Did you run the \"Configure\" function?")
	}
	file, err := getLogfilehandle(logroot, channel)
	if err != nil {
		return fmt.Errorf("couldn't open logfile at %s: %w", channel, err)
	}
	lm.Lock()
	logger.Out = file
	logger.SetFormatter(new(IRCFormatter))
	loggers[channel] = logger
	lm.Unlock()
	return nil
}

func rotateAll() error {
	for channel, _ := range loggers {
		if e := rotate(channel); e != nil {
			return e
		}
	}
	return nil
}

// configureRotator will monitor time and trigger the rotation at midnight on the 1st of a month
func configureRotator() {
	if activeRotator != nil {
		activeRotator.Stop()
		activeRotator.Clear()
	}
	activeRotator = gocron.NewScheduler(time.Local)
	if _, err := activeRotator.Every(1).Month(1).At("00:00").Do(rotateAll); err != nil {
		log.WithError(err).Error("couldn't run rotator")
		return
	}
	activeRotator.StartAsync()
}

// dateChangeLogger will make a note in the logfile whenever the date changes. Run once.
func dateChangeLogger() {
	dl := gocron.NewScheduler(time.Local)
	if _, err := dl.Every(1).Day().At("00:00").Do(logDateChange); err != nil {
		log.WithError(err).Error("couldn't run rotator")
		return
	}
	dl.StartAsync()
}

// logDateChange will log a date change for every configured channel
func logDateChange() {
	lm.Lock()
	for _, logger := range loggers {
		logger.SetFormatter(new(IRCSystemFormatter))
		logger.Infof("Date changed to %s", time.Now().Format("Jan 02 2006"))
		logger.SetFormatter(new(IRCFormatter))
	}
	defer lm.Unlock()
}
