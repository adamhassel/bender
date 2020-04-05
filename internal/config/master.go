package config

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Main struct {
	Logfile   string    `yaml:"logfile"`
	LogLevel  string    `yaml:"loglevel"`
	LogWriter io.Writer `yaml:"-"`
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

type ctxconf int

const configkey ctxconf = iota

func InitLogger(config *Config) {
	logger := logrus.New()
	level, err := logrus.ParseLevel(config.Main.LogLevel)
	if err != nil {
		fmt.Printf("unknown loglevel %q, defaulting to 'debug'")
		level = logrus.DebugLevel
	}
	logger.Out = os.Stderr
	file, err := os.OpenFile(config.Main.Logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logger.Out = file
	} else {
		logger.Info("Failed to log to file, using default stderr")
	}

	logger.SetLevel(level)
	logger.Formatter = &logrus.TextFormatter{}
	config.Main.LogWriter = logger.Writer()
	log.SetOutput(logger.Writer())
}

// ParseConfFile parses configuration in `filename`, saves it in `c` and returns an error
func ParseConfFile(filename string, c *Config) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading %q: %w", filename, err)
	}
	if err := yaml.Unmarshal(content, &c); err != nil {
		return fmt.Errorf("error parsing configuration at %s: %w", filename, err)
	}
	return nil
}

func SaveToFile(filename string, c Config) error {
	raw, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}
	if err := ioutil.WriteFile(filename, raw, 0644); err != nil {
		return fmt.Errorf("error writing file %s: %w", filename, err)
	}
	return nil
}

// ServerPort returns a string of servername:port
func (c Config) ServerPort(s string) string {
	if sc, ok := c.Servers[s]; ok {
		return s + ":" + strconv.Itoa(sc.Port)
	}
	return ""
}

// Context returns a new context from ctx with c attached
func (c Config) Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, configkey, c)
}

// FromContext extracts configuration from a config if present
func FromContext(ctx context.Context) Config {
	c := ctx.Value(configkey)
	config, ok := c.(Config)
	if ok {
		return config
	}
	return Config{}
}
