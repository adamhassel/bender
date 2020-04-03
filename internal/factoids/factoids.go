package factoids

import (
	"fmt"
	"strings"
)

func Lookup(msg string) string {
	factoidstring := strings.TrimPrefix(msg, "!? ")
	factoidstring = strings.TrimSpace(factoidstring)
	factoid, err := Get(factoidstring)
	if err == ErrNoSuchFact {
		return fmt.Sprintf("Nobody cares about %q!", factoidstring)
	}
	return fmt.Sprintf("Some say that %s is %s", factoidstring, factoid)
}

func Store(msg string) string {
	factoidstring := strings.TrimPrefix(msg, "!! ")
		f := strings.SplitN(factoidstring, " is ", 2)
		if len(f) != 2 {
			return "Argh!"
		}
		key, val := strings.TrimSpace(f[0]), strings.TrimSpace(f[1])
		if err := Set(key, val); err != nil {
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