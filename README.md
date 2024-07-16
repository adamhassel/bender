# bender
An IRC bot in Go

A work in progress.

## Installation

Following are naïve installation instructions.

### Download the source

    git clone git@github.com:adamhassel/bender.git

### Build using GNU Make

	cd bender
    make

### Or build manually if that's your preference
	
	cd bender
	go build -o bender cmd/bender/main.go
	# optional, if you want plugins:
	go build -buildmode=plugin -o <plugin.so> plugins/<plugin>/<plugin>.go
	# edit config, save in conf/conf.yml
	./bender

### Install

Make sure you cofigure the bot. Default config file is `conf/conf.yaml`. There's an example in there with reasonable defaults. Also remember to include any plugin configuration. See the README in the plugins dir for more information.

	cp -r bender plugins/*.so conf/conf.yaml <target_dir>
	cd target_dir
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
