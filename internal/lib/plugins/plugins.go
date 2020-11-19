// Package plugins implements plugins support
package plugins

import (
	"errors"
	"fmt"
	"io/ioutil"
	"plugin"

	log "github.com/sirupsen/logrus"
	irc "github.com/thoj/go-ircevent"
	"gopkg.in/yaml.v2"
)

type Plugin struct {
	*plugin.Plugin
	path string
}

type mss map[string]string

type PluginConf struct {
	mss
	Config map[string]interface{} `yaml:"config"`
}

type pluginFunc func([]string, *irc.Event) (string, bool)
type matchFunc func(string, *irc.Event) (string, bool)
type matchFuncs []matchFunc

// Exported error vars
var (
	ErrNoExportedMatchers = errors.New("plugin has no exported matchers")
)

var (
	// commands holds all commands configured by plugins
	commands map[string]pluginFunc
	// matchers holds all matchers defined in plugins. The key is the plugin name, to make it possible to have name clasges
	// in different plugins
	matchers map[string]matchFuncs
)

// loadPluginConf loads per-plugins configuration
func loadPluginConf(filename string) (map[string]interface{}, error) {
	c := make(map[string]interface{})
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %w", filename, err)
	}
	if err := yaml.Unmarshal(content, &c); err != nil {
		return nil, fmt.Errorf("error parsing configuration at %s: %w", filename, err)
	}
	return c, nil
}

func loadPlugins(pluginConf map[string]string) error {
	for pluginFile, confFile := range pluginConf {
		p, err := plugin.Open(pluginFile)
		if err != nil {
			return fmt.Errorf("error loading plugins %s: %w", pluginFile, err)
		}
		config, err := loadPluginConf(confFile)
		if err != nil {
			return fmt.Errorf("error loading plugins config: %s: %w", confFile, err)
		}
		for command, f := range config {
			val, ok := f.(string)
			if !ok {
				if command == "config" {
					if err := setPluginConf(p, f.(map[interface{}]interface{})); err != nil {
						log.Errorf("error configuring plugin %q: %s", pluginFile, err)
					}
				}
				continue
			}
			sym, err := p.Lookup(val)
			if err != nil {
				return fmt.Errorf("symbol %q lookup error: %w", f, err)
			}
			if commands == nil {
				commands = make(map[string]pluginFunc)
			}
			if c, ok := sym.(func([]string, *irc.Event) (string, bool)); ok {
				if _, ok := commands[command]; ok {
					return fmt.Errorf("command name clash: %q is already defined", command)
				}
				commands[command] = c
				continue
			}
			return fmt.Errorf("symbol %q does not match signature", f)
		}
		if err := configureMatchers(&Plugin{p, pluginFile}); err != nil {
			if errors.Is(err, ErrNoExportedMatchers) {
				continue
			}
			return err
		}
		log.Infof("Loaded plugins %q", pluginFile)
	}
	return nil
}

// configureMatcher will configure a plugin that has command-less matching functions defined
func configureMatchers(p *Plugin) error {
	l, err := p.Lookup("Matchers")
	if err != nil {
		log.Warn("plugin doesn't export matcher functions")
		return ErrNoExportedMatchers
	}
	list, ok := l.(*[]string)
	if !ok {
		return fmt.Errorf("invalid matcher export")
	}
	for _, fName := range *list {
		f, err := p.Lookup(fName)
		if err != nil {
			return fmt.Errorf("symbol %q lookup error: %w", f, err)
		}
		if matchers == nil {
			matchers = make(map[string]matchFuncs)
		}
		if m, ok := f.(func(string, *irc.Event) (string, bool)); ok {
			matchers[p.path] = append(matchers[p.path], m)
		}
	}
	return nil
}

// setPluginConf if called if plugin-specific configuration is found
func setPluginConf(p *plugin.Plugin, conf map[interface{}]interface{}) error {
	//c, ok := conf.(map[interface{}]interface{})
	/*if !ok {
		return fmt.Errorf("invalid plugin configuration: %T", conf)
	}

	*/
	cf, err := p.Lookup("Configure")
	if err != nil {
		return errors.New("configuration provided, but no Configure function")
	}
	f, ok := cf.(func(map[interface{}]interface{}) error)
	if !ok {
		return fmt.Errorf("\"Configure\" function has wrong signature: %T", cf)
	}
	return f(conf)
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

func Matchers(msg string, e *irc.Event) ([]Result, error) {
	var rv []Result
	for _, funcs := range matchers {
		for _, f := range funcs {
			msg, action := f(msg, e)
			if msg == "" {
				continue
			}
			if rv == nil {
				rv = make([]Result, 0, 1)
			}
			rv = append(rv, Result{msg, action})
		}
	}
	return rv, nil
}
