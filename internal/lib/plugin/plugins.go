// Package plugin implements plugin support
package plugins

import (
	"fmt"
	"io/ioutil"
	"log"
	"plugin"

	irc "github.com/thoj/go-ircevent"
	"gopkg.in/yaml.v2"
)

var commands map[string]pluginfunc

type pluginfunc func([]string, *irc.Event) (string, bool)

// loadPluginConf loads per-plugin configuration
func loadPluginConf(filename string) (map[string]string, error) {
	c := make(map[string]string)
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %w", filename, err)
	}
	if err := yaml.Unmarshal(content, &c); err != nil {
		return nil, fmt.Errorf("error parsing configuration at %s: %w", filename, err)
	}
	return c, nil
}

func loadPlugins(pluginconf map[string]string) error {
	for pluginfile, conffile := range pluginconf {
		p, err := plugin.Open(pluginfile)
		if err != nil {
			return fmt.Errorf("error loading plugin %s: %w", pluginfile, err)
		}
		config, err := loadPluginConf(conffile)
		if err != nil {
			return fmt.Errorf("error loading plugin config: %s: %w", conffile, err)
		}
		for command, f := range config {
			sym, err := p.Lookup(f)
			if err != nil {
				return fmt.Errorf("symbol %q lookup error: %w", f, err)
			}
			if commands == nil {
				commands = make(map[string]pluginfunc)
			}
			if c, ok := sym.(func([]string, *irc.Event) (string, bool)); ok {
				commands[command] = c
				continue
			}
			return fmt.Errorf("symbol %q does not match signature", f)
		}
		log.Printf("Loaded plugin %q", pluginfile)
	}
	return nil
}

// LoadPlugins loads plugins and their configuration into memory
func LoadPlugins(config map[string]string) error {
	return loadPlugins(config)
}

type Result struct {
	Message string
	Action  bool
}

func Execute(command string, args []string, e *irc.Event) (Result, error) {
	c, ok := commands[command]
	if !ok {
		return Result{}, fmt.Errorf("command %q not found in loaded plugins", command)
	}
	msg, action := c(args, e)
	return Result{msg, action}, nil
}
