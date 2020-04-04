package factoids

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/adamhassel/bender/internal/helpers"
	"gopkg.in/yaml.v2"
)

type Config struct {
	DatabaseFile string              `yaml:"database"`
	ReplyStrings helpers.StringSlice `yaml:"replystrings"`
}

type ctxconf int

const configkey ctxconf = iota

// ParseConfFile parses configuration in `filename` and returns a configuration and an error
func ParseConfFile(filename string) (Config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return Config{}, fmt.Errorf("error reading %q: %w", filename, err)
	}
	var c Config
	if err := yaml.Unmarshal(content, &c); err != nil {
		return Config{}, fmt.Errorf("error parsing configuration at %s: %w", filename, err)
	}
	if c.DatabaseFile == "" {
		c.DatabaseFile = DefaultDBPath
	}
	return c, nil
}

func writeConfFile(c Config) {
}

// Context returns a new context from ctx with c attached
func (c Config) Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, configkey, c)
}

func ConfigFromContext(ctx context.Context) Config {
	c := ctx.Value(configkey)
	config, ok := c.(Config)
	if ok {
		return config
	}
	return Config{}
}
