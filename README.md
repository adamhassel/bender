# bender
An IRC bot in Go

A work in progress.

## Install

    go get github.com/adamhassel/bender
	go build -o bender cmd/bender/main.go
	# optional, if you want plugins:
	go build -buildmode=plugin plugins/<plugin>/<plugin>.go -o <plugin>.so
	# edit config, save in conf/conf.yml
	./bender

## Feature list:

### Core
* Multiple channels
* multiple servers
* Ignore (e.g. other bots)
* Plugin support, see README in `plugins` dir.

### Factoid database

* Stores factoids from users
* Stores metadata about factoids: user name, time stamp
* supports verbatim replies and actions
* custom reply patterns

### Beatme

A fun friday game. `op` the bot and have it kick random channel members

## See also the TODO and Issues list for planned stuff
