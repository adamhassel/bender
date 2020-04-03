package config

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"gopkg.in/yaml.v2"
)

type Main struct {
	Logfile  string `yaml:"logfile"`
	LogLevel string `yaml:"loglevel"`
}

type Identity struct {
	Nick       string `yaml:"nick"`
	Name       string `yaml:"name"`
	Modestring string `yaml:"modestring"`
}

type ServerOpts struct {
	Port               int      `yaml:"port"`
	SSL                bool     `yaml:"ssl"`
	SkipInsecureVerify bool     `yaml:"sslskipverify"`
	Password           string   `yaml:"password"`
	Channels           []string `yaml:"channels"`
	Ignore             []string `yaml:"ignore"`
	Identity           Identity `yaml:"identity"`
}

type Config struct {
	Main     Main                  `yaml:"main"`
	Identity Identity              `yaml:"identity"`
	Servers  map[string]ServerOpts `yaml:"servers"`
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
	return c, nil
}

// ServerPort returns a string of servername:port
func (c Config) ServerPort(s string) string {
	if sc, ok := c.Servers[s]; ok {
		return s + ":" + strconv.Itoa(sc.Port)
	}
	return ""
}
