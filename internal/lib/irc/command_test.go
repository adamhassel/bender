package irc

import (
	"context"
	"reflect"
	"testing"

	"github.com/adamhassel/bender/internal/config"
)

func TestParseCommand(t *testing.T) {
	type args struct {
		ctx context.Context
		msg string
	}
	conf := config.Config{
		Main: config.Main{
			CommandChar: "!",
		},
	}
	ctx := conf.Context(context.Background())
	tests := []struct {
		name    string
		args    args
		wantCmd Command
		wantErr bool
	}{
		{
			name: "single command",
			args: args{
				ctx: ctx,
				msg: "!beatme",
			},
			wantCmd: Command{
				Command:  "beatme",
				Argument: "",
			},
			wantErr: false,
		},
		{
			name: "command with argument",
			args: args{
				ctx: ctx,
				msg: "!beatme some argument",
			},
			wantCmd: Command{
				Command:  "beatme",
				Argument: "some argument",
			},
			wantErr: false,
		},
		{
			name: "Not a command",
			args: args{
				ctx: ctx,
				msg: "hello!",
			},
			wantCmd: Command{
				Command:  "",
				Argument: "",
			},
			wantErr: true,
		},
		{
			name: "empty",
			args: args{
				ctx: ctx,
				msg: "",
			},
			wantCmd: Command{
				Command:  "",
				Argument: "",
			},
			wantErr: true,
		},
		{
			name: "command char only",
			args: args{
				ctx: ctx,
				msg: "!",
			},
			wantCmd: Command{
				Command:  "",
				Argument: "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, err := ParseCommand(tt.args.ctx, tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCmd, tt.wantCmd) {
				t.Errorf("ParseCommand() gotCmd = %v, want %v", gotCmd, tt.wantCmd)
			}
		})
	}
}
