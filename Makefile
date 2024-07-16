# List of plugins to build
PLUGINS:=urlshort chanlog
# Name of bot main executable
BOT:=bender

PLUGINS_T:=$(addsuffix .so,$(addprefix plugins/,$(PLUGINS)))
expand = plugins/$1/$1.go

# default target
all: bot plugins

bot: cmd/bender/main.go
	go build -o $(BOT) $<

clean:
	rm $(BOT)
	rm $(PLUGINS_T)

plugins: $(PLUGINS_T)

.SECONDEXPANSION:
plugins/%.so: $$(call expand,$$*)
	go build --buildmode=plugin -o $@ $<

