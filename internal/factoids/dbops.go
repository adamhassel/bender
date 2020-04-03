// Package factoids implemets a simple factoid get/set system for the bot
package factoids

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v2"
)

// factoids is the data struture used for thread safe in memorystorage of factoids
type factoids struct {
	m  sync.Mutex
	v  map[string]StringSet
	db string
}

const (
	DefaultConfFile = "conf/factoids.yml"
	DefaultDBPath   = "db/factoids.yml"
)

// f is the package-level in memory cache of factoids. It is synced to disk on every change.
var f factoids

var ErrNoSuchFact = errors.New("factoid not found")
var ErrAmbiguousKey = errors.New("ambiguous key")
var ErrFactAlreadyExists = errors.New("fact already exists")
var ErrInvalidUTF8 = errors.New("invalid UTF-8")

func init() {
	f.v = make(map[string]StringSet)
	c, err := ParseConfFile(DefaultConfFile)
	if err != nil {
		log.Print(err)
	}
	if c.DatabaseFile == "" {
		c.DatabaseFile = DefaultDBPath
	}
	if err := loadDB(c.DatabaseFile); err != nil {
		log.Print(err)
	}
	rand.Seed(time.Now().UnixNano())
}

func loadDB(filename string) error {
	f.db = filename
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error loading database at %q: %w", err)
	}
	factsfromdisk := make(map[string][]string)
	if err := yaml.Unmarshal(content, &factsfromdisk); err != nil {
		return fmt.Errorf("error parsing database at %q: %w", err)
	}
	f.m.Lock()
	defer f.m.Unlock()
	f.v = make(map[string]StringSet)
	for k, vs := range factsfromdisk {
		f.v[k] = NewStringSet(vs...)
	}
	return nil
}

// Set adds a value to a factoid key
func Set(key, value string) error {
	if !utf8.Valid([]byte(value)) {
		return ErrInvalidUTF8
	}
	f.m.Lock()
	defer f.m.Unlock()
	if _, ok := f.v[key]; !ok {
		f.v[key] = NewStringSet()
	}
	if f.v[key].Exists(value) {
		return ErrFactAlreadyExists
	}
	f.v[key].Add(value)

	return syncToDisk()
}

// Get retrieves a random fact from the factoid DB
func Get(key string) (string, error) {
	vals, err := f.getall(key)
	if err != nil || len(vals) == 0 {
		return "", err
	}
	// pick a random fact from the list
	return vals[rand.Intn(len(vals))], nil
}

func (f *factoids) getall(key string) ([]string, error) {
	f.m.Lock()
	vals, ok := f.v[key]
	f.m.Unlock()
	if !ok || len(vals) == 0 {
		return nil, ErrNoSuchFact
	}
	return vals.Slice(), nil
}

// Delete removes a value from a key matching substring `substr`. If more than one match, return error
func Delete(key, substr string) error {
	vals, err := f.getall(key)
	// getall returns an error if no fact found
	if err != nil {
		return err
	}
	// find substr in list
	var res string
	for _, val := range vals {
		if strings.HasPrefix(val, substr) {
			// if we already found an element, return
			if len(res) > 0 {
				return ErrAmbiguousKey
			}
			res = val
		}
	}
	// delete the found element
	f.m.Lock()
	f.v[key].Delete(res)
	if err := syncToDisk(); err != nil {
		return err
	}
	f.m.Unlock()
	return nil
}

func (f *factoids) delete(key string) {
	f.m.Lock()
	delete(f.v, key)
	f.m.Unlock()
}

// ListFacts lists all keys mathcing `substring`
func ListFacts(substr string) ([]string, error) {
	return nil, nil
}

// sync syncs the in memory DB to disk. THe caller should lock!
func syncToDisk() error {
	factsfordisk := make(map[string][]string)
	for k, vs := range f.v {
		factsfordisk[k] = vs.Slice()
	}
	ymldata, err := yaml.Marshal(factsfordisk)
	if err != nil {
		return fmt.Errorf("error marshalling DB: %w", err)
	}
	if err := ioutil.WriteFile(f.db, ymldata, 0644); err != nil {
		return fmt.Errorf("error syncing to file %q: %w", err)
	}
	return nil
}
