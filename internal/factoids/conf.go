package factoids

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	DatabaseFile string `yaml:"database"`
}

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