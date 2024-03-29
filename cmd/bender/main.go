package main

import (
	"context"
	"log"

	"github.com/adamhassel/bender/internal/config"
	"github.com/adamhassel/bender/internal/lib/irc"
	"github.com/adamhassel/bender/internal/lib/plugins"
)

// TODO: accept command line arg to specify config file
const defaultConffile = "conf/conf.yml"

func main() {
	var c config.Config
	err := config.ParseConfFile(defaultConffile, &c)

	if err != nil {
		log.Fatalf("%v", err)
	}

	setServerIdentity(&c)
	config.InitLogger(&c)
	ctx := c.Context(context.Background())
	if err := plugins.LoadPlugins(c.Plugins); err != nil {
		log.Println(err)
	}
	if err := irc.InitBot(ctx); err != nil {
		log.Printf("error initializing bot: %s", err)
	}
}

func setServerIdentity(c *config.Config) {
	// use global identity if none is set per server
	for server, sconf := range c.Servers {
		if sconf.Identity.Nick == "" {
			sconf.Identity.Nick = c.Identity.Nick
		}
		if sconf.Identity.Name == "" {
			sconf.Identity.Name = c.Identity.Name
		}
		if sconf.Identity.Modestring == "" {
			sconf.Identity.Modestring = c.Identity.Modestring
		}
		c.Servers[server] = sconf
	}
}
