package main

import (
	"context"
	"log"

	"github.com/adamhassel/bender/internal/config"
	"github.com/adamhassel/bender/internal/irc"
)

const defaultConffile = "conf/conf.yml"

func main() {
	conf, err := config.ParseConfFile(defaultConffile)

	if err != nil {
		log.Fatalf("%v", err)
	}
	config.InitLogger(&conf)
	ctx := conf.Context(context.Background())
	irc.InitBot(ctx, conf)
}
