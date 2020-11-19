package irc

import (
	"context"
	"errors"
	"strings"

	"github.com/adamhassel/bender/internal/config"
)

type Command struct {
	Command  string
	Argument string
}

var ErrNotCommand = errors.New("not a command")

func ParseCommand(ctx context.Context, msg string) (cmd Command, err error) {
	c := config.FromContext(ctx)
	if !strings.HasPrefix(msg, c.Main.CommandChar) {
		err = ErrNotCommand
		return
	}
	cmd.Command, cmd.Argument = splitBySpace(strings.TrimPrefix(msg, c.Main.CommandChar))
	return
}

func splitBySpace(in string) (a, b string) {
	out := strings.SplitN(in, " ", 2)
	if len(out) == 1 {
		return out[0], ""
	}
	return out[0], out[1]
}
