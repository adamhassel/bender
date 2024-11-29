// Package factoids implemets a simple factoid get/set system for the bot
package factoids

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"

	log "github.com/sirupsen/logrus"

	"github.com/adamhassel/bender/internal/helpers"
)

// factoids is the data struture used for thread safe in memorystorage of factoids
type factoids struct {
	m  sync.Mutex
	v  map[string]FactoidSet
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

// TODO: change how we load configs, maybe pass along in contexts?
func init() {
	f.v = make(map[string]FactoidSet)
	c, err := ParseConfFile(DefaultConfFile)
	if err != nil {
		log.Error(err)
	}
	if c.DatabaseFile == "" {
		c.DatabaseFile = DefaultDBPath
	}
	if err := loadDB(c.DatabaseFile); err != nil {
		log.Error(err)
	}
}

func RandomKey() string {
	keys := make(helpers.Slice[string], len(f.v))
	var i int
	for k := range f.v {
		keys[i] = k
		i++
	}
	return keys.Random()
}

func loadDB(filename string) error {
	f.db = filename
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error loading database at %q: %w", filename, err)
	}
	factsfromdisk := make(map[string][]factoid)
	//	if err := yaml.Unmarshal(content, &factsfromdisk); err != nil {
	if err := json.Unmarshal(content, &factsfromdisk); err != nil {
		return fmt.Errorf("error parsing database at %q: %w", filename, err)
	}
	f.m.Lock()
	defer f.m.Unlock()
	f.v = make(map[string]FactoidSet)
	for k, vs := range factsfromdisk {
		f.v[k] = NewFactoidSet(vs...)
	}
	return nil
}

// set adds a value to a factoid key
func set(key string, value factoid) error {
	if !utf8.Valid([]byte(value.Value)) {
		return ErrInvalidUTF8
	}
	f.m.Lock()
	defer f.m.Unlock()
	if _, ok := f.v[key]; !ok {
		f.v[key] = NewFactoidSet()
	}
	if f.v[key].Exists(value) {
		return ErrFactAlreadyExists
	}
	f.v[key].Add(value)

	return syncToDisk()
}

// get retrieves a random fact from the factoid DB
func get(key string) (factoid, error) {
	if facts, ok := f.v[key]; ok {
		return facts.Random(), nil
	}
	return factoid{}, ErrNoSuchFact
}

// getall returns all factoids for a key
func (f *factoids) getall(key string) ([]string, error) {
	f.m.Lock()
	vals, ok := f.v[key]
	f.m.Unlock()
	if !ok || len(vals) == 0 {
		return nil, ErrNoSuchFact
	}
	return vals.Values(), nil
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

// Search returns a slice of maximum of `max` factoids and an integer with the number of additional facts found
func (f *factoids) search(rex *regexp.Regexp, max int) ([]fullfactoid, int) {
	rv := make([]fullfactoid, 0, max)
	additional := 0
	for k, v := range f.v {
		for _, fact := range v.Slice() {
			if rex.MatchString(fact.Value) {
				if len(rv) >= max {
					additional++
					continue
				}
				rv = append(rv, fullfactoid{Keyword: k, factoid: fact})
			}
		}
	}
	return rv, additional
}

func (f *factoids) delete(key string) {
	f.m.Lock()
	delete(f.v, key)
	f.m.Unlock()
}

// listFacts lists all keys mathcing `substring`
func (f *factoids) listFacts(rex *regexp.Regexp) []string {
	rv := make([]string, 0, 10)
	for k := range f.v {
		if rex.MatchString(k) {
			rv = append(rv, strings.TrimSpace(k))
		}
	}
	sort.Strings(rv)
	return rv
}

// sync syncs the in memory DB to disk. THe caller should lock!
func syncToDisk() error {
	factsfordisk := make(map[string][]factoid)
	for k, vs := range f.v {
		factsfordisk[k] = vs.Slice()
	}
	jsondata, err := json.Marshal(factsfordisk)
	//ymldata, err := yaml.Marshal(factsfordisk)
	if err != nil {
		return fmt.Errorf("error marshalling DB: %w", err)
	}
	if err := os.WriteFile(f.db, jsondata, 0644); err != nil {
		return fmt.Errorf("error syncing to file %q: %w", f.db, err)
	}
	return nil
}
