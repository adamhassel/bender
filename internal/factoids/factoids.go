package factoids

import (
	"fmt"
	"strings"
)

// Lookup returns a string to output to the channel, and a bool indicating if it's an action ('/me blabla')
func Lookup(msg string) (string, bool) {
	factoidstring := strings.TrimPrefix(msg, "!? ")
	factoidstring = strings.TrimSpace(factoidstring)
	factoid, err := Get(strings.ToLower(factoidstring))
	if err == ErrNoSuchFact {
		return fmt.Sprintf("Nobody cares about %q!", factoidstring), false
	}
	if strings.HasPrefix(factoid, "<reply> ") {
		factoid = strings.TrimPrefix(factoid, "<reply> ")
		return factoid, false
	}
	if strings.HasPrefix(factoid, "<me> ") {
		return strings.TrimPrefix(factoid, "<me> "), true
	}
	return fmt.Sprintf("Some say that %s is %s", factoidstring, factoid), false
}

// Store saves a factoid to the database
func Store(msg string) string {
	factoidstring := strings.TrimPrefix(msg, "!! ")
	f := strings.SplitN(factoidstring, " is ", 2)
	if len(f) != 2 {
		return "You gotta format it right, moron."
	}
	key, val := strings.TrimSpace(f[0]), strings.TrimSpace(f[1])
	if err := Set(strings.ToLower(key), val); err != nil {
		switch err {
		case ErrFactAlreadyExists:
			return "I know that already"
		case ErrInvalidUTF8:
			return "Your factoid is not valid UTF8"
		}
		return err.Error()

	}
	return fmt.Sprintf("OK, %q is %q", key, val)
}
