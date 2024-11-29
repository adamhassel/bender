package main

import (
	"testing"

	"github.com/adamhassel/bender/internal/helpers"
)

func Test_cleanURL(t *testing.T) {
	type args struct {
		url string
	}

	cleanlist = helpers.NewSet("foo", "bar", "baz")

	var tests = []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "simple",
			args: args{
				url: "http://www.example.com?foo=bla&something=other",
			},
			want:    "http://www.example.com?something=other",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cleanURL(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("cleanURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("cleanURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}
