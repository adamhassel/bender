package factoids

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var lastfact fullfactoid

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

func Lastfact() fullfactoid {
	return lastfact
}

// Info formats an info string about f
func (f fullfactoid) Info() string {
	info := fmt.Sprintf("\"%s => %s\" was created", f.Keyword, f.Value)
	switch {
	case f == fullfactoid{}:
		info = "I haven't looked up any facts yet"
	case f.Origin != "":
		info = fmt.Sprintf("%s by %s", info, f.Origin)
		fallthrough
	case f.Created != nil:
		info = fmt.Sprintf("%s on %s", info, f.Created.Format(time.RFC822))
	case f.Origin == "" && f.Created == nil:
		info = fmt.Sprintf("I don't have any information on \"%s => %s\"", f.Keyword, f.Value)
	}
	return info
}

// Lookup returns a string to output to the channel, and a bool indicating if it's an action ('/me blabla')
func Lookup(ctx context.Context, msg string) (string, bool) {
	factoidstring := strings.TrimPrefix(msg, "!? ")
	factoidstring = strings.TrimSpace(factoidstring)
	factoid, err := get(strings.ToLower(factoidstring))
	if err == ErrNoSuchFact {
		return fmt.Sprintf("Nobody cares about %s!", factoidstring), false
	}
	if strings.HasPrefix(factoid.Value, "<reply>") {
		factoid.Value = strings.TrimSpace(strings.TrimPrefix(factoid.Value, "<reply>"))
		return factoid.Value, false
	}
	if strings.HasPrefix(factoid.Value, "<me> ") {
		return strings.TrimPrefix(factoid.Value, "<me> "), true
	}
	// pick a random replystring
	c := FromContext(ctx)
	reply := c.ReplyStrings.Random()
	lastfact = fullfactoid{factoidstring, factoid}
	return fmt.Sprintf(reply, factoidstring, factoid.Value), false
}

// Search will look through the entire database, both keywords and facts, for the regular expression in rex. It will
// return a formatted string to output toi a channel, and an error if something went wrong. It is not an error that
// nothing was found
func Search(rex string, maxresults int) ([]string, error) {
	re, err := regexp.Compile(rex)
	if err != nil {
		return nil, err
	}
	results, additional := f.search(re, maxresults)
	if len(results) == 0 {
		return []string{"No results found"}, nil
	}
	reslen := len(results) + 1
	if additional > 0 {
		reslen++
	}
	rv := make([]string, 0, reslen)
	rv = append(rv, "Check out what I found from your search:")
	for _, f := range results {
		rv = append(rv, fmt.Sprintf("%q => %q", f.Keyword, f.Value))
	}
	if additional > 0 {
		rv = append(rv, fmt.Sprintf("... and %d more results not displayed", additional))
	}
	return rv, nil
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
	now := time.Now().Round(time.Second)
	fact := factoid{Value: val, Origin: from, SplitWord: splitword, Created: &now}
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
