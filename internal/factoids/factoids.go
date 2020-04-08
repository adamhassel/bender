package factoids

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// NewReplyString adds a custom reply string to the list of reply strings
// TODO: generalize config file read/write/parse
// TODO: switch to toml for config
func NewReplyString(ctx *context.Context, msg string) string {
	c := FromContext(*ctx)
	replystring := strings.TrimPrefix(msg, "!newreply ")
	if strings.Count(replystring, "%s") != 2 {
		return "You need exactly two '%s' captures in your reply"
	}
	c.ReplyStrings = append(c.ReplyStrings, replystring)
	*ctx = c.Context(*ctx)
	if err := SaveToFile(DefaultConfFile, c); err != nil {
		logrus.Error(err)
	}
	return `OK, I'll use "` + replystring + `"in replies`
}

// Lookup returns a string to output to the channel, and a bool indicating if it's an action ('/me blabla')
func Lookup(ctx context.Context, msg string) (string, bool) {
	factoidstring := strings.TrimPrefix(msg, "!? ")
	factoidstring = strings.TrimSpace(factoidstring)
	factoid, err := get(strings.ToLower(factoidstring))
	if err == ErrNoSuchFact {
		return fmt.Sprintf("Nobody cares about %q!", factoidstring), false
	}
	if strings.HasPrefix(factoid.Value, "<reply> ") {
		factoid.Value = strings.TrimPrefix(factoid.Value, "<reply> ")
		return factoid.Value, false
	}
	if strings.HasPrefix(factoid.Value, "<me> ") {
		return strings.TrimPrefix(factoid.Value, "<me> "), true
	}
	// pick a random replystring
	c := FromContext(ctx)
	reply := c.ReplyStrings.Random()

	return fmt.Sprintf(reply, factoidstring, factoid.Value), false
}

// Store saves a factoid to the database
func Store(msg string, from string) string {
	factoidstring := strings.TrimPrefix(msg, "!! ")
	splitword := "is"
	f := strings.SplitN(factoidstring, " is ", 2)
	if len(f) != 2 {
		splitword = "er"
		f = strings.SplitN(factoidstring, " er ", 2)
		if len(f) != 2 {
			return "You gotta format it right, moron."
		}
	}
	key, val := strings.TrimSpace(f[0]), strings.TrimSpace(f[1])
	fact := factoid{Value: val, Origin: from, SplitWord: splitword, Created: time.Now().Round(time.Second)}
	if err := set(strings.ToLower(key), fact); err != nil {
		switch err {
		case ErrFactAlreadyExists:
			return "I know that already"
		case ErrInvalidUTF8:
			return "Your factoid is not valid UTF8"
		}
		return err.Error()

	}
	return fmt.Sprintf("OK, %q %s %q", key, splitword, val)
}
