package factoids

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/adamhassel/bender/internal/helpers"
)

type Config struct {
	DatabaseFile string                `yaml:"database"`
	ReplyStrings helpers.Slice[string] `yaml:"replystrings"`
}

type ctxconf int

const configkey ctxconf = iota

// ParseConfFile parses configuration in `filename` and returns a configuration and an error
func ParseConfFile(filename string) (Config, error) {
	content, err := os.ReadFile(filename)
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

func SaveToFile(filename string, c Config) error {
	raw, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}
	if err := os.WriteFile(filename, raw, 0644); err != nil {
		return fmt.Errorf("error writing file %s: %w", filename, err)
	}
	return nil
}

// Context returns a new context from ctx with c attached
func (c Config) Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, configkey, c)
}

func FromContext(ctx context.Context) Config {
	c := ctx.Value(configkey)
	config, ok := c.(Config)
	if ok {
		return config
	}
	return Config{}
}
